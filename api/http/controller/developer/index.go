package developer

import (
	"chaos/api/api/common"
	"chaos/api/codes"
	"chaos/api/log"
	"chaos/api/model"
	"chaos/api/system"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func Profile(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	devIdStr, ok := c.Get("dev_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var gameDeveloper model.GameDeveloper

	db.Model(&model.GameDeveloper{}).Where("id = ?", devIdStr).First(&gameDeveloper)
	if gameDeveloper.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game developer not found"
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = map[string]interface{}{
		"dev_name":     gameDeveloper.DevName,
		"email":        gameDeveloper.Email,
		"website":      gameDeveloper.Website,
		"country_code": gameDeveloper.CountryCode,
	}

	c.JSON(http.StatusOK, res)
}

func UpdateProfile(c *gin.Context) {
	var req UpdateProfileReq

	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "param error"
		c.JSON(http.StatusOK, res)
		return
	}

	devIdStr, ok := c.Get("dev_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var gameDeveloper model.GameDeveloper

	db.Model(&model.GameDeveloper{}).Where("id = ?", devIdStr).First(&gameDeveloper)
	if gameDeveloper.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game developer not found"
		c.JSON(http.StatusOK, res)
		return
	}

	gameDeveloper.DevName = req.DevName
	gameDeveloper.CountryCode = req.CountryCode
	gameDeveloper.Website = req.Website
	db.Save(&gameDeveloper)

	res.Data = nil
	c.JSON(http.StatusOK, res)
}

func ListGame(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	devIdStr, ok := c.Get("dev_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var gameDeveloper model.GameDeveloper

	db.Model(&model.GameDeveloper{}).Where("id = ?", devIdStr).First(&gameDeveloper)
	if gameDeveloper.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game developer not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameInfo []model.GameInfo
	db.Model(&model.GameInfo{}).Where("dev_id = ?", devIdStr).Find(&gameInfo)

	res.Data = gameInfo
	c.JSON(http.StatusOK, res)
}

func DetailGame(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	devIdStr, ok := c.Get("dev_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	gameIdStr, ok := c.GetQuery("game_id")
	if !ok {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "game_id is required"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var gameDeveloper model.GameDeveloper

	db.Model(&model.GameDeveloper{}).Where("id = ?", devIdStr).First(&gameDeveloper)
	if gameDeveloper.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game developer not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameInfo model.GameInfo
	var gameApp model.GameApp
	db.Model(&model.GameInfo{}).Where("dev_id = ? and id = ?", devIdStr, gameIdStr).First(&gameInfo)

	if gameInfo.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game not found"
		c.JSON(http.StatusOK, res)
		return
	}
	db.Model(&model.GameApp{}).Where("game_id = ?", gameInfo.ID).First(&gameApp)

	var gameSetting []model.GameSetting
	db.Model(&model.GameSetting{}).Where("game_id = ?", gameInfo.ID).Find(&gameSetting)

	res.Data = map[string]interface{}{
		"game_info":     gameInfo,
		"game_app":      gameApp,
		"game_settings": gameSetting,
	}
	c.JSON(http.StatusOK, res)
}

func UpdateGame(c *gin.Context) {
	var req UpdateGameReq

	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "param error"
		c.JSON(http.StatusOK, res)
		return
	}

	devIdStr, ok := c.Get("dev_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var gameDeveloper model.GameDeveloper

	db.Model(&model.GameDeveloper{}).Where("id = ?", devIdStr).First(&gameDeveloper)
	if gameDeveloper.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game developer not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameInfo model.GameInfo
	if req.ID != 0 {
		db.Model(&model.GameInfo{}).Where("id = ?", req.ID).First(&gameInfo)
		if gameInfo.ID == 0 {
			res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
			res.Msg = "game not found"
			c.JSON(http.StatusOK, res)
			return
		}
		if gameInfo.DevID != gameDeveloper.ID {
			res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
			res.Msg = "game not found, please check game"
			c.JSON(http.StatusOK, res)
			return
		}
	} else {
		//创建
		gameInfo.DevID = gameDeveloper.ID
		gameInfo.Status = model.GameStatusDraft
		gameInfo.AddTime = time.Now()
	}
	var shouldResetStatus = false
	if gameInfo.Name != req.Name {
		shouldResetStatus = true
		gameInfo.Name = req.Name
	}
	if gameInfo.Info != req.Info {
		shouldResetStatus = true
		gameInfo.Info = req.Info
	}
	if gameInfo.Description != req.Description {
		shouldResetStatus = true
		gameInfo.Description = req.Description
	}
	if gameInfo.Avatar != req.Avatar {
		shouldResetStatus = true
		gameInfo.Avatar = req.Avatar
	}
	if gameInfo.Image != req.Image {
		shouldResetStatus = true
		gameInfo.Image = req.Image
	}

	gameInfo.Info = req.Info
	gameInfo.Description = req.Description
	gameInfo.Avatar = req.Avatar
	gameInfo.Image = req.Image
	gameInfo.PlayUrl = req.PlayUrl

	if shouldResetStatus {
		if gameInfo.Status == model.GameStatusActive {
			gameInfo.Status = model.GameStatusTested
		}
	}
	err := db.Save(&gameInfo).Error
	if err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "failed to update game"
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = nil
	c.JSON(http.StatusOK, res)
}

func RefreshGameKey(c *gin.Context) {
	var req GameKeyReq

	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "param error"
		c.JSON(http.StatusOK, res)
		return
	}

	devIdStr, ok := c.Get("dev_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var gameDeveloper model.GameDeveloper

	db.Model(&model.GameDeveloper{}).Where("id = ?", devIdStr).First(&gameDeveloper)
	if gameDeveloper.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game developer not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameInfo model.GameInfo
	db.Model(&model.GameInfo{}).Where("id = ?", req.GameID).First(&gameInfo)
	if gameInfo.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameApp model.GameApp
	db.Model(&model.GameApp{}).Where("game_id = ?", req.GameID).First(&gameApp)
	if gameApp.ID == 0 {
		gameApp.AddTime = time.Now()
		gameApp.GameID = gameInfo.ID
	}

	u := uuid.New()
	var clientID = req.ClientID
	var clientSecret = u.String()

	if len(clientID) == 0 {
		clientID = generateID()
	}

	var existGameApp model.GameApp
	db.Model(&model.GameApp{}).Where("client_id = ?", clientID).First(&existGameApp)
	if existGameApp.ID > 0 && existGameApp.GameID != req.GameID {
		res.Code = codes.CODE_ERR_EXIST_OBJ
		res.Msg = "client id already exists"
		c.JSON(http.StatusOK, res)
		return
	}

	gameApp.ClientSecret = clientSecret
	gameApp.ClientID = clientID

	db.Save(&gameApp)

	res.Data = nil
	c.JSON(http.StatusOK, res)
}

func TestingStart(c *gin.Context) {
	var req GameTestingStartReq

	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "param error"
		c.JSON(http.StatusOK, res)
		return
	}

	devIdStr, ok := c.Get("dev_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var gameDeveloper model.GameDeveloper

	db.Model(&model.GameDeveloper{}).Where("id = ?", devIdStr).First(&gameDeveloper)
	if gameDeveloper.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game developer not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameInfo model.GameInfo
	db.Model(&model.GameInfo{}).Where("id = ?", req.GameID).First(&gameInfo)
	if gameInfo.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game not found"
		c.JSON(http.StatusOK, res)
		return
	}

	if gameInfo.DevID != gameDeveloper.ID {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game not found, please check game"
		c.JSON(http.StatusOK, res)
		return
	}

	if gameInfo.Status != model.GameStatusDraft {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "game is not draft"
		c.JSON(http.StatusOK, res)
		return
	}

	gameInfo.Status = model.GameStatusTesting
	db.Save(&gameInfo)

	res.Data = nil
	c.JSON(http.StatusOK, res)
}

func TestingFinish(c *gin.Context) {
	var req GameTestingStartReq

	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "param error"
		c.JSON(http.StatusOK, res)
		return
	}

	devIdStr, ok := c.Get("dev_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var gameDeveloper model.GameDeveloper

	db.Model(&model.GameDeveloper{}).Where("id = ?", devIdStr).First(&gameDeveloper)
	if gameDeveloper.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game developer not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameInfo model.GameInfo
	db.Model(&model.GameInfo{}).Where("id = ?", req.GameID).First(&gameInfo)
	if gameInfo.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game not found"
		c.JSON(http.StatusOK, res)
		return
	}

	if gameInfo.DevID != gameDeveloper.ID {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game not found, please check game"
		c.JSON(http.StatusOK, res)
		return
	}

	if gameInfo.Status != model.GameStatusTesting {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "game is not testing"
		c.JSON(http.StatusOK, res)
		return
	}

	gameInfo.Status = model.GameStatusTested
	db.Save(&gameInfo)

	res.Data = nil
	c.JSON(http.StatusOK, res)
}

func SaveSetting(c *gin.Context) {
	var req GameSettingSaveReq

	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "param error"
		c.JSON(http.StatusOK, res)
		return
	}

	devIdStr, ok := c.Get("dev_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var gameDeveloper model.GameDeveloper

	db.Model(&model.GameDeveloper{}).Where("id = ?", devIdStr).First(&gameDeveloper)
	if gameDeveloper.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game developer not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameInfo model.GameInfo
	db.Model(&model.GameInfo{}).Where("id = ?", req.GameID).First(&gameInfo)
	if gameInfo.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game not found"
		c.JSON(http.StatusOK, res)
		return
	}

	if gameInfo.DevID != gameDeveloper.ID {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game not found, please check game"
		c.JSON(http.StatusOK, res)
		return
	}

	if len(req.Code) == 0 {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "code is required"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameSetting model.GameSetting
	db.Model(&model.GameSetting{}).Where("game_id = ? and code = ?", req.GameID, req.Code).First(&gameSetting)
	if gameSetting.ID == 0 {
		gameSetting.GameID = gameInfo.ID
		gameSetting.AddTime = time.Now()
	}

	gameSetting.Code = req.Code
	gameSetting.Catalog = req.Catalog
	gameSetting.AmountPerPlay = req.AmountPerPlay.Mul(decimal.NewFromInt(1000000)).Round(0).BigInt().Uint64()
	db.Save(&gameSetting)

	if gameInfo.Status == model.GameStatusActive {
		gameInfo.Status = model.GameStatusTested
		db.Save(&gameInfo)
	}

	res.Data = nil
	c.JSON(http.StatusOK, res)
}

func DeleteSetting(c *gin.Context) {
	var req GameSettingDeleteReq

	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "param error"
		c.JSON(http.StatusOK, res)
		return
	}

	devIdStr, ok := c.Get("dev_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var gameDeveloper model.GameDeveloper

	db.Model(&model.GameDeveloper{}).Where("id = ?", devIdStr).First(&gameDeveloper)
	if gameDeveloper.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game developer not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameInfo model.GameInfo
	db.Model(&model.GameInfo{}).Where("id = ?", req.GameID).First(&gameInfo)
	if gameInfo.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game not found"
		c.JSON(http.StatusOK, res)
		return
	}

	if gameInfo.DevID != gameDeveloper.ID {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game not found, please check game"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameSetting model.GameSetting
	db.Model(&model.GameSetting{}).Where("game_id = ? and code = ?", req.GameID, req.Code).First(&gameSetting)
	if gameSetting.ID > 0 {
		db.Delete(&gameSetting)
		gameInfo.Status = model.GameStatusTested
		db.Save(&gameInfo)
	}

	res.Data = nil
	c.JSON(http.StatusOK, res)
}

func SubmitOnlineAudit(c *gin.Context) {
	var req GameTestingStartReq

	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "param error"
		c.JSON(http.StatusOK, res)
		return
	}

	devIdStr, ok := c.Get("dev_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var gameDeveloper model.GameDeveloper

	db.Model(&model.GameDeveloper{}).Where("id = ?", devIdStr).First(&gameDeveloper)
	if gameDeveloper.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game developer not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameInfo model.GameInfo
	db.Model(&model.GameInfo{}).Where("id = ?", req.GameID).First(&gameInfo)
	if gameInfo.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game not found"
		c.JSON(http.StatusOK, res)
		return
	}

	if gameInfo.DevID != gameDeveloper.ID {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game not found, please check game"
		c.JSON(http.StatusOK, res)
		return
	}

	// 检查游戏状态是否允许申请上架
	if gameInfo.Status != model.GameStatusTested {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "game status not allowed for online audit"
		c.JSON(http.StatusOK, res)
		return
	}

	// 更新游戏状态为待审核
	gameInfo.Status = model.GameStatusPendingApprove
	if err := db.Save(&gameInfo).Error; err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "update game status failed"
		c.JSON(http.StatusOK, res)
		return
	}

	res.Msg = "online audit submitted successfully"
	c.JSON(http.StatusOK, res)
}

func GameSessionList(c *gin.Context) {
	var req GameSessionListReq
	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "param error"
		c.JSON(http.StatusOK, res)
		return
	}

	devIdStr, ok := c.Get("dev_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	var pn = 1
	if req.Page > 0 {
		pn = req.Page
	}
	var pageSize = 20
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	db := system.GetDb()
	var gameDeveloper model.GameDeveloper

	db.Model(&model.GameDeveloper{}).Where("id = ?", devIdStr).First(&gameDeveloper)
	if gameDeveloper.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game developer not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameSessionList []model.GameSession
	query := db.Table("n_game_session ngs").
		Joins("LEFT JOIN n_game_info gi ON ngs.game_id = gi.id").
		Where("gi.dev_id = ?", devIdStr)
	if req.GameID > 0 {
		query = query.Where("ngs.game_id = ?", req.GameID)
	}
	if req.Status > 0 {
		query = query.Where("ngs.status = ?", req.Status)
	}
	if req.Testing != nil {
		query = query.Where("ngs.testing = ?", *req.Testing)
	}
	if req.StartDate != "" {
		query = query.Where("ngs.start_time >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("ngs.start_time <= ?", req.EndDate)
	}

	var total int64
	query.Count(&total)

	log.Info("current developer id:", devIdStr)

	query = query.Order("ngs.start_time DESC")

	query.Offset((pn - 1) * pageSize).Limit(pageSize).Find(&gameSessionList)

	// stmt := query.Session(&gorm.Session{DryRun: true}).Select("ngs.*").Find(&[]model.GameSession{}).Statement

	// log.Infof("Generated SQL: %s; Vars: %v", stmt.SQL.String(), stmt.Vars)

	err := query.Select("ngs.*").Find(&gameSessionList).Error
	if err != nil {
		log.Error("query game session error:", err)
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "query failed"
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = gin.H{
		"list":      gameSessionList,
		"total":     total,
		"page":      pn,
		"page_size": pageSize,
	}
	c.JSON(http.StatusOK, res)
}

/** tools **/
func randomLetters(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

// 随机生成数字串
func randomDigits(n int) string {
	digits := []rune("0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = digits[rand.Intn(len(digits))]
	}
	return string(s)
}

// 生成最终的字符串
func generateID() string {
	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())

	// 随机长度 5–8
	letterLen := rand.Intn(4) + 5 // [5,8]
	digitLen := rand.Intn(4) + 5  // [5,8]

	return fmt.Sprintf("%s-app-%s", randomLetters(letterLen), randomDigits(digitLen))
}
