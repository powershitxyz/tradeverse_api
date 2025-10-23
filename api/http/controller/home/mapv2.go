package home

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"chaos/api/api/common"
	"chaos/api/codes"
	"chaos/api/model"
	"chaos/api/service/mappkg"
)

type mapRequest struct {
	Layers []string    `json:"layers,omitempty"` // 为空=全层
	BBox   *model.BBox `json:"bbox,omitempty"`   // 为空=全图
}

// POST /spwapi/maps/:map_id/tmj
func MapTMJ(c *gin.Context) {
	res := common.Response{Timestamp: time.Now().Unix()}
	mapID := c.Param("mapId")
	if mapID == "" {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "mapId is required"
		c.JSON(http.StatusOK, res)
		return
	}

	var req mapRequest
	_ = c.ShouldBindJSON(&req) // 可为空

	assembler := mappkg.NewMapAssembler()
	tm, err := assembler.BuildFullTMJ(c, mapID, req.Layers, req.BBox)
	if err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = tm //完整的 .tmj JSON 结构
	c.JSON(http.StatusOK, res)
}
