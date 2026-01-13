package home

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"chaos/api/api/common"
	"chaos/api/codes"
	"chaos/api/log"
	"chaos/api/model"
	"chaos/api/system"
	"chaos/api/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func Public(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	db := system.GetDb()
	var pagePushes []model.PagePush
	err := db.Table("n_page_push pp").
		Where("pp.home_visible = ?", 1).
		Scan(&pagePushes).Error
	if err != nil {
		log.Error("load page push error", err)
	}

	res.Data = gin.H{
		"rpc": map[string]string{
			// "Solana":   "https://mainnet.helius-rpc.com/?api-key=4b1030d1-e346-4788-a65d-29c065efa012",
			// "Ethereum": "https://eth.llamarpc.com",
			// "Base":     "https://base-mainnet.infura.io/v3/15d81a19824c41159daec8327f691720",
			// "Arbitrum": "https://arbitrum-mainnet.infura.io/v3/15d81a19824c41159daec8327f691720",
			"Bsc": "https://binance.llamarpc.com",
		},
		"page_push": pagePushes,
	}

	c.JSON(http.StatusOK, res)
}

func Assets(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	db := system.GetDb()
	var assets []model.SysAsset
	err := db.Find(&assets).Error
	if err != nil {
		log.Error("load assets error", err)
	}

	res.Data = gin.H{
		"assets": assets,
	}

	c.JSON(http.StatusOK, res)
}

/*********************** TODO ***********************/
func Leaderboard(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	season_code := c.Param("season_code")
	var season model.SeasonInfo

	db := system.GetDb()
	if len(season_code) > 0 {
		db.Model(&model.SeasonInfo{}).Where("code = ?", season_code).First(&season)
	} else {
		err := db.Table("season_info s").
			Where("s.status = ?", []string{model.SeasonStatusActive}).
			Where("s.is_visible = ?", model.SeasonIsVisible).
			Order("id desc").
			Limit(1).
			Scan(&season).Error
		if err != nil {
			log.Error("load season info error", err)
		}
	}

	if season.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "season not found"
		c.JSON(http.StatusOK, res)
		return
	}
	if season.Status != model.SeasonStatusActive {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "season not active"
		c.JSON(http.StatusOK, res)
		return
	}

	var seasonGames []model.SeasonGame
	db.Model(&model.SeasonGame{}).Where("season_id = ?", season.ID).Find(&seasonGames)
	if len(seasonGames) == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "season game not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var scoreCaseSql = "(CASE p.game_id\n"
	var amountCaseSql = "(CASE p.game_id\n"
	for _, seasonGame := range seasonGames {
		scoreCaseSql += fmt.Sprintf("WHEN %d THEN %d\n", seasonGame.GameID, seasonGame.WeightScore)
		amountCaseSql += fmt.Sprintf("WHEN %d THEN %d\n", seasonGame.GameID, seasonGame.WeightAmount)
	}
	scoreCaseSql += "ELSE 0\n"
	amountCaseSql += "ELSE 0\n"
	scoreCaseSql += "END) * p.sum_score \n"
	amountCaseSql += "END) * p.sum_amount\n"

	var sql = fmt.Sprintf(seasonLeaderboardSql, scoreCaseSql, amountCaseSql)
	log.Info("query sql: ", sql)

	var seasonLeaderboard []SeasonLeaderboard
	db.Raw(sql, season.ID).Scan(&seasonLeaderboard)

	var _100 = decimal.NewFromInt(100)
	for i := range seasonLeaderboard {
		seasonLeaderboard[i].UserNo = utils.MaskUserNo(seasonLeaderboard[i].UserNo)
		seasonLeaderboard[i].Email = utils.MaskEmail(seasonLeaderboard[i].Email)
		seasonLeaderboard[i].WeightedScore = seasonLeaderboard[i].WeightedScore.Div(_100).Round(2)
	}

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = gin.H{
		"leaderboard": seasonLeaderboard,
	}

	c.JSON(http.StatusOK, res)
}

func LeaderboardGame(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	season_code := c.Param("season_code")
	game_id := c.Param("game_id")

	var season model.SeasonInfo
	db := system.GetDb()
	if len(season_code) > 0 {
		db.Model(&model.SeasonInfo{}).Where("code = ?", season_code).First(&season)
	} else {
		err := db.Table("season_info s").
			Where("s.status = ?", []string{model.SeasonStatusActive}).
			Where("s.is_visible = ?", model.SeasonIsVisible).
			Order("id desc").
			Limit(1).
			Scan(&season).Error
		if err != nil {
			log.Error("load season info error", err)
		}
	}

	if season.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "season not found"
		c.JSON(http.StatusOK, res)
		return
	}
	if season.Status != model.SeasonStatusActive {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "season not active"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameSessions []GameSessionWithUser
	// err := db.Table("n_game_session ngs").
	// 	Select("ngs.session_id, ngs.start_time, ngs.end_time, ngs.score, ngs.user_report_score, ngs.spend_amount_n, um.user_no, um.email, up.name, up.avatar, up.country_code").
	// 	Joins("LEFT JOIN n_user_main um ON um.id = ngs.main_id").
	// 	Joins("LEFT JOIN n_user_profile up ON up.main_id = ngs.main_id").
	// 	Where("ngs.status = ? and ngs.game_id = ?", model.GameSessionStatusSettled, game_id).
	// 	Order("ngs.score desc").Limit(20).Scan(&gameSessions).Error

	sub := `
  SELECT
    s.session_id,
    s.main_id,
    s.start_time,
    s.end_time,
    s.score,
    ROW_NUMBER() OVER (
      PARTITION BY s.main_id
      ORDER BY s.score DESC, s.end_time ASC  -- 同分时取更早结束
    ) AS rn
  FROM n_game_session s
  WHERE s.status = ? AND s.game_id = ? AND s.testing = 0
`

	err := db.Table("( "+sub+" ) AS x", model.GameSessionStatusSettled, game_id).
		Select(`
		x.start_time,
		x.end_time,
		x.score AS max_score,
		um.user_no,
		um.email,
		up.name,
		up.avatar,
		up.country_code
	`).
		Joins("LEFT JOIN n_user_main um ON um.id = x.main_id").
		Joins("LEFT JOIN n_user_profile up ON up.main_id = x.main_id").
		Where("x.rn = 1").
		Order("x.score DESC, x.end_time ASC").
		Limit(20).
		Scan(&gameSessions).Error
	if err != nil {
		log.Error("load game session error", err)
	}
	for i := range gameSessions {
		gameSessions[i].UserNo = utils.MaskUserNo(gameSessions[i].UserNo)
		gameSessions[i].Email = utils.MaskEmail(gameSessions[i].Email)
	}

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = gin.H{
		"leaderboard": gameSessions,
	}

	c.JSON(http.StatusOK, res)
}

func Game(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	db := system.GetDb()
	var games []model.GameWithTags

	err := db.Table("game_info g").
		Select(`
        g.*,
        COALESCE(JSON_ARRAYAGG(c.name), JSON_ARRAY()) AS tags
    `).
		Joins("LEFT JOIN game_tag gt ON gt.game_id = g.id").
		Joins("LEFT JOIN game_category c ON c.id = gt.cat_id").
		Where("g.status IN ?", []string{model.GameStatusActive, model.GameStatusWaitingOnline}).
		Group("g.id").
		Limit(200).
		Scan(&games).Error

	if err != nil {
		log.Error("load game info error", err)
	}

	// 把 JSON string 解析到 []string
	for i := range games {
		var arr []string
		if err := json.Unmarshal([]byte(games[i].TagsRaw), &arr); err == nil {
			games[i].Tags = arr
		}
	}

	res.Data = games

	c.JSON(http.StatusOK, res)
}

func GameDetail(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	gameIdStr := c.Param("game_id")
	gameID, err := strconv.ParseUint(gameIdStr, 10, 64)
	if err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "game_id is required"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var game model.GameWithTags

	err = db.Table("game_info g").
		Select(`
        g.*,
        COALESCE(JSON_ARRAYAGG(c.name), JSON_ARRAY()) AS tags
    `).
		Joins("LEFT JOIN game_tag gt ON gt.game_id = g.id").
		Joins("LEFT JOIN game_category c ON c.id = gt.cat_id").
		Where("g.id = ? and g.status IN ?", gameID, []string{model.GameStatusActive, model.GameStatusWaitingOnline}).
		Group("g.id").
		First(&game).Error

	if err != nil {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game not found"
		c.JSON(http.StatusOK, res)
		return
	}

	// query game developer information
	var gameDeveloper model.GameDeveloper
	err = db.Table("game_developer gd").Where("gd.id = ?", game.DevID).First(&gameDeveloper).Error
	if err != nil {
		log.Error("load game developer info error", err)
	}

	// query game season information
	var gameSeason *model.SeasonInfo
	err = db.Table("season_info s").
		Joins("LEFT JOIN season_game sg ON s.id = sg.season_id").
		Where("s.status = ?", model.SeasonStatusActive).
		Where("sg.game_id = ?", gameID).
		Order("s.id desc").
		Limit(1).
		First(&gameSeason).Error
	if err != nil {
		log.Error("load game season info error", err)
	}

	res.Data = gin.H{
		"game_info":   game,
		"game_season": gameSeason,
		"game_developer": gin.H{
			"dev_name":     gameDeveloper.DevName,
			"country_code": gameDeveloper.CountryCode,
			"website":      gameDeveloper.Website,
			"avatar":       "",
		},
	}

	c.JSON(http.StatusOK, res)
}

func Season(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	db := system.GetDb()
	var season model.SeasonInfo

	err := db.Table("season_info s").
		Where("s.status = ?", []string{model.SeasonStatusActive}).
		Where("s.is_visible = ?", model.SeasonIsVisible).
		Order("id desc").
		Limit(1).
		Scan(&season).Error

	if err != nil {
		log.Error("load season info error", err)
	}

	var games []model.GameWithSeason
	err = db.Table("game_info g").Joins("LEFT JOIN season_game sg ON g.id = sg.game_id").
		Where("sg.season_id = ?", season.ID).
		Where("g.status IN ?", []string{model.GameStatusActive, model.GameStatusWaitingOnline}).
		Select("g.*, sg.season_id as season_id").
		Scan(&games).Error
	if err != nil {
		log.Error("load game info error", err)
	}

	for j := range games {
		season.IncludeGames = append(season.IncludeGames, games[j].GameInfo)
	}

	var countJoined int64
	err = db.Model(&model.SeasonUser{}).Where("season_id = ?", season.ID).Count(&countJoined).Error
	if err != nil {
		log.Error("count season user error", err)
	}
	season.JoinCount = uint64(countJoined)

	res.Data = []model.SeasonInfo{season}

	c.JSON(http.StatusOK, res)
}
