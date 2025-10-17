package oauth

import (
	"chaos/api/model"
	"chaos/api/system"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ====== Secrets（建议来自环境变量/配置中心）======
const jwtSecretKey = "replace-with-secure-secret"
const cookieSecretKey = "replace-with-strong-cookie-secret"

var jwtSecret = []byte(jwtSecretKey)

// ====== Client 注册表（演示；生产改为DB/配置）======
type Client struct {
	ID           string
	RedirectURIs []string // 完整匹配；如需前缀匹配要非常谨慎
}

// ====== 提供给外部 router 使用的会话中间件 ======
// 在 router.Init() 中调用：r.Use(oauth.Sessions(true)) // true 表示 HTTPS 场景
func Sessions(secure bool) gin.HandlerFunc {
	store := cookie.NewStore([]byte(cookieSecretKey))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   secure,               // 生产必须 true（全站 HTTPS）
		SameSite: http.SameSiteLaxMode, // 如需要跨站携带，考虑 SameSite=None + Secure
	})
	return sessions.Sessions("sid", store)
}

// ====== Demo 登录页/登录逻辑（可换成你自己的页面与校验）======
func LoginPage(c *gin.Context) {
	templatePaths := []string{
		"templates/login.html",
		"../templates/login.html",
		"./templates/login.html",
		"../../templates/login.html",
	}

	var tmpl *template.Template
	var err error

	for _, path := range templatePaths {
		tmpl, err = template.ParseFiles(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		// 如果模板文件不存在，使用内联模板作为后备
		tmpl, _ = template.New("login").Parse(fallbackTpl)
	}

	err = tmpl.Execute(c.Writer, map[string]any{
		"ReturnURL": c.Query("return_url"),
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
	}
}

func LoginPost(c *gin.Context) {
	email := strings.TrimSpace(strings.ToLower(c.PostForm("username")))
	password := c.PostForm("password")
	returnURL := c.PostForm("return_url")

	// 添加调试日志
	fmt.Printf("Login attempt - Email: %s, ReturnURL: %s\n", email, returnURL)

	// 构建登录页面URL，包含错误参数
	loginURL := "/oauth/login"
	if returnURL != "" {
		loginURL += "?return_url=" + url.QueryEscape(returnURL)
	}

	if email == "" || password == "" {
		fmt.Printf("Login failed - Empty email or password\n")
		errorMsg := "Username and password are required"
		if returnURL != "" {
			loginURL += "&error=" + url.QueryEscape(errorMsg)
		} else {
			loginURL += "?error=" + url.QueryEscape(errorMsg)
		}
		c.Redirect(http.StatusFound, loginURL)
		return
	}

	// 查询用户（只取必要字段）
	db := system.GetDb()
	var user model.UserMain
	if err := db.
		Select("id, email, password"). // 假设列名为 password_hash
		Where("email = ?", email).
		First(&user).Error; err != nil {
		// 统一失败提示
		fmt.Printf("Login failed - User not found: %v\n", err)
		errorMsg := "Invalid username or password"
		if returnURL != "" {
			loginURL += "&error=" + url.QueryEscape(errorMsg)
		} else {
			loginURL += "?error=" + url.QueryEscape(errorMsg)
		}
		c.Redirect(http.StatusFound, loginURL)
		return
	}

	if user.Password != password {
		fmt.Printf("Login failed - Password mismatch for user: %s\n", email)
		errorMsg := "Invalid username or password"
		if returnURL != "" {
			loginURL += "&error=" + url.QueryEscape(errorMsg)
		} else {
			loginURL += "?error=" + url.QueryEscape(errorMsg)
		}
		c.Redirect(http.StatusFound, loginURL)
		return
	}

	// 会话：防止会话固定攻击，先清再写
	sess := sessions.Default(c)
	sess.Clear()
	// 存稳定ID，别存邮箱；如需还可存角色/权限快照
	sess.Set("uid", fmt.Sprintf("%d", user.ID))
	// 可选：存最近登录时间/IP等（注意隐私）
	if err := sess.Save(); err != nil {
		fmt.Printf("Login failed - Session save error: %v\n", err)
		errorMsg := "Session save failed, please try again"
		if returnURL != "" {
			loginURL += "&error=" + url.QueryEscape(errorMsg)
		} else {
			loginURL += "?error=" + url.QueryEscape(errorMsg)
		}
		c.Redirect(http.StatusFound, loginURL)
		return
	}

	if returnURL == "" {
		returnURL = "/"
	}

	fmt.Printf("Login successful - Redirecting to: %s\n", returnURL)
	c.Redirect(http.StatusFound, returnURL)
}

func Logout(c *gin.Context) {
	s := sessions.Default(c)
	s.Clear()
	_ = s.Save()
	c.String(http.StatusOK, "logged out")
}

// ====== 授权码存储（演示：内存；生产换 Redis/DB 并设置 TTL）======
type codeData struct {
	UserID        string
	ClientID      string
	RedirectURI   string
	CodeChallenge string
	ExpiresAt     time.Time
}

var codeStore = struct {
	mu sync.Mutex
	m  map[string]codeData
}{m: make(map[string]codeData)}

func putCode(code string, data codeData) {
	codeStore.mu.Lock()
	defer codeStore.mu.Unlock()
	codeStore.m[code] = data
}
func takeCode(code string) (codeData, bool) {
	codeStore.mu.Lock()
	defer codeStore.mu.Unlock()
	data, ok := codeStore.m[code]
	if ok {
		delete(codeStore.m, code) // 一次性
	}
	return data, ok
}
func genCode(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ====== Session 获取当前用户 ======
func currentUserID(c *gin.Context) (string, bool) {
	sess := sessions.Default(c)
	uid, _ := sess.Get("uid").(string)
	if uid == "" {
		return "", false
	}
	return uid, true
}

// ====== /oauth/authorize ======
// GET /oauth/authorize?response_type=code&client_id=xxx&redirect_uri=xxx&state=yyy&code_challenge=...&code_challenge_method=S256
func AuthorizeHandler(c *gin.Context) {
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	responseType := c.Query("response_type")
	state := c.Query("state")
	codeChallenge := c.Query("code_challenge")
	codeChallengeMethod := c.Query("code_challenge_method")

	// 参数校验
	if clientID == "" || redirectURI == "" || responseType != "code" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	db := system.GetDb()
	var appDev model.GameApp
	db.Where("client_id = ?", clientID).First(&appDev)
	if appDev.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown_client"})
		return
	}
	if !strings.HasPrefix(redirectURI, appDev.OauthCallback) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_redirect_uri"})
		return
	}

	// client & redirect_uri 校验
	// cl, ok := registeredClients[clientID]
	// if !ok {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "unknown_client"})
	// 	return
	// }
	// if !isAllowedRedirectURI(cl, redirectURI) {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_redirect_uri"})
	// 	return
	// }

	// PKCE（如 method=S256 必须带 challenge）
	if codeChallengeMethod == "S256" && codeChallenge == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_pkce_params"})
		return
	}

	// 检查登录会话
	userID, ok := currentUserID(c)
	if !ok {
		loginURL := "/oauth/login?return_url=" + url.QueryEscape(c.Request.RequestURI)
		c.Redirect(http.StatusFound, loginURL)
		return
	}

	// TODO: 可选——展示同意页/Scope

	// 颁发一次性 code
	code, err := genCode(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}
	var pkce string
	if codeChallengeMethod == "S256" && codeChallenge != "" {
		pkce = codeChallenge
	}
	putCode(code, codeData{
		UserID:        userID,
		ClientID:      clientID,
		RedirectURI:   redirectURI,
		CodeChallenge: pkce,
		ExpiresAt:     time.Now().Add(5 * time.Minute),
	})

	// 回跳拼参
	cb, err := url.Parse(redirectURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": err.Error(),
		})
		return
	}
	q := cb.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	cb.RawQuery = q.Encode()

	c.Redirect(http.StatusFound, cb.String())
}

// ====== /oauth/token ======
// POST: grant_type=authorization_code&code=...&redirect_uri=...&code_verifier=...
func TokenHandler(c *gin.Context) {
	type tokenReq struct {
		GrantType    string `form:"grant_type" json:"grant_type"`
		Code         string `form:"code" json:"code"`
		RedirectURI  string `form:"redirect_uri" json:"redirect_uri"`
		CodeVerifier string `form:"code_verifier" json:"code_verifier"`
	}
	var req tokenReq
	if err := c.ShouldBind(&req); err != nil ||
		req.GrantType != "authorization_code" ||
		req.Code == "" || req.RedirectURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}

	data, ok := takeCode(req.Code)
	if !ok || time.Now().After(data.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_or_expired_code"})
		return
	}
	if data.RedirectURI != req.RedirectURI {
		c.JSON(http.StatusBadRequest, gin.H{"error": "redirect_uri_mismatch"})
		return
	}

	// PKCE 校验
	if data.CodeChallenge != "" {
		if req.CodeVerifier == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "pkce_required"})
			return
		}
		h := sha256.Sum256([]byte(req.CodeVerifier))
		if base64.RawURLEncoding.EncodeToString(h[:]) != data.CodeChallenge {
			c.JSON(http.StatusBadRequest, gin.H{"error": "pkce_verification_failed"})
			return
		}
	}

	const duration = 10 * time.Hour
	// 签发 Access Token（JWT）
	claims := jwt.MapClaims{
		"sub": data.UserID,
		"aud": data.ClientID,
		"exp": time.Now().Add(duration).Unix(),
		"iat": time.Now().Unix(),
	}
	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tk.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": signed,
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}
