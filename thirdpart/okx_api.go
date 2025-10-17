package thirdpart

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"chaos/api/log"
)

type OKXClient struct {
	apiKey     string
	secretKey  string
	passphrase string
	projectID  string
	baseURL    string
	httpClient *http.Client
	mu         sync.Mutex // 确保并发安全
}

type TokenInfo struct {
	FromToken struct {
		Symbol   string
		Decimals int
		Price    string
	} `json:"fromToken"`
	ToToken struct {
		Symbol   string
		Decimals int
		Price    string
	} `json:"toToken"`
}

type WalletAsset struct {
	TotalValue string `json:"totalValue"`
}

type AssetDetail struct {
	ChainIndex      string `json:"chainIndex"`
	TokenAddress    string `json:"tokenAddress"`
	Symbol          string `json:"symbol"`
	Balance         string `json:"balance"`
	TokenPrice      string `json:"tokenPrice"`
	TokenType       string `json:"tokenType"`
	IsRiskToken     bool   `json:"isRiskToken"`
	TransferAmount  string `json:"transferAmount"`
	AvailableAmount string `json:"availableAmount"`
	RawBalance      string `json:"rawBalance"`
	Address         string `json:"address"`
	PriceChange     string `json:"price_change"`
}

type TokenPriceRequest struct {
	ChainIndex   string `json:"chainIndex"`
	TokenAddress string `json:"tokenAddress"`
}
type TokenPriceResponse struct {
	TokenPriceRequest
	Time  string `json:"time"`
	Price string `json:"price"`
}

var okxClient *OKXClient

func GetOKXClient() (*OKXClient, error) {
	if okxClient != nil {
		return okxClient, nil
	}
	return newOKXClient()
}

func newOKXClient() (*OKXClient, error) {
	client := &OKXClient{
		apiKey:     os.Getenv("OKX_API_KEY"),
		secretKey:  os.Getenv("OKX_SECRET_KEY"),
		passphrase: os.Getenv("OKX_API_PASSPHRASE"),
		projectID:  os.Getenv("OKX_PROJECT_ID"),
		baseURL:    "https://www.okx.com",
		httpClient: &http.Client{Timeout: 10 * time.Second}, // 配置 HTTP 连接池
	}

	if client.apiKey == "" || client.secretKey == "" || client.passphrase == "" || client.projectID == "" {
		return nil, fmt.Errorf("缺少必要的 OKX API 环境变量")
	}

	return client, nil
}

func (c *OKXClient) getHeaders(timestamp, method, requestPath, queryString string) map[string]string {
	stringToSign := timestamp + method + requestPath + queryString

	// HMAC-SHA256 签名
	h := hmac.New(sha256.New, []byte(c.secretKey))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return map[string]string{
		"Content-Type":         "application/json",
		"OK-ACCESS-KEY":        c.apiKey,
		"OK-ACCESS-SIGN":       signature,
		"OK-ACCESS-TIMESTAMP":  timestamp,
		"OK-ACCESS-PASSPHRASE": c.passphrase,
		"OK-ACCESS-PROJECT":    c.projectID,
	}
}

func (c *OKXClient) Quote(fromToken, toToken, amount, slippage string) (*TokenInfo, error) {
	const maxRetries = 3

	endpoint := "/api/v5/dex/aggregator/quote"
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// 构造查询参数
	params := url.Values{}
	params.Set("chainId", "501")
	params.Set("fromTokenAddress", fromToken)
	params.Set("toTokenAddress", toToken)
	params.Set("amount", amount)
	params.Set("slippage", slippage)

	queryString := "?" + params.Encode()
	headers := c.getHeaders(timestamp, "GET", endpoint, queryString)

	// 重试机制
	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", c.baseURL+endpoint+queryString, nil)
		if err != nil {
			return nil, fmt.Errorf("构造 HTTP 请求失败: %v", err)
		}
		// 设置 Headers
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		// 发送 HTTP 请求
		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Infof("[OKX]请求失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2) // 退避策略
			continue
		}
		defer resp.Body.Close()

		// 处理错误状态码
		if resp.StatusCode != http.StatusOK {
			var errMsg bytes.Buffer
			errMsg.ReadFrom(resp.Body)
			log.Infof("[OKX]HTTP 状态码错误 (尝试 %d/%d): %d, 响应: %s\n", attempt, maxRetries, resp.StatusCode, errMsg.String())
			time.Sleep(time.Second * 2)
			continue
		}

		// 解析 JSON
		var response struct {
			Code string `json:"code"`
			Msg  string `json:"msg"`
			Data []struct {
				FromToken struct {
					TokenSymbol    string `json:"tokenSymbol"`
					Decimal        string `json:"decimal"`
					TokenUnitPrice string `json:"tokenUnitPrice"`
				} `json:"fromToken"`
				ToToken struct {
					TokenSymbol    string `json:"tokenSymbol"`
					Decimal        string `json:"decimal"`
					TokenUnitPrice string `json:"tokenUnitPrice"`
				} `json:"toToken"`
			} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Infof("[OKX]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}

		// 校验响应数据
		if response.Code != "0" || len(response.Data) == 0 {
			if response.Code == "82000" {
				return nil, fmt.Errorf(response.Msg)
			}
			log.Infof("[OKX]查询代币信息失败 (尝试 %d/%d): %v\n", attempt, maxRetries, response)
			time.Sleep(time.Second * 2)
			continue
		}

		// 解析返回数据
		quote := response.Data[0]
		fromDecimals, _ := strconv.Atoi(quote.FromToken.Decimal)
		toDecimals, _ := strconv.Atoi(quote.ToToken.Decimal)

		return &TokenInfo{
			FromToken: struct {
				Symbol   string
				Decimals int
				Price    string
			}{
				Symbol:   quote.FromToken.TokenSymbol,
				Decimals: fromDecimals,
				Price:    quote.FromToken.TokenUnitPrice,
			},
			ToToken: struct {
				Symbol   string
				Decimals int
				Price    string
			}{
				Symbol:   quote.ToToken.TokenSymbol,
				Decimals: toDecimals,
				Price:    quote.ToToken.TokenUnitPrice,
			},
		}, nil
	}

	return nil, fmt.Errorf("request failed, try again")
}

func (c *OKXClient) Swap(fromToken, toToken, amount, slippage string, userWallet string) (interface{}, error) {
	const maxRetries = 3

	endpoint := "/api/v5/dex/aggregator/swap"
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// 构造查询参数
	params := url.Values{}
	params.Set("chainId", "501")
	params.Set("fromTokenAddress", fromToken)
	params.Set("toTokenAddress", toToken)
	params.Set("amount", amount)
	params.Set("slippage", slippage)
	params.Set("userWalletAddress", userWallet)

	queryString := "?" + params.Encode()
	headers := c.getHeaders(timestamp, "GET", endpoint, queryString)

	// 重试机制
	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", c.baseURL+endpoint+queryString, nil)
		if err != nil {
			return nil, fmt.Errorf("构造 HTTP 请求失败: %v", err)
		}
		// 设置 Headers
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		// 发送 HTTP 请求
		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Infof("[OKX]请求失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2) // 退避策略
			continue
		}
		defer resp.Body.Close()

		// 处理错误状态码
		if resp.StatusCode != http.StatusOK {
			var errMsg bytes.Buffer
			errMsg.ReadFrom(resp.Body)
			log.Infof("[OKX]HTTP 状态码错误 (尝试 %d/%d): %d, 响应: %s\n", attempt, maxRetries, resp.StatusCode, errMsg.String())
			time.Sleep(time.Second * 2)
			continue
		}

		// 解析 JSON
		var response map[string]interface{}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Infof("[OKX]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}

		// 校验响应数据
		if response["code"].(string) != "0" {
			return nil, fmt.Errorf(response["code"].(string) + "-" + response["msg"].(string))
		}

		// 解析返回数据
		quote := response["data"]
		return quote, nil
	}

	return nil, fmt.Errorf("request failed, try again")
}

func (c *OKXClient) WalletValueSpec(wallet, chains string, assetType string, excludeRisk bool) ([]WalletAsset, error) {
	const maxRetries = 3

	endpoint := "/api/v5/wallet/asset/total-value-by-address"
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// 构造查询参数
	params := url.Values{}
	params.Set("address", wallet)
	params.Set("chains", chains)
	params.Set("assetType", assetType)
	if excludeRisk {
		params.Set("excludeRiskToken", "true")
	} else {
		params.Set("excludeRiskToken", "false")
	}

	queryString := "?" + params.Encode()
	headers := c.getHeaders(timestamp, "GET", endpoint, queryString)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", c.baseURL+endpoint+queryString, nil)
		if err != nil {
			return nil, fmt.Errorf("构造 HTTP 请求失败: %v", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Infof("[OKX]请求失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2) // 退避策略
			continue
		}

		if resp.StatusCode != http.StatusOK {
			var errMsg bytes.Buffer
			errMsg.ReadFrom(resp.Body)
			log.Infof("[OKX]HTTP 状态码错误 (尝试 %d/%d): %d, 响应: %s\n", attempt, maxRetries, resp.StatusCode, errMsg.String())
			time.Sleep(time.Second * 2)
			continue
		}

		var response struct {
			Code string        `json:"code"`
			Msg  string        `json:"msg"`
			Data []WalletAsset `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Infof("[OKX]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}

		if response.Code != "0" {
			return nil, fmt.Errorf(response.Code + "-" + response.Msg)
		}

		quote := response.Data
		return quote, nil
	}

	return nil, fmt.Errorf("request failed, try again")
}

func (c *OKXClient) WalletAssets(wallet, chains string, filter string) ([]AssetDetail, error) {
	const maxRetries = 3

	endpoint := "/api/v5/wallet/asset/all-token-balances-by-address"
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// 构造查询参数
	params := url.Values{}
	params.Set("address", wallet)
	params.Set("chains", chains)
	params.Set("filter", filter)

	queryString := "?" + params.Encode()
	headers := c.getHeaders(timestamp, "GET", endpoint, queryString)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", c.baseURL+endpoint+queryString, nil)
		if err != nil {
			return nil, fmt.Errorf("构造 HTTP 请求失败: %v", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Infof("[OKX]请求失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2) // 退避策略
			continue
		}

		if resp.StatusCode != http.StatusOK {
			var errMsg bytes.Buffer
			errMsg.ReadFrom(resp.Body)
			log.Infof("[OKX]HTTP 状态码错误 (尝试 %d/%d): %d, 响应: %s\n", attempt, maxRetries, resp.StatusCode, errMsg.String())
			time.Sleep(time.Second * 2)
			continue
		}

		var response struct {
			Code string `json:"code"`
			Msg  string `json:"msg"`
			Data []struct {
				TokenAssets []AssetDetail `json:"tokenAssets"`
			} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Infof("[OKX]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}

		if response.Code != "0" {
			return nil, fmt.Errorf(response.Code + "-" + response.Msg)
		}

		quote := response.Data
		if len(quote) == 0 {
			return []AssetDetail{}, nil
		}
		return quote[0].TokenAssets, nil
	}

	return nil, fmt.Errorf("request failed, try again")
}

func (c *OKXClient) CoinPrice(requestData []TokenPriceRequest) ([]TokenPriceResponse, error) {
	const maxRetries = 3

	endpoint := "/api/v5/wallet/token/current-price"
	timestamp := time.Now().UTC().Format(time.RFC3339)

	jsonBody, _ := json.Marshal(requestData)

	headers := c.getHeaders(timestamp, "POST", endpoint, string(jsonBody))

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("POST", c.baseURL+endpoint, bytes.NewBuffer(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("构造 HTTP 请求失败: %v", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Infof("[OKX]请求失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2) // 退避策略
			continue
		}

		if resp.StatusCode != http.StatusOK {
			var errMsg bytes.Buffer
			errMsg.ReadFrom(resp.Body)
			log.Infof("[OKX]HTTP 状态码错误 (尝试 %d/%d): %d, 响应: %s\n", attempt, maxRetries, resp.StatusCode, errMsg.String())
			time.Sleep(time.Second * 2)
			continue
		}

		var response struct {
			Code string               `json:"code"`
			Msg  string               `json:"msg"`
			Data []TokenPriceResponse `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Infof("[OKX]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}

		if response.Code != "0" {
			return nil, fmt.Errorf(response.Code + "-" + response.Msg)
		}

		quote := response.Data
		if len(quote) == 0 {
			return []TokenPriceResponse{}, nil
		}
		return quote, nil
	}

	return nil, fmt.Errorf("request failed, try again")
}

func (c *OKXClient) WalletTransactions(wallet, chains string, tokenAddress string, beginTs, endTs int, limit int) ([]TransactionListResponse, error) {
	const maxRetries = 3

	endpoint := "/api/v5/wallet/post-transaction/transactions-by-address"
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// 构造查询参数
	params := url.Values{}
	params.Set("chains", chains)
	params.Set("address", wallet)
	if len(tokenAddress) > 0 {
		params.Set("tokenAddress", tokenAddress)
	}
	if beginTs > 0 {
		params.Set("begin", fmt.Sprintf("%d", beginTs))
	}
	if endTs > 0 {
		params.Set("end", fmt.Sprintf("%d", endTs))
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	queryString := "?" + params.Encode()
	headers := c.getHeaders(timestamp, "GET", endpoint, queryString)

	// 重试机制
	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", c.baseURL+endpoint+queryString, nil)
		if err != nil {
			return nil, fmt.Errorf("构造 HTTP 请求失败: %v", err)
		}
		// 设置 Headers
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		// 发送 HTTP 请求
		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Infof("[OKX]请求失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2) // 退避策略
			continue
		}
		defer resp.Body.Close()

		// 处理错误状态码
		if resp.StatusCode != http.StatusOK {
			var errMsg bytes.Buffer
			errMsg.ReadFrom(resp.Body)
			log.Infof("[OKX]HTTP 状态码错误 (尝试 %d/%d): %d, 响应: %s\n", attempt, maxRetries, resp.StatusCode, errMsg.String())
			time.Sleep(time.Second * 2)
			continue
		}

		// 解析 JSON
		var response struct {
			Code string                    `json:"code"`
			Msg  string                    `json:"msg"`
			Data []TransactionListResponse `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Infof("[OKX]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}

		// 校验响应数据
		if response.Code != "0" {
			return nil, fmt.Errorf(response.Code + "-" + response.Msg)
		}

		// 解析返回数据
		quote := response.Data
		return quote, nil
	}

	return nil, fmt.Errorf("request failed, try again")
}
