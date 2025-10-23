package mappkg

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"chaos/api/model"
	"chaos/api/system"
)

/* ---------- Tiled 结构（只放 tmj 需要的字段，命名与 tmj 一致） ---------- */

type TiledMap struct {
	CompressionLevel int               `json:"compressionlevel"`
	Infinite         bool              `json:"infinite"`
	Orientation      string            `json:"orientation"`
	RenderOrder      string            `json:"renderorder"`
	TiledVersion     string            `json:"tiledversion"`
	Type             string            `json:"type"`    // "map"
	Version          string            `json:"version"` // "1.10"
	Width            int               `json:"width"`
	Height           int               `json:"height"`
	TileWidth        int               `json:"tilewidth"`
	TileHeight       int               `json:"tileheight"`
	NextLayerID      int               `json:"nextlayerid"`
	NextObjectID     int               `json:"nextobjectid"`
	Layers           []any             `json:"layers"`   // []TiledLayer*（混合）
	Tilesets         []json.RawMessage `json:"tilesets"` // 原样透传
}

type TiledTileLayer struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Type    string  `json:"type"` // "tilelayer"
	Visible bool    `json:"visible"`
	Opacity float64 `json:"opacity"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
	Data    []int   `json:"data"`
	X       int     `json:"x"`
	Y       int     `json:"y"`
}

type TiledImageLayer struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"` // "imagelayer"
	Visible     bool    `json:"visible"`
	Opacity     float64 `json:"opacity"`
	Image       string  `json:"image"`
	ImageWidth  int     `json:"imagewidth,omitempty"`
	ImageHeight int     `json:"imageheight,omitempty"`
	RepeatX     bool    `json:"repeatx"`
	RepeatY     bool    `json:"repeaty"`
	X           int     `json:"x"`
	Y           int     `json:"y"`
}

type TiledObject struct {
	ID         int64           `json:"id"`
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Gid        int             `json:"gid,omitempty"`
	X          float64         `json:"x"`
	Y          float64         `json:"y"`
	Width      float64         `json:"width"`
	Height     float64         `json:"height"`
	Rotation   float64         `json:"rotation"`
	Visible    bool            `json:"visible"`
	Properties json.RawMessage `json:"properties,omitempty"`
}

type TiledObjectLayer struct {
	ID        int           `json:"id"`
	Name      string        `json:"name"`
	Type      string        `json:"type"` // "objectgroup"
	Visible   bool          `json:"visible"`
	Opacity   float64       `json:"opacity"`
	DrawOrder string        `json:"draworder"` // "topdown"
	Objects   []TiledObject `json:"objects"`
	X         int           `json:"x"`
	Y         int           `json:"y"`
}

/* ---------- DB rows ---------- */

type metaRow struct {
	HeaderJSON   []byte
	TilesetsJSON []byte
}

type imgRow struct {
	Name    string
	Z       int
	Image   string
	Opacity float64
	RepeatX bool
	RepeatY bool
	X       int
	Y       int
	Visible bool
}

type layerMetaRow struct {
	Layer  string
	ZIndex int
	Kind   string
}

type chunkRow = model.ChunkRow
type entityRow = model.EntityRow

/* ---------- Service ---------- */

type MapAssembler struct{}

func NewMapAssembler() *MapAssembler { return &MapAssembler{} }

// BuildFullTMJ 组装完整 tmj；
// - layersFilter 为空表示全部图层；
// - bbox 为 nil 表示全图，否则按视口拼接（仍返回 tmj 结构，但 data 为裁剪后尺寸）。
func (s *MapAssembler) BuildFullTMJ(ctx context.Context, mapID string, layersFilter []string, bbox *model.BBox) (*TiledMap, error) {
	// 1) meta
	mh, err := s.loadMeta(ctx, mapID)
	if err != nil {
		return nil, fmt.Errorf("loadMeta: %w", err)
	}
	var tm TiledMap
	if err := json.Unmarshal(mh.HeaderJSON, &tm); err != nil {
		return nil, fmt.Errorf("unmarshal header_json: %w", err)
	}
	if err := json.Unmarshal(mh.TilesetsJSON, &tm.Tilesets); err != nil {
		return nil, fmt.Errorf("unmarshal tilesets_json: %w", err)
	}

	// 若做“视口裁剪”，替换 width/height（tile 数），保持 tile 尺寸不变
	var crop *cropInfo
	if bbox != nil {
		c := computeCrop(&tm, *bbox)
		tm.Width, tm.Height = c.Width, c.Height
		crop = &c
	}

	// 2) imagelayers
	imgs, err := s.loadImageLayers(ctx, mapID)
	if err != nil {
		return nil, fmt.Errorf("loadImageLayers: %w", err)
	}

	// 3) layer 元数据（决定层顺序）
	lMetas, err := s.loadLayerMetas(ctx, mapID, layersFilter)
	if err != nil {
		return nil, fmt.Errorf("loadLayerMetas: %w", err)
	}

	// 4) chunks（只拿到需要的范围）
	tileLayerNames := pickTileLayerNames(lMetas, imgs)
	chunks, err := s.loadChunks(ctx, mapID, tileLayerNames, &tm, crop)
	if err != nil {
		return nil, fmt.Errorf("loadChunks: %w", err)
	}

	// 5) entities → objectgroup
	entities, err := s.loadEntities(ctx, mapID, crop)
	if err != nil {
		return nil, fmt.Errorf("loadEntities: %w", err)
	}

	// 6) 逐层组装
	layersAny := make([]any, 0, len(lMetas)+len(imgs))
	// 6.1 image layers
	for _, im := range imgs {
		if !nameAllowed(im.Name, layersFilter) {
			continue
		}
		layersAny = append(layersAny, TiledImageLayer{
			ID: im.Z, Name: im.Name, Type: "imagelayer", Visible: im.Visible, Opacity: im.Opacity,
			Image: im.Image, RepeatX: im.RepeatX, RepeatY: im.RepeatY, X: im.X, Y: im.Y,
		})
	}
	// 6.2 tile layers
	for _, lm := range lMetas {
		if !nameAllowed(lm.Layer, layersFilter) {
			continue
		}
		data := assembleTileLayer(&tm, crop, chunks[lm.Layer])
		layersAny = append(layersAny, TiledTileLayer{
			ID: lm.ZIndex, Name: lm.Layer, Type: "tilelayer", Visible: true, Opacity: 1,
			Width: tm.Width, Height: tm.Height, Data: data, X: 0, Y: 0,
		})
	}
	// 6.3 object layer（一个或多个；这里合并为一个“业务层”）
	if len(entities) > 0 && nameAllowed("gamecenter", layersFilter) {
		objs := make([]TiledObject, 0, len(entities))
		for _, e := range entities {
			objs = append(objs, TiledObject{
				ID: e.EntityID, Name: "", Type: e.Type,
				X: float64(e.X), Y: float64(e.Y),
				Width: float64(e.W), Height: float64(e.H),
				Visible: true, Rotation: 0,
				Properties: e.CompJSON, // 保持 properties
			})
		}
		layersAny = append(layersAny, TiledObjectLayer{
			ID: 999999, Name: "gamecenter", Type: "objectgroup", Visible: true, Opacity: 1,
			DrawOrder: "topdown", Objects: objs, X: 0, Y: 0,
		})
	}

	tm.Layers = layersAny
	return &tm, nil
}

/* ---------- data access ---------- */

func (s *MapAssembler) loadMeta(ctx context.Context, mapID string) (*metaRow, error) {
	const q = `SELECT header_json, tilesets_json FROM map_meta WHERE map_id = ?`
	rows, err := system.QueryContext(ctx, q, mapID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	var r metaRow
	if err := rows.Scan(&r.HeaderJSON, &r.TilesetsJSON); err != nil {
		return nil, err
	}
	return &r, rows.Err()
}

func (s *MapAssembler) loadImageLayers(ctx context.Context, mapID string) ([]imgRow, error) {
	const q = `
SELECT name, z_index, image, opacity, repeatx, repeaty, x, y, visible
FROM map_imagelayer WHERE map_id = ? ORDER BY z_index ASC`
	rows, err := system.QueryContext(ctx, q, mapID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []imgRow
	for rows.Next() {
		var r imgRow
		if err := rows.Scan(&r.Name, &r.Z, &r.Image, &r.Opacity, &r.RepeatX, &r.RepeatY, &r.X, &r.Y, &r.Visible); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *MapAssembler) loadLayerMetas(ctx context.Context, mapID string, filter []string) ([]layerMetaRow, error) {
	q := `SELECT layer, z_index, kind FROM map_layer_meta WHERE map_id = ?`
	args := []any{mapID}
	if len(filter) > 0 {
		q += " AND layer IN (" + strings.Repeat("?,", len(filter)-1) + "?)"
		for _, v := range filter {
			args = append(args, v)
		}
	}
	q += " ORDER BY z_index ASC"

	rows, err := system.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []layerMetaRow
	for rows.Next() {
		var r layerMetaRow
		if err := rows.Scan(&r.Layer, &r.ZIndex, &r.Kind); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *MapAssembler) loadChunks(ctx context.Context, mapID string, layers []string, tm *TiledMap, crop *cropInfo) (map[string][]chunkRow, error) {
	if len(layers) == 0 {
		return map[string][]chunkRow{}, nil
	}

	// 计算需要的 chunk 范围（全图或视口）
	minCX, maxCX, minCY, maxCY := 0, math.MaxInt32, 0, math.MaxInt32
	if crop != nil {
		minCX, maxCX, minCY, maxCY = crop.ChunkMinX, crop.ChunkMaxX, crop.ChunkMinY, crop.ChunkMaxY
	} else {
		// 全图范围（根据 width/height 推断）
		maxCX = int(math.Ceil(float64(tm.Width)/float64(model.ChunkW))) - 1
		maxCY = int(math.Ceil(float64(tm.Height)/float64(model.ChunkH))) - 1
	}

	inClause := "(" + strings.Repeat("?,", len(layers)-1) + "?)"
	q := fmt.Sprintf(`
SELECT map_id, layer, cx, cy, rev, payload, updated_at
FROM map_chunk_layer
WHERE map_id = ?
  AND layer IN %s
  AND cx BETWEEN ? AND ?
  AND cy BETWEEN ? AND ?
ORDER BY layer, cy, cx`, inClause)

	args := make([]any, 0, 1+len(layers)+4)
	args = append(args, mapID)
	for _, l := range layers {
		args = append(args, l)
	}
	args = append(args, minCX, maxCX, minCY, maxCY)

	ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	rows, err := system.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string][]chunkRow)
	for rows.Next() {
		var r chunkRow
		if err := rows.Scan(&r.MapID, &r.Layer, &r.CX, &r.CY, &r.Rev, &r.Payload, &r.Updated); err != nil {
			return nil, err
		}
		out[r.Layer] = append(out[r.Layer], r)
	}
	return out, rows.Err()
}

func (s *MapAssembler) loadEntities(ctx context.Context, mapID string, crop *cropInfo) ([]entityRow, error) {
	q := `
SELECT entity_id, map_id, type, x, y, w, h, comp_json, chunk_x, chunk_y, updated_at
FROM map_entity WHERE map_id = ?`
	args := []any{mapID}

	if crop != nil {
		q += `
  AND (x + w) > ? AND x < ? AND (y + h) > ? AND y < ?`
		args = append(args, crop.PixelMinX, crop.PixelMaxX, crop.PixelMinY, crop.PixelMaxY)
	}
	q += ` ORDER BY updated_at DESC, entity_id ASC`

	rows, err := system.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []entityRow
	for rows.Next() {
		var e entityRow
		if err := rows.Scan(&e.EntityID, &e.MapID, &e.Type, &e.X, &e.Y, &e.W, &e.H, &e.CompJSON, &e.ChunkX, &e.ChunkY, &e.Updated); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

/* ---------- assemble helpers ---------- */

type cropInfo struct {
	// 视口对应的像素范围 & chunk/tile 范围
	PixelMinX, PixelMaxX int
	PixelMinY, PixelMaxY int
	ChunkMinX, ChunkMaxX int
	ChunkMinY, ChunkMaxY int
	TileMinX, TileMaxX   int
	TileMinY, TileMaxY   int
	Width, Height        int // 裁剪后 tmj.width/height（单位：tile）
}

func computeCrop(tm *TiledMap, bbox model.BBox) cropInfo {
	if bbox.MinX > bbox.MaxX {
		bbox.MinX, bbox.MaxX = bbox.MaxX, bbox.MinX
	}
	if bbox.MinY > bbox.MaxY {
		bbox.MinY, bbox.MaxY = bbox.MaxY, bbox.MinY
	}

	minTX := int(math.Floor(bbox.MinX / float64(tm.TileWidth)))
	maxTX := int(math.Floor((bbox.MaxX - 1) / float64(tm.TileWidth)))
	minTY := int(math.Floor(bbox.MinY / float64(tm.TileHeight)))
	maxTY := int(math.Floor((bbox.MaxY - 1) / float64(tm.TileHeight)))

	minCX := int(math.Floor(float64(minTX) / float64(model.ChunkW)))
	maxCX := int(math.Floor(float64(maxTX) / float64(model.ChunkW)))
	minCY := int(math.Floor(float64(minTY) / float64(model.ChunkH)))
	maxCY := int(math.Floor(float64(maxTY) / float64(model.ChunkH)))

	return cropInfo{
		PixelMinX: int(bbox.MinX), PixelMaxX: int(bbox.MaxX),
		PixelMinY: int(bbox.MinY), PixelMaxY: int(bbox.MaxY),
		ChunkMinX: minCX, ChunkMaxX: maxCX,
		ChunkMinY: minCY, ChunkMaxY: maxCY,
		TileMinX: minTX, TileMaxX: maxTX,
		TileMinY: minTY, TileMaxY: maxTY,
		Width:  maxTX - minTX + 1,
		Height: maxTY - minTY + 1,
	}
}

func pickTileLayerNames(lms []layerMetaRow, imgs []imgRow) []string {
	imgSet := map[string]bool{}
	for _, i := range imgs {
		imgSet[i.Name] = true
	}
	out := make([]string, 0, len(lms))
	for _, lm := range lms {
		if !imgSet[lm.Layer] { // 不是 imagelayer，就当作 tilelayer
			out = append(out, lm.Layer)
		}
	}
	return out
}

func nameAllowed(name string, filter []string) bool {
	if len(filter) == 0 {
		return true
	}
	for _, f := range filter {
		if f == name {
			return true
		}
	}
	return false
}

func assembleTileLayer(tm *TiledMap, crop *cropInfo, rows []chunkRow) []int {
	// 视口尺寸或全图尺寸
	width := tm.Width
	height := tm.Height
	offsetTileX, offsetTileY := 0, 0
	if crop != nil {
		width, height = crop.Width, crop.Height
		offsetTileX, offsetTileY = crop.TileMinX, crop.TileMinY
	}

	data := make([]int, width*height) // 默认为 0
	for _, r := range rows {
		cp, err := decodeChunkPayload(r)
		if err != nil {
			continue
		}
		chW, chH := cp.ChunkTiles[0], cp.ChunkTiles[1]
		baseX := r.CX*chW - offsetTileX
		baseY := r.CY*chH - offsetTileY
		if baseX >= width || baseY >= height || baseX+chW <= -0 || baseY+chH <= -0 {
			continue
		}
		for yy := 0; yy < chH; yy++ {
			globalY := baseY + yy
			if globalY < 0 || globalY >= height {
				continue
			}
			for xx := 0; xx < chW; xx++ {
				globalX := baseX + xx
				if globalX < 0 || globalX >= width {
					continue
				}
				val := cp.Data[yy*chW+xx]
				if val == 0 {
					continue
				}
				data[globalY*width+globalX] = val
			}
		}
	}
	return data
}

func decodeChunkPayload(row chunkRow) (*model.ChunkPayload, error) {
	gr, err := gzip.NewReader(bytes.NewReader(row.Payload))
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	raw, err := io.ReadAll(gr)
	if err != nil {
		return nil, err
	}
	var cp model.ChunkPayload
	if err := json.Unmarshal(raw, &cp); err != nil {
		return nil, err
	}
	return &cp, nil
}
