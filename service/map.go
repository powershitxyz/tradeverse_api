package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"chaos/api/model"
	"chaos/api/system"
)

type MapService struct{}

func NewMapService() *MapService { return &MapService{} }

// ------------------------------------------------------------
// 1) 地图区域加载主入口
// ------------------------------------------------------------

type ViewportResult struct {
	Chunks   []*model.ChunkPayload `json:"chunks"`
	Entities []model.EntityRow     `json:"entities"`
}

func (s *MapService) LoadViewport(ctx context.Context, mapID string, layers []string, bbox model.BBox) (*ViewportResult, error) {
	chunks, err := s.queryNearbyChunks(ctx, mapID, layers, bbox)
	if err != nil {
		return nil, fmt.Errorf("queryNearbyChunks: %w", err)
	}
	outChunks := make([]*model.ChunkPayload, 0, len(chunks))
	for _, row := range chunks {
		cp, err := decodeChunkPayload(row)
		if err != nil {
			return nil, fmt.Errorf("decode chunk layer=%s cx=%d cy=%d: %w", row.Layer, row.CX, row.CY, err)
		}
		cp.Rev = row.Rev
		outChunks = append(outChunks, cp)
	}

	ents, err := s.queryEntitiesInView(ctx, mapID, bbox)
	if err != nil {
		return nil, fmt.Errorf("queryEntitiesInView: %w", err)
	}

	return &ViewportResult{Chunks: outChunks, Entities: ents}, nil
}

// ------------------------------------------------------------
// 2) Chunk 查询
// ------------------------------------------------------------

func (s *MapService) queryNearbyChunks(ctx context.Context, mapID string, layers []string, bbox model.BBox) ([]model.ChunkRow, error) {
	minCX, maxCX, minCY, maxCY := bboxToChunkRange(bbox)

	// 构建 layer IN 子句和 FIELD 排序
	inClause := buildInPlaceholders(len(layers)) // 例如 "(?,?,?)"
	fieldList := strings.Trim(inClause, "()")    // 例如 "?,?,?" —— 无括号

	sqlStr := fmt.Sprintf(`
SELECT map_id, layer, cx, cy, rev, payload, updated_at
FROM map_chunk_layer
WHERE map_id = ?
  AND layer IN %s
  AND cx BETWEEN ? AND ?
  AND cy BETWEEN ? AND ?
ORDER BY FIELD(layer, %s), cy, cx;
`, inClause, fieldList)

	// 参数顺序：map_id, layers..., cxMin, cxMax, cyMin, cyMax, layers...
	args := make([]any, 0, 1+len(layers)*2+4)
	args = append(args, mapID)
	for _, l := range layers {
		args = append(args, l)
	}
	args = append(args, minCX, maxCX, minCY, maxCY)
	for _, l := range layers {
		args = append(args, l)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := system.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("queryNearbyChunks sql: %w", err)
	}
	defer rows.Close()

	var result []model.ChunkRow
	for rows.Next() {
		var r model.ChunkRow
		if err := rows.Scan(&r.MapID, &r.Layer, &r.CX, &r.CY, &r.Rev, &r.Payload, &r.Updated); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

// ------------------------------------------------------------
// 3) 实体查询
// ------------------------------------------------------------

func (s *MapService) queryEntitiesInView(ctx context.Context, mapID string, bbox model.BBox) ([]model.EntityRow, error) {

	const q = `
SELECT entity_id, map_id, type, x, y, w, h, comp_json, chunk_x, chunk_y, updated_at
FROM map_entity
WHERE map_id = ?
  AND (x + w) > ?  -- 右 > 视口左
  AND x < ?        -- 左 < 视口右
  AND (y + h) > ?  -- 下 > 视口上
  AND y < ?        -- 上 < 视口下
ORDER BY updated_at DESC, entity_id ASC;
`
	args := []any{mapID, int(bbox.MinX), int(bbox.MaxX), int(bbox.MinY), int(bbox.MaxY)}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := system.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.EntityRow
	for rows.Next() {
		var e model.EntityRow
		if err := rows.Scan(&e.EntityID, &e.MapID, &e.Type, &e.X, &e.Y, &e.W, &e.H,
			&e.CompJSON, &e.ChunkX, &e.ChunkY, &e.Updated); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ------------------------------------------------------------
// 4) 辅助方法
// ------------------------------------------------------------

func bboxToChunkRange(b model.BBox) (minCX, maxCX, minCY, maxCY int) {
	if b.MinX > b.MaxX {
		b.MinX, b.MaxX = b.MaxX, b.MinX
	}
	if b.MinY > b.MaxY {
		b.MinY, b.MaxY = b.MaxY, b.MinY
	}
	minTX := int(math.Floor(b.MinX / model.TileW))
	maxTX := int(math.Floor((b.MaxX - 1) / model.TileW))
	minTY := int(math.Floor(b.MinY / model.TileH))
	maxTY := int(math.Floor((b.MaxY - 1) / model.TileH))
	minCX = int(math.Floor(float64(minTX) / model.ChunkW))
	maxCX = int(math.Floor(float64(maxTX) / model.ChunkW))
	minCY = int(math.Floor(float64(minTY) / model.ChunkH))
	maxCY = int(math.Floor(float64(maxTY) / model.ChunkH))
	return
}

func buildInPlaceholders(n int) string {
	if n <= 0 {
		return "(NULL)"
	}
	var b strings.Builder
	b.WriteString("(")
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString("?")
	}
	b.WriteString(")")
	return b.String()
}

func decodeChunkPayload(row model.ChunkRow) (*model.ChunkPayload, error) {
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
	if cp.Layer != row.Layer || cp.CX != row.CX || cp.CY != row.CY {
		return &cp, fmt.Errorf("payload mismatch: layer=%s cx=%d cy=%d", row.Layer, row.CX, row.CY)
	}
	return &cp, nil
}
