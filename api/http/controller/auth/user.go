package auth

import (
	"chaos/api/api/common"
	"chaos/api/api/http/controller/preauth"
	"chaos/api/api/service"
	"chaos/api/chain"
	"chaos/api/codes"
	"chaos/api/log"
	"chaos/api/model"
	"chaos/api/system"
	"chaos/api/tools"
	"chaos/api/utils"
	"context"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"os"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var userSeasonLock sync.Map
var userRequestLock sync.Map

const expiryDuration = 10 * time.Hour

func Profile(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	mainIdStr, ok := c.Get("main_id")
	_, _ = c.Get("provider_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var userMain model.UserMain
	var userProfile model.UserProfile
	var userRef model.UserRef
	var userProviders []model.UserProvider
	db.Model(&model.UserMain{}).Where("id = ?", mainIdStr).First(&userMain)
	if userMain.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user not found"
		c.JSON(http.StatusOK, res)
		return
	}
	db.Model(&model.UserProfile{}).Where("main_id = ?", mainIdStr).First(&userProfile)
	db.Model(&model.UserRef{}).Where("main_id = ?", mainIdStr).First(&userRef)
	db.Model(&model.UserProvider{}).Where("main_id = ?", mainIdStr).Find(&userProviders)

	res.Data = map[string]interface{}{
		"user_main":      userMain,
		"user_profile":   userProfile,
		"user_ref":       userRef.RefCode,
		"user_providers": userProviders,
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

	mainIdStr, ok := c.Get("main_id")
	_, _ = c.Get("provider_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var userMain model.UserMain
	var userProfile model.UserProfile

	db.Model(&model.UserMain{}).Where("id = ?", mainIdStr).First(&userMain)
	if userMain.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user not found"
		c.JSON(http.StatusOK, res)
		return
	}
	db.Model(&model.UserProfile{}).Where("main_id = ?", mainIdStr).First(&userProfile)

	if userProfile.ID == 0 {
		userProfile.MainID = userMain.ID
	}
	if len(req.Avatar) > 0 {
		userProfile.Avatar = req.Avatar
	}
	if len(req.Name) > 0 {
		userProfile.Name = req.Name
	}
	if len(req.Bio) > 0 {
		userProfile.Bio = req.Bio
	}

	if req.Birthday != "" {
		if birthday, err := time.Parse("2006-01-02", req.Birthday); err == nil {
			userProfile.Birthday = birthday.Format("2006-01-02")
		}
	}
	userProfile.CountryCode = req.CountryCode
	userProfile.Timezone = req.Timezone
	// userProfile.XUri = req.XUri

	err := db.Save(&userProfile).Error
	if err != nil {
		log.Error(err)
	}

	c.JSON(http.StatusOK, res)
}

func SendEmail(c *gin.Context) {
	var req SendEmailReq
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

	mainIdStr, ok := c.Get("main_id")
	_, _ = c.Get("provider_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var userMain model.UserMain

	db.Model(&model.UserMain{}).Where("id = ?", mainIdStr).First(&userMain)
	if userMain.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user not found"
		c.JSON(http.StatusOK, res)
		return
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		res.Code = codes.CODE_ERR_REQFORMAT
		res.Msg = "Invalid email format"
		c.JSON(http.StatusOK, res)
		return
	}

	if strings.EqualFold(req.Email, tools.Deref(userMain.Email)) {
		res.Code = codes.CODE_ERR_REQFORMAT
		res.Msg = "email already binded"
		c.JSON(http.StatusOK, res)
		return
	}

	var existUserMain model.UserMain
	db.Model(&model.UserMain{}).Where("email = ?", req.Email).First(&existUserMain)
	if existUserMain.ID != 0 {
		res.Code = codes.CODE_ERR_REQFORMAT
		res.Msg = "email already exists"
		c.JSON(http.StatusOK, res)
		return
	}

	err := utils.SendVerifyCodeMailAPIWithUserMainId(req.Email, "10", userMain.ID)
	if err != nil {
		log.Error("send email err", err)
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "send email failed"
		c.JSON(http.StatusOK, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

func VerifyEmail(c *gin.Context) {
	var req VerifyEmailReq
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

	mainIdStr, ok := c.Get("main_id")
	_, _ = c.Get("provider_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var userMain model.UserMain

	db.Model(&model.UserMain{}).Where("id = ?", mainIdStr).First(&userMain)
	if userMain.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user not found"
		c.JSON(http.StatusOK, res)
		return
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		res.Code = codes.CODE_ERR_REQFORMAT
		res.Msg = "Invalid email format"
		c.JSON(http.StatusOK, res)
		return
	}

	var verifyProcess model.VerificationProcess
	db.Model(&model.VerificationProcess{}).
		Where("target = ? and code = ? and type = ? and sort = ?", req.Email, req.Code, "10", "10").
		First(&verifyProcess)
	if verifyProcess.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "verification code not sent"
		c.JSON(http.StatusOK, res)
		return
	}

	if verifyProcess.MainID != userMain.ID {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "verification code not sent"
		c.JSON(http.StatusOK, res)
		return
	}

	if time.Now().After(verifyProcess.AddTime.Add(time.Duration(verifyProcess.ValidatePeriod) * time.Second)) {
		res.Code = codes.CODE_ERR_REQ_EXPIRED
		res.Msg = "verification code expired"
		c.JSON(http.StatusOK, res)
		return
	}

	userMain.Email = &req.Email
	userMain.Status = "20"
	db.Save(&userMain)

	c.JSON(http.StatusOK, res)
}

func VerifyWallet(c *gin.Context) {
	var req preauth.VerifyAuthRequest
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_REQFORMAT
		res.Msg = "invalid request" + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	mainIdStr, ok := c.Get("main_id")
	_, _ = c.Get("provider_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var userMain model.UserMain

	db.Model(&model.UserMain{}).Where("id = ?", mainIdStr).First(&userMain)
	if userMain.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var authObj model.AuthMessage
	err := db.Model(&model.AuthMessage{}).
		Where("id = ?", req.ID).
		First(&authObj).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
			res.Msg = "record not found"
			c.JSON(http.StatusOK, res)
			return
		}
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	log.Info("verify message obj is: ", authObj.ID, authObj.AuthKey, authObj.AuthMsg, authObj.ExpireTime)
	log.Infof("verify message req is: %v", req)
	if authObj.ExpireTime.Before(time.Now()) {
		res.Code = codes.CODE_ERR_REQ_EXPIRED
		res.Msg = "please get a new message"
		c.JSON(http.StatusOK, res)
		return
	}

	var existUserProvider model.UserProvider
	db.Model(&model.UserProvider{}).Where("provider_id = ? and provider_type = ?", authObj.AuthKey, "wallet").First(&existUserProvider)
	if existUserProvider.ID != 0 && existUserProvider.MainID != userMain.ID {
		res.Code = codes.CODE_ERR_REQFORMAT
		res.Msg = "wallet already binded"
		c.JSON(http.StatusOK, res)
		return
	}

	// start to verify message
	if !authObj.ComputeAuthDigest(req.Sign) {
		res.Code = codes.CODE_ERR_SIG_COMMON
		res.Msg = "invalid sign"
		c.JSON(http.StatusOK, res)
		return
	}

	db.Model(&model.UserProvider{}).Where("main_id = ? and provider_type = ?", mainIdStr, "wallet").First(&existUserProvider)

	existUserProvider.MainID = userMain.ID
	existUserProvider.ProviderID = authObj.AuthKey
	existUserProvider.ProviderType = "wallet"
	existUserProvider.ProviderLabel = "bsc"
	existUserProvider.AddTime = time.Now()
	db.Save(&existUserProvider)

	c.JSON(http.StatusOK, res)
}

func QueryBalance(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	mainIdStr, ok := c.Get("main_id")
	_, _ = c.Get("provider_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var userMain model.UserMain

	db.Model(&model.UserMain{}).Where("id = ?", mainIdStr).First(&userMain)
	if userMain.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var userProvider model.UserProvider
	db.Model(&model.UserProvider{}).Where("main_id = ? and provider_type = ?", mainIdStr, "wallet").First(&userProvider)

	if userProvider.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user provider not found, please bind wallet first"
		c.JSON(http.StatusOK, res)
		return
	}

	n, err := tools.GetGlobalClient().GetNToken()

	var chainBalance *big.Int
	if err != nil {
		log.Error("can not get n", err)
	} else {
		chainBalance, err = n.BalanceOf(context.Background(), userProvider.ProviderID)
		if err != nil {
			log.Error("can not get n balance", err)
		}
	}

	accountBalance, err := service.QueryAccountBalance(userMain.ID, 0)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error("can not get account balance", err)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		accountBalance = model.AccountBalance{
			Available:  0,
			Frozen:     0,
			MainID:     userMain.ID,
			AssetID:    0,
			UpdateTime: time.Now(),
		}
		err = db.Create(&accountBalance).Error
		if err != nil {
			log.Error("can not create account balance", err)
		}
	}

	res.Data = gin.H{
		"chain_balance": chainBalance,
		"account_balance": map[string]interface{}{
			"available":         accountBalance.Available,
			"frozen":            accountBalance.Frozen,
			"pendingWithdrawal": accountBalance.Withdrawal,
		},
	}

	c.JSON(http.StatusOK, res)
}

func FetchTransactions(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	mainIdStr, ok := c.Get("main_id")
	_, _ = c.Get("provider_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	pnStr := c.DefaultQuery("pn", "1")
	pn, err := strconv.Atoi(pnStr)
	if err != nil {
		pn = 1
	}
	typeStr := c.DefaultQuery("type", "all")
	typeCode := model.AvailFlowType(typeStr)

	db := system.GetDb()
	query := db.Model(&model.AccountFlow{}).
		Where("main_id = ? AND asset_id = ?", mainIdStr, 0)

	if typeCode != -1 {
		query = query.Where("biz_type = ?", typeCode)
	}

	limit := 20
	offset := (pn - 1) * limit

	// Build query conditions
	query = query.Model(&model.AccountFlow{}).Order("add_time desc")

	// Get total count
	var total int64
	query.Count(&total)

	var flows []model.AccountFlow
	err = query.Order("add_time desc").Offset(offset).Limit(limit).Find(&flows).Error
	if err != nil {
		log.Error("query account flow error:", err)
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "query failed"
		c.JSON(http.StatusOK, res)
		return
	}

	// Format return data
	var transactions []gin.H
	for _, flow := range flows {
		// Map biz_type to frontend type
		var txType string
		switch flow.BizType {
		case model.FlowRecharge:
			txType = "deposit"
		case model.FlowWithdraw:
			txType = "withdraw"
		case model.FlowFreeze:
			txType = "freeze"
		case model.FlowUnfreeze:
			txType = "unfreeze"
		case model.FlowSpend:
			txType = "spend"
		case model.FlowRefund:
			txType = "refund"
		default:
			txType = fmt.Sprintf("%d", flow.BizType)
		}

		// Determine status based on direction
		var status string
		if flow.Status == 1 {
			status = "confirmed"
		} else if flow.Status == 0 {
			status = "pending"
		} else {
			status = "failed"
		}

		transactions = append(transactions, gin.H{
			"id":        flow.ID,
			"type":      txType,
			"token":     "N",
			"amount":    fmt.Sprintf("%.6f", float64(flow.Amount)/1000000), // Assuming 6 decimal precision
			"timestamp": flow.AddTime.Format("2006-01-02T15:04:05Z"),
			"status":    status,
		})
	}

	res.Data = gin.H{
		"transactions": transactions,
		"pagination": gin.H{
			"page":     pn,
			"limit":    limit,
			"total":    total,
			"has_more": offset+limit < int(total),
		},
	}

	c.JSON(http.StatusOK, res)
}

func QueryGameSession(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	mainIdStr, ok := c.Get("main_id")
	_, _ = c.Get("provider_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	pnStr := c.DefaultQuery("pn", "1")
	pn, err := strconv.Atoi(pnStr)
	if err != nil {
		pn = 1
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	statusStr := c.DefaultQuery("status", "")
	status := model.GameSessionStatus(statusStr)

	type GameSessionWithGameInfo struct {
		model.GameSession
		GameInfo model.GameInfo `gorm:"embedded"`
	}

	db := system.GetDb()
	baseQuery := db.Table("n_game_session ngs").Joins("LEFT JOIN n_game_info gi ON ngs.game_id = gi.id").
		Where("ngs.main_id = ?", mainIdStr)
	if status != 0 {
		baseQuery = baseQuery.Where("ngs.status = ?", status)
	}

	// Get total count first (before applying limit/offset)
	var total int64
	err = baseQuery.Count(&total).Error
	if err != nil {
		log.Error("count game session error:", err)
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "count failed"
		c.JSON(http.StatusOK, res)
		return
	}

	offset := (pn - 1) * limit

	var gameSessions []GameSessionWithGameInfo
	err = baseQuery.Select("ngs.*, gi.*").Order("ngs.start_time desc").Offset(offset).Limit(limit).Find(&gameSessions).Error
	if err != nil {
		log.Error("query game session error:", err)
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "query failed"
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = gin.H{
		"game_sessions": gameSessions,
		"pagination": gin.H{
			"page":     pn,
			"limit":    limit,
			"total":    total,
			"has_more": offset+limit < int(total),
		},
	}

	c.JSON(http.StatusOK, res)
}

func ConfirmGameSession(c *gin.Context) {
	var req ConfirmGameSessionReq
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "param error"
		c.JSON(http.StatusOK, res)
		return
	}

	mainIdStr, ok := c.Get("main_id")
	_, _ = c.Get("provider_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var gameSession model.GameSession
	err := db.Model(&model.GameSession{}).Where("session_id = ?", req.SessionID).First(&gameSession).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "query game session failed"
		c.JSON(http.StatusOK, res)
		return
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game session not found"
		c.JSON(http.StatusOK, res)
		return
	}

	if fmt.Sprintf("%d", gameSession.MainID) != mainIdStr {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "game session not owned by user"
		c.JSON(http.StatusOK, res)
		return
	}
	if gameSession.Status != model.GameSessionStatusEnd {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "game session not ended, or not in correct status"
		c.JSON(http.StatusOK, res)
		return
	}

	gameSession.UserReportScore = req.Score
	if !gameSession.Score.Equal(req.Score) {
		gameSession.Status = model.GameSessionStatusAuditing
	} else {
		gameSession.Status = model.GameSessionStatusSettled
	}
	if err := db.Save(&gameSession).Error; err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "update game session failed"
		c.JSON(http.StatusOK, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

func JoinSeason(c *gin.Context) {
	var req JoinSeasonReq
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "param error"
		c.JSON(http.StatusOK, res)
		return
	}

	mainIdStr, ok := c.Get("main_id")
	_, _ = c.Get("provider_id")

	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var seasonInfo model.SeasonInfo
	err := db.Model(&model.SeasonInfo{}).Where("id = ?", req.SeasonID).First(&seasonInfo).Error
	if err != nil {
		log.Error("query season info failed", err)
	}
	if seasonInfo.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "season not found"
		c.JSON(http.StatusOK, res)
		return
	}

	if seasonInfo.Status != model.SeasonStatusActive {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "season not active for join"
		c.JSON(http.StatusOK, res)
		return
	}

	var seasonUserKey string
	if seasonInfo.LimitUserCount == 0 {
		seasonUserKey = fmt.Sprintf("%s_%d", mainIdStr, req.SeasonID)
	} else {
		seasonUserKey = fmt.Sprintf("season_%d", req.SeasonID)
	}

	// var seasonUserKey = fmt.Sprintf("%s_%d", "mainIdStr", req.SeasonID)
	_, loaded := userSeasonLock.LoadOrStore(seasonUserKey, struct{}{})
	defer func() {
		userSeasonLock.Delete(seasonUserKey)
	}()
	if loaded {
		res.Code = codes.CODE_ERR_PROCESSING
		res.Msg = "operating, please do not repeat the operation"
		c.JSON(http.StatusOK, res)
		return
	}

	if seasonInfo.LimitUserCount > 0 {
		var countJoined int64
		err = db.Model(&model.SeasonUser{}).Where("season_id = ?", req.SeasonID).Count(&countJoined).Error
		if err != nil {
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "query season user failed, please try again later"
			c.JSON(http.StatusOK, res)
			return
		}
		if uint64(countJoined) >= seasonInfo.LimitUserCount {
			res.Code = codes.CODE_ERR_BAD_PARAMS
			res.Msg = "season user count limit reached"
			c.JSON(http.StatusOK, res)
			return
		}
	}

	var existSeasonUser model.SeasonUser
	err = db.Model(&model.SeasonUser{}).Where("season_id = ? and main_id = ?", req.SeasonID, mainIdStr).First(&existSeasonUser).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "query season user failed, please try again later"
		c.JSON(http.StatusOK, res)
		return
	}
	if existSeasonUser.ID == 0 {
		mainId, err := strconv.ParseUint(mainIdStr.(string), 10, 64)
		if err != nil {
			res.Code = codes.CODE_ERR_BAD_PARAMS
			res.Msg = "unexpected user lookup error"
			c.JSON(http.StatusOK, res)
			return
		}
		existSeasonUser.SeasonID = req.SeasonID
		existSeasonUser.MainID = mainId
		existSeasonUser.JoinTime = time.Now()
		err = db.Save(&existSeasonUser).Error
		if err != nil {
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "save season user failed, please try again later"
			c.JSON(http.StatusOK, res)
			return
		}
	}

	c.JSON(http.StatusOK, res)
}

func CheckSeasonJoined(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	mainIdStr, ok := c.Get("main_id")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	seasonIdStr := c.Query("season_id")
	if seasonIdStr == "" {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "season_id required"
		c.JSON(http.StatusOK, res)
		return
	}

	seasonId, err := strconv.ParseUint(seasonIdStr, 10, 64)
	if err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "invalid season_id"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()
	var existSeasonUser model.SeasonUser
	err = db.Model(&model.SeasonUser{}).Where("season_id = ? and main_id = ?", seasonId, mainIdStr).First(&existSeasonUser).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "query season user failed"
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = gin.H{
		"joined":    existSeasonUser.ID > 0,
		"join_time": existSeasonUser.JoinTime,
	}

	c.JSON(http.StatusOK, res)
}

/*
type: recharge/withdraw/lock
*/
func BalanceTopupReport(c *gin.Context) {
	var req TopupReq
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

	log.Infof("BalanceTopupReport [start] request: %v", req)

	if req.Type != "recharge" && req.Type != "withdraw" && req.Type != "lock" && req.Type != "cancel" {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "invalid type"
		c.JSON(http.StatusOK, res)
		return
	}

	mainIdStr, ok := c.Get("main_id")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	if (req.Type == "withdraw" || req.Type == "lock" || req.Type == "cancel") && req.RefID == 0 {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "ref_id required"
		c.JSON(http.StatusOK, res)
		return
	}

	var hashPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{64}$`)
	if !hashPattern.MatchString(req.TxHash) {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "invalid tx hash"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()

	var userMain model.UserMain
	err := db.Model(&model.UserMain{}).Where("id = ?", mainIdStr).First(&userMain).Error
	if err != nil {
		log.Error("query user main failed", err)
	}
	if userMain.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user not found"
		c.JSON(http.StatusOK, res)
		return
	}
	var existAccountBalanceFlowForLocOrWithdraw model.AccountBalanceFlow
	if req.Type == "withdraw" || req.Type == "lock" || req.Type == "cancel" {
		// says user had request lock ticket and sign open lock transaction
		err = db.Model(&model.AccountBalanceFlow{}).Where("id = ?", req.RefID).First(&existAccountBalanceFlowForLocOrWithdraw).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error("query account balance flow failed", err, req.RefID)
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "query account balance flow failed"
			c.JSON(http.StatusOK, res)
			return
		}
		if existAccountBalanceFlowForLocOrWithdraw.ID == 0 {
			res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
			res.Msg = "account balance flow not found"
			c.JSON(http.StatusOK, res)
			return
		}
		switch req.Type {
		case "lock":
			// update lock ticket chain transaction hash
			existAccountBalanceFlowForLocOrWithdraw.TxHash = req.TxHash
			existAccountBalanceFlowForLocOrWithdraw.UpdateTime = time.Now()
			err = db.Save(&existAccountBalanceFlowForLocOrWithdraw).Error
			if err != nil {
				res.Code = codes.CODE_ERR_UNKNOWN
				res.Msg = "update account balance flow failed"
				c.JSON(http.StatusOK, res)
				return
			}

			// notify chain to update lock ticket status
			chain.AppendTopupTx(req.ChainID, req.TxHash, userMain.ID, existAccountBalanceFlowForLocOrWithdraw.ID, model.BalanceFlowOpFreeze)
			c.JSON(http.StatusOK, res)
			return
		case "cancel":
			var existAccountBalanceFlow = model.AccountBalanceFlow{
				MainID:     userMain.ID,
				TxHash:     req.TxHash,
				AddTime:    time.Now(),
				UpdateTime: time.Now(),
				Amount:     existAccountBalanceFlowForLocOrWithdraw.RealAmount,
				Op:         model.BalanceFlowOpUnfreeze,
				Status:     model.BalanceFlowStatusPending,
				ChainID:    fmt.Sprintf("%d", req.ChainID),
				RefFlowID:  existAccountBalanceFlowForLocOrWithdraw.ID,
			}
			err = db.Save(&existAccountBalanceFlow).Error
			if err != nil {
				log.Error("update account balance flow failed", err)
				res.Code = codes.CODE_ERR_UNKNOWN
				res.Msg = "update account balance flow failed"
				c.JSON(http.StatusOK, res)
				return
			}
			// notify chain to update lock ticket status
			chain.AppendTopupTx(req.ChainID, req.TxHash, userMain.ID, existAccountBalanceFlow.ID, model.BalanceFlowOpUnfreeze)
			c.JSON(http.StatusOK, res)
			return
		}
	}

	err = db.Model(&model.AccountBalanceFlow{}).
		Where("main_id = ? and tx_hash = ?", userMain.ID, req.TxHash).
		First(&existAccountBalanceFlowForLocOrWithdraw).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error("query account balance flow failed", err, userMain.ID, req.TxHash)
	}
	if existAccountBalanceFlowForLocOrWithdraw.ID > 0 && existAccountBalanceFlowForLocOrWithdraw.ID != req.RefID {
		res.Code = codes.CODE_ERR_REPEAT
		res.Msg = "tx hash already exists"
		c.JSON(http.StatusOK, res)
		return
	}

	amount := req.Amount.Mul(decimal.NewFromInt(1000000)).Round(0).BigInt().Uint64()
	if req.Type == "withdraw" {
		amount = existAccountBalanceFlowForLocOrWithdraw.RealAmount
	}

	var op int
	if req.Type == "recharge" {
		op = model.BalanceFlowOpRecharge
	} else {
		op = model.BalanceFlowOpWithdraw
	}

	var existAccountBalanceFlow = model.AccountBalanceFlow{
		MainID:     userMain.ID,
		TxHash:     req.TxHash,
		AddTime:    time.Now(),
		UpdateTime: time.Now(),
		Amount:     amount,
		Op:         op,
		Status:     model.BalanceFlowStatusPending,
		ChainID:    fmt.Sprintf("%d", req.ChainID),
		RefFlowID:  existAccountBalanceFlowForLocOrWithdraw.ID,
	}

	tx := db.Begin()
	committed := false
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorf("BalanceTopupReport [panic] transaction report: %v failed: %v", req, r)
			return
		}
		if !committed {
			log.Errorf("BalanceTopupReport [rollback] transaction report: %v", req)
			tx.Rollback()
		}
	}()
	var userAccount model.AccountBalance
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("main_id = ? and asset_id = ?", userMain.ID, 0).First(&userAccount).Error
	if err != nil {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user account not found"
		c.JSON(http.StatusOK, res)
		return
	}
	if req.Type == "lock" {
		userAccount.Withdrawal += amount
	}

	userAccount.UpdateTime = time.Now()
	if err := tx.Save(&userAccount).Error; err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "update user account balance failed"
		c.JSON(http.StatusOK, res)
		return
	}

	if err = tx.Create(&existAccountBalanceFlow).Error; err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "handle topup transaction failed, please try report it manually"
		c.JSON(http.StatusOK, res)
		return
	}

	if err := tx.Commit().Error; err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "commit transaction failed"
		c.JSON(http.StatusOK, res)
		return
	}

	committed = true

	chain.AppendTopupTx(req.ChainID, req.TxHash, userMain.ID, existAccountBalanceFlow.ID, op)

	log.Infof("BalanceTopupReport [success] transaction report: %v", req)
	c.JSON(http.StatusOK, res)
}

func BalanceTopupList(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	pnStr := c.DefaultQuery("pn", "1")
	pn, err := strconv.Atoi(pnStr)
	if err != nil {
		pn = 1
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	mainIdStr, ok := c.Get("main_id")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	offset := (pn - 1) * limit
	db := system.GetDb()
	var accountBalanceFlows []model.AccountBalanceFlow
	query := db.Model(&model.AccountBalanceFlow{}).Where("main_id = ?", mainIdStr)

	var total int64
	err = query.Count(&total).Error
	if err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "query account balance flow failed"
		c.JSON(http.StatusOK, res)
		return
	}

	query = query.Offset(offset).Limit(limit).Order("add_time desc")
	err = query.Find(&accountBalanceFlows).Error
	if err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "query account balance flow failed"
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = gin.H{
		"results": accountBalanceFlows,
		"pagination": gin.H{
			"page":     pn,
			"limit":    limit,
			"total":    total,
			"has_more": offset+limit < int(total),
		},
	}

	c.JSON(http.StatusOK, res)
}

func BalanceWithdrawRequest(c *gin.Context) {
	var req WithdrawReq
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

	mainIdStr, ok := c.Get("main_id")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	var requestLockKey = fmt.Sprintf("withdraw-lock-%s", mainIdStr)
	_, loaded := userSeasonLock.LoadOrStore(requestLockKey, struct{}{})
	defer func() {
		userSeasonLock.Delete(requestLockKey)
	}()
	if loaded {
		res.Code = codes.CODE_ERR_PROCESSING
		res.Msg = "operating, please do not repeat the operation"
		c.JSON(http.StatusOK, res)
		return
	}

	// init system parameters
	chainID := req.ChainID
	contractAddr := os.Getenv("TOPUP_CONTRACT")
	privKeyHex := os.Getenv("WITHDRAW_LOCK_PK")
	var computeAmount = req.Amount.Mul(decimal.NewFromInt(1000000)).Round(0).BigInt().Uint64()

	db := system.GetDb()
	var userMain model.UserMain
	err := db.Model(&model.UserMain{}).Where("id = ?", mainIdStr).First(&userMain).Error
	if err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "query user main failed"
		c.JSON(http.StatusOK, res)
		return
	}

	var existLockFlow model.AccountBalanceFlow
	db.Model(&model.AccountBalanceFlow{}).
		Where("main_id = ? and op = ? AND status IN ?",
			userMain.ID, model.BalanceFlowOpFreeze,
			[]int{model.BalanceFlowStatusPending, model.BalanceFlowStatusPendingWithDraw}).
		First(&existLockFlow)

	if existLockFlow.ID > 0 {
		var _ = big.NewInt(int64(existLockFlow.RealAmount))
		var expiry = uint64(existLockFlow.LockExpiry.Unix())
		// var nonce = new(big.Int).SetBytes(existLockFlow.LockNonce)
		var nonce = hex32ToBigInt(existLockFlow.LockNonce)
		if computeAmount != existLockFlow.RealAmount {
			res.Code = codes.CODE_ERR_BAD_PARAMS
			res.Msg = "lock amount is not equal to the request amount"
			c.JSON(http.StatusOK, res)
			return
		}
		if existLockFlow.LockExpiry.Before(time.Now()) {
			res.Code = codes.CODE_ERR_REQ_EXPIRED
			res.Msg = "lock expiry is before the current time"
			c.JSON(http.StatusOK, res)
			return
		}
		var opType = "lock"
		if existLockFlow.Status == model.BalanceFlowStatusPendingWithDraw {
			opType = "withdraw"
		}
		res.Data = gin.H{
			"operation_id":   existLockFlow.ID,
			"operation_type": opType,
			"lock_id":        existLockFlow.LockID,
			"digest":         "",
			"signature":      existLockFlow.LockSig,
			"expiry":         expiry,
			"nonce":          nonce.String(),
			"amount":         existLockFlow.RealAmount,
		}
		c.JSON(http.StatusOK, res)
		return
	}

	tx := db.Begin()
	committed := false
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("panic", r)
			return
		}
		if !committed {
			tx.Rollback()
		}
	}()

	var userAccount model.AccountBalance
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("main_id = ? and asset_id = ?", userMain.ID, 0).First(&userAccount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
			res.Msg = "user account not found"
			c.JSON(http.StatusOK, res)
			return
		} else {
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "query user account balance failed"
			c.JSON(http.StatusOK, res)
			return
		}
	}
	if userAccount.Available < computeAmount {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "insufficient balance"
		c.JSON(http.StatusOK, res)
		return
	}

	userAccount.Available -= computeAmount
	userAccount.Withdrawal += computeAmount
	userAccount.UpdateTime = time.Now()

	if err := tx.Save(&userAccount).Error; err != nil {
		log.Error("update user account balance failed", err)
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "update and lock user account balance failed"
		c.JSON(http.StatusOK, res)
		return
	}

	// 生成复合合约签名（EIP712）
	// 构造必要参数

	var up model.UserProvider
	db.Model(&model.UserProvider{}).Where("main_id = ? and provider_type = ?", userMain.ID, "wallet").First(&up)
	userAddr := up.ProviderID
	if len(contractAddr) == 0 || len(privKeyHex) == 0 {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "sign config missing"
		c.JSON(http.StatusOK, res)
		return
	}
	if len(userAddr) == 0 {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "user address missing"
		c.JSON(http.StatusOK, res)
		return
	}

	// 生成 lockId
	var timeNow = time.Now()
	expiry := uint64(timeNow.Add(expiryDuration).Unix())

	lockIdHexStr := hexutil.Encode(crypto.Keccak256([]byte(fmt.Sprintf("%d-%d", userMain.ID, timeNow.UnixNano()))))
	log.Info(lockIdHexStr)
	amountBI := new(big.Int).SetUint64(computeAmount)

	// nonce := big.NewInt(timeNow.Unix())
	nonce := generateInternalNonce()

	_, lockIdHex, err := tools.NewLockID32()
	if err != nil {
		log.Error("new lock id failed", err)
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "new lock id failed"
		c.JSON(http.StatusOK, res)
		return
	}

	ctx := context.Background()
	digest, sig, err := chain.BuildAndSignLockAuth(
		ctx,
		chainID,
		contractAddr,
		privKeyHex,
		up.ProviderID,
		lockIdHex,
		amountBI,
		expiry,
		nonce,
	)
	if err != nil {
		log.Error("build and sign lock auth failed", err)
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "build and sign lock auth failed"
		c.JSON(http.StatusOK, res)
		return
	}

	expiryTime := time.Unix(int64(expiry), 0)
	userAccountFlow := model.AccountBalanceFlow{
		MainID:     userMain.ID,
		AssetID:    0,
		Op:         model.BalanceFlowOpFreeze,
		Status:     model.BalanceFlowStatusPending,
		Amount:     computeAmount,
		RealAmount: computeAmount,
		AddTime:    timeNow,
		UpdateTime: timeNow,
		LockID:     lockIdHex,
		LockSig:    hexutil.Encode(sig),
		LockExpiry: &expiryTime,
		LockNonce:  bigIntToHex32(nonce),
		LockAddr:   userAddr,
	}

	if err := tx.Create(&userAccountFlow).Error; err != nil {
		log.Error("create user account flow failed", err)
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "create user account flow failed"
		c.JSON(http.StatusOK, res)
		return
	}

	if err := tx.Commit().Error; err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "commit transaction failed"
		c.JSON(http.StatusOK, res)
		return
	}

	committed = true

	res.Data = gin.H{
		"operation_id":   userAccountFlow.ID,
		"operation_type": "lock",
		"lock_id":        lockIdHex,
		"digest":         hexutil.Encode(digest[:]),
		"signature":      hexutil.Encode(sig),
		"expiry":         expiry,
		"nonce":          nonce.String(),
		"amount":         computeAmount,
	}

	c.JSON(http.StatusOK, res)
}

func BalanceWithdrawCheck(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	mainIdStr, ok := c.Get("main_id")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	operationIdStr := c.Query("operation_id")
	operationId, err := strconv.ParseUint(operationIdStr, 10, 64)
	if err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "invalid operation_id"
		c.JSON(http.StatusOK, res)
		return
	}

	var db = system.GetDb()
	var accountBalanceFlow model.AccountBalanceFlow
	err = db.Model(&model.AccountBalanceFlow{}).Where("main_id = ? and id = ?", mainIdStr, operationId).First(&accountBalanceFlow).Error
	if err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "query account balance flow failed"
		c.JSON(http.StatusOK, res)
		return
	}
	var opType string
	var status string
	switch accountBalanceFlow.Op {
	case model.BalanceFlowOpRecharge:
		opType = "topup"
	case model.BalanceFlowOpFreeze:
		opType = "lock"
	case model.BalanceFlowOpWithdraw:
		opType = "withdraw"
	case model.BalanceFlowOpUnfreeze:
		opType = "unlock"
	default:
		opType = "unknown"
	}
	switch accountBalanceFlow.Status {
	case model.BalanceFlowStatusPending:
		status = "pending"
	case model.BalanceFlowStatusSuccess:
		status = "success"
	case model.BalanceFlowStatusFailed:
		status = "failed"
	case model.BalanceFlowStatusPendingWithDraw:
		status = "pendingWithDraw"
	case model.BalanceFlowStatusCanceled:
		status = "canceled"
	}

	res.Data = gin.H{
		"operation_id":   accountBalanceFlow.ID,
		"operation_type": opType,
		"status":         status,
		"tx_hash":        accountBalanceFlow.TxHash,
		"lock_id":        accountBalanceFlow.LockID,
		"digest":         "",
		"signature":      accountBalanceFlow.LockSig,
		"expiry":         accountBalanceFlow.LockExpiry,
		"nonce":          accountBalanceFlow.LockNonce,
		"amount":         accountBalanceFlow.RealAmount,
	}

	c.JSON(http.StatusOK, res)
}

func generateInternalNonce() *big.Int {
	ts := time.Now().UnixMilli()
	rnd := rand.Int63n(1e8)
	combined := fmt.Sprintf("%d-%d", ts, rnd)
	return new(big.Int).SetBytes(crypto.Keccak256([]byte(combined)))
	// return hexutil.Encode(crypto.Keccak256([]byte(combined)))
}

func bigIntToHex32(n *big.Int) string {
	b := n.Bytes()
	if len(b) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(b):], b)
		b = padded
	}
	return hexutil.Encode(b)
}

func hex32ToBigInt(s string) *big.Int {
	b, err := hexutil.Decode(s)
	if err != nil {
		return nil
	}
	return new(big.Int).SetBytes(b)
}
