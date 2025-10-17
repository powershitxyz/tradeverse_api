package oauth

import (
	"chaos/api/api/common"
	"chaos/api/codes"
	"chaos/api/log"
	"chaos/api/model"
	"chaos/api/system"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 判断是否唯一键冲突（MySQL 1062）
func isDup(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1062
	}
	return false
}

// 如果你想在 router 里拿到 secret，可用这个 accessor
func GetJWTSecret() []byte { return jwtSecret }

// Gin 中间件：校验 Bearer JWT，验证签名&过期时间，把 sub/aud 放进 context
func AuthzMiddleware(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		ah := c.GetHeader("Authorization")
		if !strings.HasPrefix(ah, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_bearer"})
			return
		}
		tokenStr := strings.TrimSpace(strings.TrimPrefix(ah, "Bearer "))

		tok, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// 只接受 HS256
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, jwt.ErrTokenUnverifiable
			}
			return secret, nil
		})
		if err != nil || !tok.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			return
		}

		claims, ok := tok.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_claims"})
			return
		}

		// 过期校验（稳妥起见手动检查 exp）
		if exp, ok := claims["exp"].(float64); ok && int64(exp) < time.Now().Unix() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token_expired"})
			return
		}

		// 将关键信息放入 context
		if sub, _ := claims["sub"].(string); sub != "" {
			c.Set("sub", sub)
		}
		if aud, _ := claims["aud"].(string); aud != "" {
			c.Set("aud", aud)
		}

		c.Next()
	}
}

func AdvancedAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = c.GetString("aud") // 应用ID
		_ = c.GetString("sub") // 用户ID

		// 暂时禁止所有高级接口调用
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code":      403,
			"msg":       "Advanced APIs are temporarily disabled",
			"error":     "insufficient_permissions",
			"timestamp": time.Now().Unix(),
		})

		// 将来的扩展逻辑可以放在这里
		// 例如：
		// if !hasAdvancedPermission(clientId, userId) {
		//     c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient_permissions"})
		//     return
		// }

		// c.Next() // 权限验证通过时调用
	}
}

func MeHandler(c *gin.Context) {
	_ = c.GetString("aud") //appid
	userId := c.GetString("sub")

	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	var userMain model.UserMain
	var userProfile model.UserProfile
	db := system.GetDb()
	db.Model(&model.UserMain{}).Where("id = ?", userId).First(&userMain)

	if userMain.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user not found"
		c.JSON(http.StatusOK, res)
		return
	}

	db.Model(&model.UserProfile{}).Where("main_id = ?", userId).First(&userProfile)
	var userAccount model.AccountBalance
	db.Model(&model.AccountBalance{}).Where("main_id = ? and asset_id = ?", userId, 0).First(&userAccount)

	res.Data = gin.H{
		"user_no":      userMain.UserNo,
		"user_name":    userProfile.Name,
		"avatar":       userProfile.Avatar,
		"bio":          userProfile.Bio,
		"birthday":     userProfile.Birthday,
		"country_code": userProfile.CountryCode,
		"timezone":     userProfile.Timezone,
		"balance":      userAccount.Available,
	}
	c.JSON(http.StatusOK, res)
}

func GameSessionInitHandler(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	clientId := c.GetString("aud") //appid
	userId := c.GetString("sub")

	var userMain model.UserMain
	db := system.GetDb()
	db.Model(&model.UserMain{}).Where("id = ?", userId).First(&userMain)

	if userMain.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameApp model.GameApp
	db.Model(&model.GameApp{}).Where("client_id = ?", clientId).First(&gameApp)

	var gameInfo model.GameInfo
	db.Table("n_game_info gi").Joins("LEFT JOIN n_game_app ga ON gi.id = ga.game_id").
		Where("ga.client_id = ?", clientId).First(&gameInfo)

	if gameInfo.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game info not found"
		c.JSON(http.StatusOK, res)
		return
	}
	if gameInfo.Status == model.GameStatusDraft || gameInfo.Status == model.GameStatusInactive {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "game not active or testing"
		c.JSON(http.StatusOK, res)
		return
	}

	log.Infof("game session init handler: gameId: %d, clientId: %s, userId: %s, userMainId: %d", gameApp.GameID, clientId, userId, userMain.ID)

	if gameApp.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game app not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var existGameSession model.GameSession
	err := db.Model(&model.GameSession{}).Where("main_id = ? AND game_id = ? and status = ?", userMain.ID, gameApp.GameID, model.GameSessionStatusStart).First(&existGameSession).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "query game session failed"
		c.JSON(http.StatusOK, res)
		return
	} else if err == nil {
		res.Data = gin.H{
			"session_id": existGameSession.SessionID,
		}
		c.JSON(http.StatusOK, res)
		return
	}

	var testing = 0
	if gameInfo.Status != model.GameStatusActive {
		testing = 1
	}
	sessionId := uuid.New().String()
	gameSession := model.GameSession{
		SessionID: sessionId,
		MainID:    userMain.ID,
		GameID:    gameApp.GameID,
		Status:    model.GameSessionStatusStart,
		StartTime: time.Now(),
		Testing:   testing,
	}

	if err := db.Create(&gameSession).Error; err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "create session failed"
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = gin.H{
		"session_id": sessionId,
	}
	c.JSON(http.StatusOK, res)
}

func GameStartHandler(c *gin.Context) {
	var req GameStartReq

	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "invalid param"
		c.JSON(http.StatusOK, res)
		return
	}
	if len(req.ExternalID) > 128 {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "external_id too long"
		c.JSON(http.StatusOK, res)
		return
	}
	if len(req.Remark) > 128 {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "remark too long"
		c.JSON(http.StatusOK, res)
		return
	}

	clientId := c.GetString("aud") //appid
	userId := c.GetString("sub")

	var userMain model.UserMain
	db := system.GetDb()
	db.Model(&model.UserMain{}).Where("id = ?", userId).First(&userMain)

	if userMain.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameSession model.GameSession
	err := db.Model(&model.GameSession{}).Where("session_id = ?", req.SessionID).First(&gameSession).Error
	if err != nil {
		log.Errorf("query game session failed: %s, err: %s", req.SessionID, err.Error())
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game session not found"
		c.JSON(http.StatusOK, res)
		return
	}
	if gameSession.Status != model.GameSessionStatusStart {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "game session not started"
		c.JSON(http.StatusOK, res)
		return
	}
	if gameSession.MainID != userMain.ID {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "session not owned by user"
		c.JSON(http.StatusOK, res)
		return
	}

	if len(req.PlaySettingCode) == 0 {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "play setting code is required"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameSetting model.GameSetting
	db.Model(&model.GameSetting{}).Where("game_id = ? and code = ?", gameSession.GameID, req.PlaySettingCode).First(&gameSetting)

	if gameSetting.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game play setting not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var gameInfo model.GameInfo
	db.Model(&model.GameInfo{}).Where("id = ?", gameSession.GameID).First(&gameInfo)
	if gameInfo.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game info not found"
		c.JSON(http.StatusOK, res)
		return
	}
	if gameInfo.Status == model.GameStatusDraft || gameInfo.Status == model.GameStatusInactive {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "game not active or testing"
		c.JSON(http.StatusOK, res)
		return
	}

	if gameInfo.Status != model.GameStatusActive && gameSetting.AmountPerPlay > 0 {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "game is testing, but play setting amount per play is not 0"
		c.JSON(http.StatusOK, res)
		return
	}

	// if gameSetting.AmountPerPlay == 0 {
	// 	// return success if amount per play is 0
	// 	res.Data = gin.H{
	// 		"operation_id":   "0",
	// 		"operation_type": "start",
	// 	}
	// 	c.JSON(http.StatusOK, res)
	// 	return
	// }

	{
		var existing model.AccountFlow
		err := db.Where(
			"main_id = ? AND session_id = ? AND biz_type = ? AND status = ?",
			userMain.ID, gameSession.SessionID, model.FlowFreeze, model.FlowStatusPending,
		).First(&existing).Error

		if err == nil && existing.ID > 0 {
			res.Data = gin.H{
				"operation_id":   existing.ID,
				"operation_type": "freeze",
			}
			c.JSON(http.StatusOK, res)
			return
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "query running game failed"
			c.JSON(http.StatusOK, res)
			return
		}
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("panic", r)
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "internal error"
			c.JSON(http.StatusOK, res)
			return
		}
	}()

	// 带行锁读取余额，避免并发扣错
	var userAccount model.AccountBalance
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("main_id = ? AND asset_id = ?", userMain.ID, 0).
		First(&userAccount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if gameSetting.AmountPerPlay > 0 {
				// 需要扣费但账户不存在
				tx.Rollback()
				res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
				res.Msg = "please top up your balance"
				c.JSON(http.StatusOK, res)
				return
			} else {
				// 免费游戏，账户不存在也允许继续
				// 创建一个余额为0的账户记录
				userAccount = model.AccountBalance{
					MainID:    userMain.ID,
					AssetID:   0,
					Available: 0,
					Frozen:    0,
				}
				if err := tx.Create(&userAccount).Error; err != nil {
					tx.Rollback()
					res.Code = codes.CODE_ERR_UNKNOWN
					res.Msg = "create user account failed"
					c.JSON(http.StatusOK, res)
					return
				}
			}
		} else {
			// 其他数据库错误
			tx.Rollback()
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "query user account balance failed"
			c.JSON(http.StatusOK, res)
			return
		}
	}
	if userAccount.Available < gameSetting.AmountPerPlay {
		tx.Rollback()
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "insufficient balance"
		c.JSON(http.StatusOK, res)
		return
	}
	accountFlow := model.AccountFlow{
		MainID:         userMain.ID,
		AssetID:        uint64(0),
		BizType:        model.FlowFreeze,
		Amount:         gameSetting.AmountPerPlay,
		Direction:      model.DirectionNone,
		ExternalID:     req.ExternalID,
		ExternalRemark: req.Remark,
		Status:         model.FlowStatusPending,
		AddTime:        time.Now(),
		UpdateTime:     time.Now(),
		ClientID:       clientId,
		GameID:         gameSession.GameID,
		SessionID:      gameSession.SessionID,
	}

	if err := tx.Create(&accountFlow).Error; err != nil {
		if isDup(err) {
			// 再查一次，返回已有的那条
			var existing model.AccountFlow
			if e := tx.Where(
				"main_id = ? AND session_id = ? AND biz_type = ? AND status = ?",
				userMain.ID, gameSession.SessionID, model.FlowFreeze, model.FlowStatusPending,
			).First(&existing).Error; e == nil {
				if cerr := tx.Commit().Error; cerr != nil {
					tx.Rollback()
				}
				res.Data = gin.H{"operation_id": existing.ID, "operation_type": "freeze"}
				c.JSON(http.StatusOK, res)
				return
			}
		}
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "freeze operation failed"
		c.JSON(http.StatusOK, res)
		return
	}

	userAccount.Available -= gameSetting.AmountPerPlay
	userAccount.Frozen += gameSetting.AmountPerPlay
	if err := tx.Save(&userAccount).Error; err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "freeze operation failed (balance update)"
		c.JSON(http.StatusOK, res)
		return
	}
	if err := tx.Commit().Error; err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "freeze commit failed"
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = gin.H{
		"operation_id":   accountFlow.ID,
		"operation_type": "freeze",
	}

	c.JSON(http.StatusOK, res)
}

func GameEndHandler(c *gin.Context) {
	var req GameEndReq

	res := common.Response{Timestamp: time.Now().Unix(), Code: codes.CODE_SUCCESS, Msg: "success"}

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "invalid param"
		c.JSON(http.StatusOK, res)
		return
	}

	clientId := c.GetString("aud") // appid（记录用）
	userId := c.GetString("sub")

	db := system.GetDb()

	// 1) 用户校验
	var userMain model.UserMain
	if err := db.Model(&model.UserMain{}).Where("id = ?", userId).First(&userMain).Error; err != nil || userMain.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user not found"
		c.JSON(http.StatusOK, res)
		return
	}

	// 2) 会话校验
	var gameSession model.GameSession
	if err := db.Model(&model.GameSession{}).Where("session_id = ?", req.SessionID).First(&gameSession).Error; err != nil {
		log.Errorf("query game session failed: %s, err: %s", req.SessionID, err.Error())
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game session not found"
		c.JSON(http.StatusOK, res)
		return
	}
	if gameSession.Status != model.GameSessionStatusStart {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "game session not started"
		c.JSON(http.StatusOK, res)
		return
	}
	if gameSession.MainID != userMain.ID {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "session not owned by user"
		c.JSON(http.StatusOK, res)
		return
	}
	// 2.1 校验 session 属于当前 client（防跨应用）
	{
		var gameApp model.GameApp
		if err := db.Model(&model.GameApp{}).Where("client_id = ?", clientId).First(&gameApp).Error; err != nil || gameApp.ID == 0 {
			res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
			res.Msg = "game app not found"
			c.JSON(http.StatusOK, res)
			return
		}
		if gameApp.GameID != gameSession.GameID {
			res.Code = codes.CODE_ERR_BAD_PARAMS
			res.Msg = "session not belong to this client"
			c.JSON(http.StatusOK, res)
			return
		}
	}

	// 3) 查冻结流水（非锁，先拿到基本信息）
	var freezeFlow model.AccountFlow
	if err := db.Model(&model.AccountFlow{}).Where("id = ?", req.GameStartID).First(&freezeFlow).Error; err != nil || freezeFlow.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "game flow not found"
		c.JSON(http.StatusOK, res)
		return
	}
	// 属主/类型/会话绑定校验
	if freezeFlow.MainID != userMain.ID {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "game flow not owned by user"
		c.JSON(http.StatusOK, res)
		return
	}
	if freezeFlow.BizType != model.FlowFreeze {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "invalid flow type"
		c.JSON(http.StatusOK, res)
		return
	}
	if freezeFlow.SessionID != gameSession.SessionID {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "flow not match session"
		c.JSON(http.StatusOK, res)
		return
	}

	// 4) 幂等：如果已经有对应 spend，直接返回（带上 session_id 更稳）
	{
		var spent model.AccountFlow
		if err := db.Where("ref_flow_id = ? AND biz_type = ? AND session_id = ?",
			freezeFlow.ID, model.FlowSpend, gameSession.SessionID).
			First(&spent).Error; err == nil && spent.ID > 0 {
			res.Data = gin.H{
				"operation_id":   spent.ID,
				"operation_type": "spend",
			}
			c.JSON(http.StatusOK, res)
			return
		}
	}

	// 5) 事务：锁冻结单 & 锁余额 → 写 spend → 扣冻结 → 改冻结状态 → 结束会话
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "internal error"
			c.JSON(http.StatusOK, res)
		}
	}()

	// 5.1 冻结单加锁确认仍为 pending（带 session 约束）
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND session_id = ? AND biz_type = ? AND status = ?",
			req.GameStartID, gameSession.SessionID, model.FlowFreeze, model.FlowStatusPending).
		First(&freezeFlow).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Code = codes.CODE_ERR_BAD_PARAMS
			res.Msg = "game flow not pending"
		} else {
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "load flow failed"
		}
		c.JSON(http.StatusOK, res)
		return
	}

	// 5.2 锁余额
	var userAccount model.AccountBalance
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("main_id = ? AND asset_id = ?", userMain.ID, 0).
		First(&userAccount).Error; err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user account not found"
		c.JSON(http.StatusOK, res)
		return
	}
	if userAccount.Frozen < freezeFlow.Amount {
		tx.Rollback()
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "frozen balance insufficient"
		c.JSON(http.StatusOK, res)
		return
	}

	// 5.3 写 spend（方向=Out，引用 freeze）
	spendFlow := model.AccountFlow{
		MainID:     userMain.ID,
		AssetID:    uint64(0),
		BizType:    model.FlowSpend,
		Amount:     freezeFlow.Amount,
		Direction:  model.DirectionOut,
		RefFlowID:  freezeFlow.ID,
		Status:     model.FlowStatusDone,
		AddTime:    time.Now(),
		UpdateTime: time.Now(),
		ClientID:   clientId,
		GameID:     freezeFlow.GameID,
		SessionID:  gameSession.SessionID,
	}
	if err := tx.Create(&spendFlow).Error; err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "create spend failed"
		c.JSON(http.StatusOK, res)
		return
	}

	// 5.4 扣冻结
	userAccount.Frozen -= freezeFlow.Amount
	if err := tx.Save(&userAccount).Error; err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "update balance failed"
		c.JSON(http.StatusOK, res)
		return
	}

	// 5.5 更新冻结单状态
	freezeFlow.Status = model.FlowStatusDone
	freezeFlow.UpdateTime = time.Now()
	if err := tx.Save(&freezeFlow).Error; err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "update freeze failed"
		c.JSON(http.StatusOK, res)
		return
	}

	// 5.6 结束会话
	gameSession.Status = model.GameSessionStatusEnd
	gameSession.EndTime = time.Now()
	gameSession.Score = req.Score
	gameSession.SpendAmountN = spendFlow.Amount
	if err := tx.Save(&gameSession).Error; err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "update game session failed"
		c.JSON(http.StatusOK, res)
		return
	}

	// 5.7 提交（失败时不要再 Rollback）
	if err := tx.Commit().Error; err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "commit failed"
		c.JSON(http.StatusOK, res)
		return
	}

	// 6) 成功返回
	res.Data = gin.H{
		"operation_id":   spendFlow.ID,
		"operation_type": "spend",
	}
	c.JSON(http.StatusOK, res)
}

/***************** Advanced Money Control API *****************/

func FreezeHandler(c *gin.Context) {
	var req FreezeReq

	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "invalid param"
		c.JSON(http.StatusOK, res)
		return
	}
	if req.Amount == 0 || len(req.ExternalID) == 0 {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "invalid param"
		c.JSON(http.StatusOK, res)
		return
	}

	if len(req.ExternalID) > 128 {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "external_id too long"
		c.JSON(http.StatusOK, res)
		return
	}
	if len(req.Remark) > 128 {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "remark too long"
		c.JSON(http.StatusOK, res)
		return
	}

	clientId := c.GetString("aud") //appid
	userId := c.GetString("sub")

	var userMain model.UserMain

	db := system.GetDb()
	db.Model(&model.UserMain{}).Where("id = ?", userId).First(&userMain)

	if userMain.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var userAccount model.AccountBalance
	db.Model(&model.AccountBalance{}).Where("main_id = ? and asset_id = ?", userId, 0).First(&userAccount)

	if userAccount.Available < req.Amount {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "insufficient balance"
		c.JSON(http.StatusOK, res)
		return
	}
	tx := db.Begin()
	accountFlow := model.AccountFlow{
		MainID:         userMain.ID,
		AssetID:        uint64(0),
		BizType:        model.FlowFreeze,
		Amount:         req.Amount,
		Direction:      1,
		ExternalID:     req.ExternalID,
		ExternalRemark: req.Remark,
		Status:         model.FlowStatusPending,
		AddTime:        time.Now(),
		UpdateTime:     time.Now(),
		ClientID:       clientId,
	}
	err := tx.Save(&accountFlow).Error
	if err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "freeze operation failed"
		c.JSON(http.StatusOK, res)
		return
	}
	userAccount.Available -= req.Amount
	userAccount.Frozen += req.Amount
	err = tx.Save(&userAccount).Error
	if err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "freeze operation failed"
		c.JSON(http.StatusOK, res)
		return
	}
	tx.Commit()

	res.Data = gin.H{
		"operation_id":   accountFlow.ID,
		"operation_type": "freeze",
	}

	c.JSON(http.StatusOK, res)
}

func UnFreezeHandler(c *gin.Context) {
	var req SpendOrUnfreezeReq

	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "invalid param"
		c.JSON(http.StatusOK, res)
		return
	}
	if req.FreezeID == 0 {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "invalid param"
		c.JSON(http.StatusOK, res)
		return
	}

	clientId := c.GetString("aud") //appid
	userId := c.GetString("sub")

	var userMain model.UserMain

	db := system.GetDb()
	db.Model(&model.UserMain{}).Where("id = ?", userId).First(&userMain)

	if userMain.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var userAccount model.AccountBalance
	db.Model(&model.AccountBalance{}).Where("main_id = ? and asset_id = ?", userId, 0).First(&userAccount)

	var accountFlow model.AccountFlow
	db.Model(&model.AccountFlow{}).Where("id = ?", req.FreezeID).First(&accountFlow)
	if accountFlow.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "freeze not found"
		c.JSON(http.StatusOK, res)
		return
	}
	if accountFlow.Status != model.FlowStatusPending {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "freeze been reversed"
		c.JSON(http.StatusOK, res)
		return
	}

	tx := db.Begin()

	accountFlow.Status = model.FlowStatusReversed
	accountFlow.UpdateTime = time.Now()
	err := tx.Save(&accountFlow).Error
	if err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "freeze operation failed"
		c.JSON(http.StatusOK, res)
		return
	}

	unfreezeAccountFlow := model.AccountFlow{
		MainID:     userMain.ID,
		AssetID:    uint64(0),
		BizType:    model.FlowUnfreeze,
		Amount:     accountFlow.Amount,
		Direction:  0,
		RefFlowID:  accountFlow.ID,
		Status:     model.FlowStatusDone,
		AddTime:    time.Now(),
		UpdateTime: time.Now(),
		ClientID:   clientId,
	}
	err = tx.Save(&unfreezeAccountFlow).Error
	if err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "unfreeze operation failed"
		c.JSON(http.StatusOK, res)
		return
	}
	userAccount.Frozen -= accountFlow.Amount
	userAccount.Available += accountFlow.Amount
	err = tx.Save(&userAccount).Error
	if err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "unfreeze operation failed"
		c.JSON(http.StatusOK, res)
		return
	}

	tx.Commit()

	res.Data = gin.H{
		"operation_id":   unfreezeAccountFlow.ID,
		"operation_type": "unfreeze",
	}

	c.JSON(http.StatusOK, res)
}

func SpendHandler(c *gin.Context) {
	var req SpendOrUnfreezeReq

	res := common.Response{}
	res.Timestamp = time.Now().Unix()
	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = nil

	if err := c.ShouldBindJSON(&req); err != nil {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "invalid param"
		c.JSON(http.StatusOK, res)
		return
	}
	if req.FreezeID == 0 {
		res.Code = codes.CODE_ERR_BAD_PARAMS
		res.Msg = "invalid param"
		c.JSON(http.StatusOK, res)
		return
	}

	clientId := c.GetString("aud") //appid
	userId := c.GetString("sub")

	var userMain model.UserMain

	db := system.GetDb()
	db.Model(&model.UserMain{}).Where("id = ?", userId).First(&userMain)

	if userMain.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var userAccount model.AccountBalance
	db.Model(&model.AccountBalance{}).Where("main_id = ? and asset_id = ?", userId, 0).First(&userAccount)

	var accountFlow model.AccountFlow
	db.Model(&model.AccountFlow{}).Where("id = ?", req.FreezeID).First(&accountFlow)
	if accountFlow.ID == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "freeze not found"
		c.JSON(http.StatusOK, res)
		return
	}
	if accountFlow.Status != model.FlowStatusPending {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "freeze been reversed"
		c.JSON(http.StatusOK, res)
		return
	}

	tx := db.Begin()

	accountFlow.Status = model.FlowStatusDone
	accountFlow.UpdateTime = time.Now()
	err := tx.Save(&accountFlow).Error
	if err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "spend operation failed"
		c.JSON(http.StatusOK, res)
		return
	}

	spendAccountFlow := model.AccountFlow{
		MainID:     userMain.ID,
		AssetID:    uint64(0),
		BizType:    model.FlowSpend,
		Amount:     accountFlow.Amount,
		Direction:  0,
		RefFlowID:  accountFlow.ID,
		Status:     model.FlowStatusDone,
		AddTime:    time.Now(),
		UpdateTime: time.Now(),
		ClientID:   clientId,
	}
	err = tx.Save(&spendAccountFlow).Error
	if err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "spend operation failed"
		c.JSON(http.StatusOK, res)
		return
	}
	userAccount.Frozen -= accountFlow.Amount
	err = tx.Save(&userAccount).Error
	if err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "spend operation failed"
		c.JSON(http.StatusOK, res)
		return
	}

	tx.Commit()

	res.Data = gin.H{
		"operation_id":   spendAccountFlow.ID,
		"operation_type": "spend",
	}

	c.JSON(http.StatusOK, res)
}
