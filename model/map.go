package model

import (
	"encoding/json"
	"time"
)

const (
	TileW  = 64 // 像素/Tile
	TileH  = 64
	ChunkW = 16 // Tile/Chunk
	ChunkH = 16
)

type BBox struct {
	MinX, MinY float64 // 像素坐标
	MaxX, MaxY float64
}

type ChunkRow struct {
	MapID   string
	Layer   string
	CX      int
	CY      int
	Rev     int64
	Payload []byte // gzip 二进制
	Updated time.Time
}

type ChunkPayload struct {
	MapID      string `json:"mapId"`
	Layer      string `json:"layer"`
	CX         int    `json:"cx"`
	CY         int    `json:"cy"`
	TileSize   [2]int `json:"tileSize"`
	ChunkTiles [2]int `json:"chunkTiles"`
	Data       []int  `json:"data"`
	Rev        int64  `json:"rev,omitempty"` // ← 新增，用于增量
}

type EntityRow struct {
	EntityID int64
	MapID    string
	Type     string
	X, Y     int
	W, H     int
	CompJSON json.RawMessage
	ChunkX   int
	ChunkY   int
	Updated  time.Time
}
