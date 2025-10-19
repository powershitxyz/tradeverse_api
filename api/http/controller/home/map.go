package home

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"chaos/api/api/common"
	"chaos/api/codes"
	"chaos/api/model"
	"chaos/api/service"
)

// ---------- 请求与响应结构 ----------

type ClientRev struct {
	Layer string `json:"layer" binding:"required"`
	CX    int    `json:"cx"    binding:"required"`
	CY    int    `json:"cy"    binding:"required"`
	Rev   int64  `json:"rev"   binding:"required"`
}

type ViewportReq struct {
	Layers       []string    `json:"layers"`                   // 例如 ["land1","walk_level1","deco"]；为空则有默认
	BBox         model.BBox  `json:"bbox"  binding:"required"` // 视口像素坐标
	MarginTiles  int         `json:"marginTiles"`              // 可选：以 tile 为单位的边缘预取
	MarginPixels int         `json:"marginPixels"`             // 可选：以像素为单位的边缘预取（与 MarginTiles 二选一）
	ClientRevs   []ClientRev `json:"clientRevs"`               // 前端缓存的块版本（用于增量）
}

type ViewportResp struct {
	Chunks    []*model.ChunkPayload `json:"chunks"`    // 仅返回有变化（或前端未持有）的块
	Unchanged []ClientRev           `json:"unchanged"` // 前端已持有且版本匹配的块
	Entities  []model.EntityRow     `json:"entities"`
}

// ---------- 工具：边缘预取 ----------

func expandBBox(b model.BBox, marginTiles, marginPixels int) model.BBox {
	const (
		tw = model.TileW
		th = model.TileH
	)
	if marginTiles > 0 {
		pxX := float64(marginTiles * tw)
		pxY := float64(marginTiles * th)
		b.MinX -= pxX
		b.MinY -= pxY
		b.MaxX += pxX
		b.MaxY += pxY
	} else if marginPixels > 0 {
		px := float64(marginPixels)
		b.MinX -= px
		b.MinY -= px
		b.MaxX += px
		b.MaxY += px
	}
	return b
}

// ---------- Gin Handler ----------

// POST /maps/:mapId/viewport
// Body: ViewportReq (JSON)
// 返回：ViewportResp
func GetMapViewport(c *gin.Context) {
	res := common.Response{Timestamp: time.Now().Unix(), Code: codes.CODE_SUCCESS, Msg: "success"}

	mapID := c.Param("mapId")
	if mapID == "" {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "mapId is required"
		c.JSON(http.StatusOK, res)
		return
	}

	var req ViewportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "invalid json body: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	// 默认图层（按你的工程需要可调整）
	if len(req.Layers) == 0 {
		req.Layers = []string{"land1", "walk_level1", "deco", "land2", "walk_level2", "connect"}
	}

	// 边缘预取
	expanded := expandBBox(req.BBox, req.MarginTiles, req.MarginPixels)

	// 查询
	svc := service.NewMapService()
	vp, err := svc.LoadViewport(c.Request.Context(), mapID, req.Layers, expanded)
	if err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	// 增量过滤：把客户端已持有且 rev 相同的块剔除，只返回变化块
	known := make(map[string]int64, len(req.ClientRevs))
	for _, cr := range req.ClientRevs {
		k := cr.Layer + ":" + itoa(cr.CX) + ":" + itoa(cr.CY)
		known[k] = cr.Rev
	}

	var changed []*model.ChunkPayload
	var unchanged []ClientRev

	for _, ch := range vp.Chunks {
		k := ch.Layer + ":" + itoa(ch.CX) + ":" + itoa(ch.CY)
		if r, ok := known[k]; ok && r == int64(findRev(vp, ch.Layer, ch.CX, ch.CY)) {
			// 客户端版本与服务端一致：不返回 payload
			unchanged = append(unchanged, ClientRev{Layer: ch.Layer, CX: ch.CX, CY: ch.CY, Rev: r})
			continue
		}
		changed = append(changed, ch)
	}

	resp := ViewportResp{
		Chunks:    changed,
		Unchanged: unchanged,
		Entities:  vp.Entities,
	}
	res.Data = resp
	c.JSON(http.StatusOK, res)
}

// findRev 在 service.LoadViewport 的结果中找到该块的 rev（为了增量比较）
// 说明：你的 ChunkPayload 里没有 rev 字段，但数据库有 rev。为了简洁，这里从同批查询的 raw rows 取 rev，
// 已经在 service 层封装：把 rev 附加到 ChunkPayload 的一个“扩展字段”更好，见下方 service 小改动。
func findRev(vp *service.ViewportResult, layer string, cx, cy int) int64 {
	// 如果你按下面“服务层小改”把 Rev 附加进 ChunkPayload 的扩展字段，这里可以直接读
	// 这里先兜底返回 0（表示客户端永远拿不到“未变化”判断，会全部返回）
	// 建议按下一节把 Rev 透出。
	return 0
}

// 简单 itoa（避免 strconv 带来过多噪音）
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + (n % 10))
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
