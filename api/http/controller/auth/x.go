package auth

import (
	"chaos/api/api/common"
	"chaos/api/codes"
	"chaos/api/log"
	"chaos/api/model"
	"chaos/api/security"
	"chaos/api/system"
	"chaos/api/utils"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var xTokenStore = utils.NewMemoryStore()
var xReqMap sync.Map // request_token -> main_id

// validateUserToken 验证用户token并返回用户ID
func validateUserToken(token string) (uint64, error) {
	// 解密token
	decrypted, err := security.Decrypt(token)
	if err != nil {
		return 0, fmt.Errorf("invalid token")
	}

	// 解析token
	tokenArr := strings.Split(decrypted, "|")
	if len(tokenArr) != 4 {
		return 0, fmt.Errorf("invalid token format")
	}

	// 检查过期时间
	expireTs, err := strconv.ParseInt(tokenArr[3], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid token format")
	}

	if time.Now().Unix()-expireTs > int64(common.TOKEN_DURATION.Seconds()) {
		return 0, fmt.Errorf("token expired")
	}

	// 获取用户ID
	userID, err := strconv.ParseUint(tokenArr[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID")
	}

	return userID, nil
}

func newXClient(c *gin.Context) *utils.Client {
	cb := os.Getenv("X_CALLBACK_URL")
	if cb == "" {
		scheme := c.Request.Header.Get("X-Forwarded-Proto")
		if scheme == "" {
			scheme = "http" // 开发环境使用http
		}
		// 使用完整的后端API地址
		cb = fmt.Sprintf("%s://%s/spwapi/preauth/thirdpart/x/callback", scheme, c.Request.Host)
	}
	cfg := utils.Config{
		ConsumerKey:    os.Getenv("X_CONSUMER_API_KEY"),
		ConsumerSecret: os.Getenv("X_CONSUMER_API_SECRET"),
		CallbackURL:    cb,
		BearerToken:    os.Getenv("X_BEAR_TOKEN"),
	}
	return utils.New(cfg, xTokenStore)
}

// GET /preauth/thirdpart/x/login
func XLogin(c *gin.Context) {
	// 从URL参数获取用户token
	userToken := c.Query("token")
	if userToken == "" {
		// 重定向到前端页面并显示错误
		redirectURL := "/#/profile?error=Missing+user+token"
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	// 验证用户token并获取用户ID
	mainID, err := validateUserToken(userToken)
	if err != nil {
		// 重定向到前端页面并显示错误
		redirectURL := fmt.Sprintf("/#/profile?error=%s", template.URLQueryEscaper(err.Error()))
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	cli := newXClient(c)
	authURL, reqToken, reqSecret, err := cli.StartLogin(c)
	if err != nil {
		// 重定向到前端页面并显示错误
		redirectURL := fmt.Sprintf("/#/profile?error=%s", template.URLQueryEscaper(err.Error()))
		c.Redirect(http.StatusFound, redirectURL)
		return
	}
	// 备份 request_secret 到短期 Cookie（10 分钟）
	cookieVal := base64.StdEncoding.EncodeToString([]byte(reqSecret))
	http.SetCookie(c.Writer, &http.Cookie{Name: "X_REQ_SEC", Value: cookieVal, Path: "/", HttpOnly: true, MaxAge: 600})
	http.SetCookie(c.Writer, &http.Cookie{Name: "X_REQ_TOK", Value: reqToken, Path: "/", HttpOnly: true, MaxAge: 600})

	// 存储用户ID和request token的映射
	xReqMap.Store(reqToken, fmt.Sprintf("%d", mainID))
	go func(tok string) {
		time.Sleep(10 * time.Minute)
		xReqMap.Delete(tok)
	}(reqToken)

	// 直接重定向到Twitter授权页面
	c.Redirect(http.StatusFound, authURL)
}

// GET /auth/thirdpart/x/callback
func XCallback(c *gin.Context) {
	oauthToken := c.Query("oauth_token")
	oauthVerifier := c.Query("oauth_verifier")

	cli := newXClient(c)

	// 从内存映射中获取用户ID
	v, ok := xReqMap.Load(oauthToken)
	if !ok {
		// 返回JSON错误响应
		res := common.Response{Code: codes.CODE_ERR_UNKNOWN, Msg: "Invalid or expired request token", Timestamp: time.Now().Unix()}
		c.JSON(http.StatusOK, res)
		return
	}
	mainIDStr := fmt.Sprintf("%v", v)
	mainID, err := strconv.ParseUint(mainIDStr, 10, 64)
	if err != nil {
		// 返回JSON错误响应
		res := common.Response{Code: codes.CODE_ERR_UNKNOWN, Msg: "Invalid user ID", Timestamp: time.Now().Unix()}
		c.JSON(http.StatusOK, res)
		return
	}
	log.Info("Processing callback for user ID:", mainID)

	// 优先使用内存；未命中则尝试 Cookie 兜底
	_, user, err := cli.HandleCallback(c, oauthToken, oauthVerifier)
	if err != nil {
		// fallback via cookie
		secCookie, err1 := c.Request.Cookie("X_REQ_SEC")
		tokCookie, err2 := c.Request.Cookie("X_REQ_TOK")
		if err1 == nil && err2 == nil && tokCookie.Value == oauthToken {
			b, _ := base64.StdEncoding.DecodeString(secCookie.Value)
			if t, u, e2 := cli.ExchangeWithSecret(c, oauthToken, string(b), oauthVerifier); e2 == nil {
				_, user, err = t, u, nil
			}
		}
	}
	if err != nil {
		log.Error("x callback error", err)
		// 返回JSON错误响应
		res := common.Response{Code: codes.CODE_ERR_UNKNOWN, Msg: err.Error(), Timestamp: time.Now().Unix()}
		c.JSON(http.StatusOK, res)
		return
	}

	// 检查是否已经绑定
	db := system.GetDb()
	var up model.UserProvider
	db.Model(&model.UserProvider{}).Where("provider_type = ? AND provider_label = ? AND provider_id = ?", "thirdpart", "x", user.IDStr).First(&up)
	if up.ID > 0 {
		// 返回JSON错误响应
		res := common.Response{Code: codes.CODE_ERR_EXIST_OBJ, Msg: "This Twitter account has already been bound to another user", Timestamp: time.Now().Unix()}
		c.JSON(http.StatusOK, res)
		return
	}

	// 创建新的绑定
	up.MainID = mainID
	up.ProviderType = "thirdpart"
	up.ProviderLabel = "x"
	up.ProviderID = user.IDStr
	up.AddTime = time.Now()
	if err := db.Save(&up).Error; err != nil {
		log.Error("Failed to save user provider", err)
	}

	// 把user里面的相关信息全部更新到 profile 表中
	var prof model.UserProfile
	db.Model(&model.UserProfile{}).Where("main_id = ?", mainID).First(&prof)
	avatar := user.ProfileImageURLHttps
	if avatar == "" {
		avatar = user.ProfileImageURL
	}
	avatar = strings.Replace(avatar, "_normal", "", 1)
	if prof.ID == 0 {
		prof = model.UserProfile{MainID: mainID}
	}
	if user.Name != "" {
		prof.Name = user.Name
	}
	if user.Description != "" {
		prof.Bio = user.Description
	}
	if user.ScreenName != "" {
		prof.XUri = "https://x.com/" + user.ScreenName
	}
	if avatar != "" {
		prof.Avatar = avatar
	}
	_ = db.Save(&prof).Error
	xReqMap.Delete(oauthToken)

	// 清理短期 Cookie
	http.SetCookie(c.Writer, &http.Cookie{Name: "X_REQ_SEC", Value: "", Path: "/", HttpOnly: true, MaxAge: -1})
	http.SetCookie(c.Writer, &http.Cookie{Name: "X_REQ_TOK", Value: "", Path: "/", HttpOnly: true, MaxAge: -1})

	// 返回JSON成功响应
	res := common.Response{
		Code:      codes.CODE_SUCCESS,
		Msg:       "Twitter connected successfully",
		Timestamp: time.Now().Unix(),
		Data: gin.H{
			"twitter_id":        user.ID,
			"twitter_id_str":    user.IDStr,
			"name":              user.Name,
			"screen_name":       user.ScreenName,
			"description":       user.Description,
			"profile_image_url": user.ProfileImageURLHttps,
		},
	}
	c.JSON(http.StatusOK, res)
}
