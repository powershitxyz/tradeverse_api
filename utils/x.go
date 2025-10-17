package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/dghubble/oauth1"
)

// ====== Config & Models ======

type Config struct {
	ConsumerKey    string // X_CONSUMER_API_KEY
	ConsumerSecret string // X_CONSUMER_API_SECRET
	CallbackURL    string // e.g. https://nnnnn.fun/thirdpart/x/callback

	// 可选：应用级 Bearer（App-only auth，用于部分 v2 只读接口）
	BearerToken string // X_BEAR_TOKEN
}

type Tokens struct {
	AccessToken  string // X_ACCESS_TOKEN
	AccessSecret string // X_ACCESS_SECRET
}

type UserV1 struct {
	ID                   int64  `json:"id"`
	IDStr                string `json:"id_str"`
	Name                 string `json:"name"`
	ScreenName           string `json:"screen_name"`
	Description          string `json:"description"`
	ProfileImageURL      string `json:"profile_image_url"`
	ProfileImageURLHttps string `json:"profile_image_url_https"`
	Location             string `json:"location"`
	URL                  string `json:"url"`
}

// ====== 临时 token 存储，用于回调换取 access token ======

type TokenStore interface {
	SaveRequestToken(oauthToken, oauthTokenSecret string, ttl time.Duration) error
	GetRequestSecret(oauthToken string) (string, error)
	DeleteRequestToken(oauthToken string) error
}

type memoryStore struct {
	m sync.Map // oauthToken -> {secret, expireAt}
}

type memVal struct {
	Secret   string
	ExpireAt time.Time
}

func NewMemoryStore() TokenStore { return &memoryStore{} }

func (s *memoryStore) SaveRequestToken(tok, sec string, ttl time.Duration) error {
	s.m.Store(tok, memVal{Secret: sec, ExpireAt: time.Now().Add(ttl)})
	return nil
}
func (s *memoryStore) GetRequestSecret(tok string) (string, error) {
	v, ok := s.m.Load(tok)
	if !ok {
		return "", errors.New("request token not found")
	}
	mv := v.(memVal)
	if time.Now().After(mv.ExpireAt) {
		s.m.Delete(tok)
		return "", errors.New("request token expired")
	}
	return mv.Secret, nil
}
func (s *memoryStore) DeleteRequestToken(tok string) error {
	s.m.Delete(tok)
	return nil
}

// ====== XClient ======

type Client struct {
	cfg        Config
	oauthCfg   *oauth1.Config
	httpClient *http.Client
	store      TokenStore
}

func New(cfg Config, store TokenStore) *Client {
	if store == nil {
		store = NewMemoryStore()
	}
	oCfg := oauth1.Config{
		ConsumerKey:    cfg.ConsumerKey,
		ConsumerSecret: cfg.ConsumerSecret,
		CallbackURL:    cfg.CallbackURL,
		Endpoint: oauth1.Endpoint{
			RequestTokenURL: "https://api.twitter.com/oauth/request_token",
			AuthorizeURL:    "https://api.twitter.com/oauth/authenticate",
			AccessTokenURL:  "https://api.twitter.com/oauth/access_token",
		},
	}
	return &Client{
		cfg:        cfg,
		oauthCfg:   &oCfg,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		store:      store,
	}
}

// Step 1: 发起登录，返回授权 URL（前端重定向到这个 URL）
func (c *Client) StartLogin(ctx context.Context) (authURL, requestToken, requestSecret string, err error) {
	reqToken, reqSecret, err := c.oauthCfg.RequestToken()
	if err != nil {
		return "", "", "", err
	}
	// 保存临时 token -> secret（5 分钟）
	_ = c.store.SaveRequestToken(reqToken, reqSecret, 5*time.Minute)

	u, err := c.oauthCfg.AuthorizationURL(reqToken)
	if err != nil {
		return "", "", "", err
	}
	authURL = u.String()
	return authURL, reqToken, reqSecret, nil
}

// Step 2: 回调处理（/thirdpart/x/callback）
// 根据 oauth_token + oauth_verifier 换取用户 AccessToken/Secret，并读取基本资料
func (c *Client) HandleCallback(ctx context.Context, oauthToken, oauthVerifier string) (tok Tokens, user *UserV1, err error) {
	// 取回 request secret
	reqSecret, err := c.store.GetRequestSecret(oauthToken)
	if err != nil {
		return tok, nil, err
	}
	// 换取 access token
	accessToken, accessSecret, err := c.oauthCfg.AccessToken(oauthToken, reqSecret, oauthVerifier)
	if err != nil {
		return tok, nil, err
	}
	// 注意：不在此处立即删除 request token，避免浏览器并发/重复回调导致找不到。
	// 交由内存 TTL 自动过期即可；如需手动清理，可在业务层成功绑定后清理。

	tok = Tokens{AccessToken: accessToken, AccessSecret: accessSecret}

	// 拉取用户资料（v1.1 verify_credentials）
	u, err := c.VerifyCredentials(ctx, tok, true, true)
	if err != nil {
		return tok, nil, err
	}
	return tok, u, nil
}

// 备用：直接使用已知的 request secret 进行交换（内存丢失兜底）
func (c *Client) ExchangeWithSecret(ctx context.Context, oauthToken, oauthTokenSecret, oauthVerifier string) (tok Tokens, user *UserV1, err error) {
	accessToken, accessSecret, err := c.oauthCfg.AccessToken(oauthToken, oauthTokenSecret, oauthVerifier)
	if err != nil {
		return tok, nil, err
	}
	tok = Tokens{AccessToken: accessToken, AccessSecret: accessSecret}
	u, err := c.VerifyCredentials(ctx, tok, true, true)
	if err != nil {
		return tok, nil, err
	}
	return tok, u, nil
}

// 使用用户 token 请求 v1.1: account/verify_credentials
func (c *Client) VerifyCredentials(ctx context.Context, tok Tokens, includeEmail, skipStatus bool) (*UserV1, error) {
	params := url.Values{}
	if includeEmail {
		params.Set("include_email", "true")
	}
	if skipStatus {
		params.Set("skip_status", "true")
	}
	endpoint := "https://api.twitter.com/1.1/account/verify_credentials.json?" + params.Encode()

	httpClient := c.userHTTPClient(tok)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.New("verify_credentials failed: " + resp.Status + " " + string(body))
	}
	var u UserV1
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

// 示例：发一条推文（v1.1）
func (c *Client) PostTweet(ctx context.Context, tok Tokens, status string) error {
	endpoint := "https://api.twitter.com/1.1/statuses/update.json"
	form := url.Values{}
	form.Set("status", status)

	httpClient := c.userHTTPClient(tok)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBufferString(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return errors.New("post tweet failed: " + resp.Status + " " + string(body))
	}
	return nil
}

// （可选）App-only Bearer 调用 v2 公开只读接口
func (c *Client) appHTTPClient() *http.Client {
	return c.httpClient
}
func (c *Client) doAppBearerGET(ctx context.Context, urlStr string) (*http.Response, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	req.Header.Set("Authorization", "Bearer "+c.cfg.BearerToken)
	return c.appHTTPClient().Do(req)
}

// 生成带用户签名的 http.Client
func (c *Client) userHTTPClient(tok Tokens) *http.Client {
	token := oauth1.NewToken(tok.AccessToken, tok.AccessSecret)
	return c.oauthCfg.Client(oauth1.NoContext, token)
}
