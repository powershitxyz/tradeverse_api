package main

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type TiledMap struct {
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	TileWidth  int     `json:"tilewidth"`
	TileHeight int     `json:"tileheight"`
	Layers     []Layer `json:"layers"`
}

type Layer struct {
	ID         int           `json:"id"`
	Name       string        `json:"name"`
	Type       string        `json:"type"` // tilelayer | objectgroup | imagelayer ...
	Visible    bool          `json:"visible"`
	Opacity    float64       `json:"opacity"`
	Width      int           `json:"width"`      // for tilelayer
	Height     int           `json:"height"`     // for tilelayer
	Data       []int         `json:"data"`       // for tilelayer (array form)
	Objects    []TiledObject `json:"objects"`    // for objectgroup
	Properties []TiledProp   `json:"properties"` // optional
}

type TiledObject struct {
	ID         int         `json:"id"`
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Gid        int         `json:"gid"`
	X          float64     `json:"x"`
	Y          float64     `json:"y"`
	Width      float64     `json:"width"`
	Height     float64     `json:"height"`
	Rotation   float64     `json:"rotation"`
	Visible    bool        `json:"visible"`
	Properties []TiledProp `json:"properties"`
}

type TiledProp struct {
	Name  string          `json:"name"`
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

func main() {
	input := flag.String("input", "map.tmj", "Path to Tiled JSON (.tmj/.json)")
	mapID := flag.String("map-id", "", "Map ID (default: derive from file name)")
	chunk := flag.String("chunk", "16x16", "Chunk size in tiles, e.g. 16x16 / 32x32")
	rev := flag.Int("rev", 1, "Chunk revision")
	outDir := flag.String("out", "out_sql", "Output directory")
	flag.Parse()

	chW, chH, err := parseChunk(*chunk)
	must(err)

	data, err := os.ReadFile(*input)
	must(err)

	var tm TiledMap
	must(json.Unmarshal(data, &tm))

	if *mapID == "" {
		*mapID = sanitizeMapID(strings.TrimSuffix(filepath.Base(*input), filepath.Ext(*input)))
	}

	gridW := int(math.Ceil(float64(tm.Width) / float64(chW)))
	gridH := int(math.Ceil(float64(tm.Height) / float64(chH)))

	_ = os.MkdirAll(*outDir, 0o755)
	chunksSQLPath := filepath.Join(*outDir, "map_chunk_layer_inserts.sql")
	entitiesSQLPath := filepath.Join(*outDir, "map_entity_inserts.sql")

	chunksBuf := &strings.Builder{}
	entitiesBuf := &strings.Builder{}

	fmt.Fprintf(chunksBuf, "-- INSERTs for map_chunk_layer (map_id=%s, chunk=%dx%d, grid=%dx%d, tile=%dx%d px)\n",
		*mapID, chW, chH, gridW, gridH, tm.TileWidth, tm.TileHeight)
	fmt.Fprintf(entitiesBuf, "-- INSERTs for map_entity (map_id=%s)\n", *mapID)

	// 1) tile layers -> per-layer per-chunk
	for _, layer := range tm.Layers {
		if layer.Type != "tilelayer" {
			continue
		}
		lname := layer.Name
		if lname == "" {
			lname = fmt.Sprintf("layer_%d", layer.ID)
		}
		// sanity
		if layer.Width != tm.Width || layer.Height != tm.Height {
			// 固定尺寸图（非 infinite）应相等
			fmt.Fprintf(os.Stderr, "warn: tilelayer %q size (%dx%d) != map size (%dx%d)\n",
				lname, layer.Width, layer.Height, tm.Width, tm.Height)
		}
		// 切块
		for cy := 0; cy < gridH; cy++ {
			for cx := 0; cx < gridW; cx++ {
				sx := cx * chW
				sy := cy * chH
				sub := extractTileChunk(layer.Data, tm.Width, tm.Height, sx, sy, chW, chH)

				payload := map[string]any{
					"mapId":      *mapID,
					"layer":      lname,
					"cx":         cx,
					"cy":         cy,
					"tileSize":   []int{tm.TileWidth, tm.TileHeight},
					"chunkTiles": []int{chW, chH},
					"data":       sub,
				}
				j, _ := json.Marshal(payload)
				hexGz := gzipToHex(j)

				// 你的表结构只有 payload 列，所以直接 UNHEX 写入
				fmt.Fprintf(chunksBuf,
					"REPLACE INTO map_chunk_layer (map_id, layer, cx, cy, rev, payload) VALUES ('%s','%s',%d,%d,%d,UNHEX('%s'));\n",
					escapeSQL(*mapID), escapeSQL(lname), cx, cy, *rev, hexGz)
			}
		}
	}

	// 2) object layers -> entities (主归属块按锚点决定)
	for _, layer := range tm.Layers {
		if layer.Type != "objectgroup" {
			continue
		}
		for _, obj := range layer.Objects {
			// 聚合 properties
			props := map[string]any{}
			for _, p := range obj.Properties {
				var v any
				_ = json.Unmarshal(p.Value, &v)
				props[p.Name] = v
			}
			compJSON, _ := json.Marshal(props)

			// 主块（按对象锚点像素坐标 -> tile -> chunk）
			tx := int(obj.X) / tm.TileWidth
			ty := int(obj.Y) / tm.TileHeight
			pcx := tx / chW
			pcy := ty / chH

			otype := obj.Type
			if otype == "" {
				otype = obj.Name
			}
			if otype == "" {
				otype = "object"
			}
			fmt.Fprintf(entitiesBuf,
				"INSERT INTO map_entity (map_id, type, x, y, w, h, comp_json, chunk_x, chunk_y) VALUES "+
					"('%s','%s',%d,%d,%d,%d,'%s',%d,%d);\n",
				escapeSQL(*mapID),
				escapeSQL(otype),
				int(obj.X), int(obj.Y), int(obj.Width), int(obj.Height),
				escapeSQL(string(compJSON)),
				pcx, pcy)
		}
	}

	must(os.WriteFile(chunksSQLPath, []byte(chunksBuf.String()), 0o644))
	must(os.WriteFile(entitiesSQLPath, []byte(entitiesBuf.String()), 0o644))

	fmt.Printf("OK\nmap_id: %s\nmap tiles: %dx%d (tile %dx%d px)\nchunk: %dx%d -> grid %dx%d\n",
		*mapID, tm.Width, tm.Height, tm.TileWidth, tm.TileHeight, chW, chH, gridW, gridH)
	fmt.Println("files:")
	fmt.Println("  ", chunksSQLPath)
	fmt.Println("  ", entitiesSQLPath)
}

// ---- helpers ----

func parseChunk(s string) (int, int, error) {
	re := regexp.MustCompile(`^\s*(\d+)\s*[xX]\s*(\d+)\s*$`)
	m := re.FindStringSubmatch(s)
	if len(m) != 3 {
		return 0, 0, fmt.Errorf("bad --chunk value: %q", s)
	}
	return atoi(m[1]), atoi(m[2]), nil
}

func atoi(s string) int {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

func sanitizeMapID(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_\-]`)
	id := re.ReplaceAllString(name, "_")
	if len(id) > 64 {
		id = id[:64]
	}
	if id == "" {
		id = "map"
	}
	return id
}

func extractTileChunk(data []int, mapW, mapH, sx, sy, cw, ch int) []int {
	out := make([]int, 0, cw*ch)
	for y := sy; y < min(sy+ch, mapH); y++ {
		start := y*mapW + sx
		end := min(start+(cw), y*mapW+mapW)
		row := append([]int(nil), data[start:end]...)
		// 右边不够补 0
		if len(row) < cw {
			row = append(row, make([]int, cw-len(row))...)
		}
		out = append(out, row...)
	}
	// 底部不够补 0
	rows := min(ch, mapH-sy)
	if rows < ch {
		out = append(out, make([]int, cw*(ch-rows))...)
	}
	return out
}

func gzipToHex(b []byte) string {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write(b)
	_ = gz.Close()
	return strings.ToUpper(hex.EncodeToString(buf.Bytes()))
}

func escapeSQL(s string) string {
	// 最简转义：单引号 -> 两个单引号；反斜杠 -> 双反斜杠
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `''`)
	return s
}

func must(err error) {
	if err != nil {
		if err == io.EOF {
			return
		}
		panic(err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
