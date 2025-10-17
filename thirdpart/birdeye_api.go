package thirdpart

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"chaos/api/api/common"
	"chaos/api/log"
)

type BirdeydClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	mu         sync.Mutex // 确保并发安全
}

var birdClient *BirdeydClient

func GetBirdClient() (*BirdeydClient, error) {
	if birdClient != nil {
		return birdClient, nil
	}
	return newBirdClient()
}

func newBirdClient() (*BirdeydClient, error) {
	client := &BirdeydClient{
		apiKey:     os.Getenv("BIRDEYE_API_KEY"),
		baseURL:    "https://public-api.birdeye.so",
		httpClient: &http.Client{Timeout: 10 * time.Second}, // 配置 HTTP 连接池
	}

	if client.apiKey == "" {
		return nil, fmt.Errorf("缺少必要的 Bird API 环境变量")
	}

	return client, nil
}

func (c *BirdeydClient) getHeaders(chain string) map[string]string {
	hm := map[string]string{
		"accept":    "application/json",
		"X-API-KEY": c.apiKey,
	}
	if len(chain) > 0 {
		hm["x-chain"] = chain
	}
	return hm
}

func (c *BirdeydClient) TokenTrending(chain string, params TokenListV3Request, pn int, ps int) ([]TokenListV3Resp, error) {
	const maxRetries = 3
	endpoint := "/defi/v3/token/list"
	p := common.QueryParams[TokenListV3Request]{
		Data: params,
	}
	q := fmt.Sprintf(p.BuildQueryString()+"&offset=%d&limit=%d", pn, ps)
	headers := c.getHeaders(chain)
	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", c.baseURL+endpoint+q, nil)
		if err != nil {
			return nil, fmt.Errorf("构造 HTTP 请求失败: %v", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Infof("[Birdeye]请求失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2) // 退避策略
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errMsg bytes.Buffer
			errMsg.ReadFrom(resp.Body)
			log.Infof("[Birdeye]HTTP 状态码错误 (尝试 %d/%d): %d, 响应: %s\n", attempt, maxRetries, resp.StatusCode, errMsg.String())
			time.Sleep(time.Second * 2)
			continue
		}

		type itemStruct struct {
			Items []TokenListV3Resp `json:"items"`
		}
		var response BirdResponse[itemStruct]

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Infof("[Birdeye]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}

		return response.Data.Items, nil
	}

	return nil, fmt.Errorf("request failed, try again")
}

func (c *BirdeydClient) FetchKlineData(ca, chain string, interval string, currency string, from, to int64) ([]OHLCVResp, error) {
	const maxRetries = 3

	endpoint := "/defi/ohlcv"

	params := url.Values{}
	params.Set("address", ca)
	params.Set("type", interval)
	params.Set("currency", currency)
	params.Set("time_from", fmt.Sprintf("%d", from))
	params.Set("time_to", fmt.Sprintf("%d", to))

	queryString := "?" + params.Encode()
	headers := c.getHeaders(chain)

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
			log.Infof("[Birdeye]请求失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2) // 退避策略
			continue
		}
		defer resp.Body.Close()

		// 处理错误状态码
		if resp.StatusCode != http.StatusOK {
			var errMsg bytes.Buffer
			errMsg.ReadFrom(resp.Body)
			log.Infof("[Birdeye]HTTP 状态码错误 (尝试 %d/%d): %d, 响应: %s\n", attempt, maxRetries, resp.StatusCode, errMsg.String())
			time.Sleep(time.Second * 2)
			continue
		}

		// 解析 JSON
		type itemStruct struct {
			Items []OHLCVResp `json:"items"`
		}
		var response BirdResponse[itemStruct]

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Infof("[Birdeye]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}

		return response.Data.Items, nil
	}

	return nil, fmt.Errorf("request failed, try again")
}

func (c *BirdeydClient) TokenHolder(chain, ca string, pn int, ps int) ([]TokenHolderResp, error) {
	const maxRetries = 3
	endpoint := "/defi/v3/token/holder"
	params := url.Values{}
	params.Set("address", ca)
	params.Set("limit", fmt.Sprintf("%d", ps))
	params.Set("offset", fmt.Sprintf("%d", pn))

	queryString := "?" + params.Encode()
	headers := c.getHeaders(chain)
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
			log.Infof("[Birdeye]请求失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errMsg bytes.Buffer
			errMsg.ReadFrom(resp.Body)
			log.Infof("[Birdeye]HTTP 状态码错误 (尝试 %d/%d): %d, 响应: %s\n", attempt, maxRetries, resp.StatusCode, errMsg.String())
			time.Sleep(time.Second * 2)
			continue
		}

		type itemStruct struct {
			Items []TokenHolderResp `json:"items"`
		}
		var response BirdResponse[itemStruct]

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Infof("[Birdeye]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}

		return response.Data.Items, nil
	}

	return nil, fmt.Errorf("request failed, try again")
}

func (c *BirdeydClient) TokenSearch(params TokenSearchRequest, pn int, ps int) ([]TokenSearchTokenResp, []TokenSearchMarketResp, error) {
	const maxRetries = 3
	endpoint := "/defi/v3/search"

	p := common.QueryParams[TokenSearchRequest]{
		Data: params,
	}
	queryString := fmt.Sprintf(p.BuildQueryString()+"&offset=%d&limit=%d", pn, ps)

	headers := c.getHeaders("")
	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", c.baseURL+endpoint+queryString, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("构造 HTTP 请求失败: %v", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Infof("[Birdeye]请求失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errMsg bytes.Buffer
			errMsg.ReadFrom(resp.Body)
			log.Infof("[Birdeye]HTTP 状态码错误 (尝试 %d/%d): %d, 响应: %s\n", attempt, maxRetries, resp.StatusCode, errMsg.String())
			time.Sleep(time.Second * 2)
			continue
		}

		type itemStruct struct {
			Items []TokenSearchCategory `json:"items"`
		}
		var response BirdResponse[itemStruct]

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Infof("[Birdeye]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}

		var tokens []TokenSearchTokenResp
		var markets []TokenSearchMarketResp
		for _, category := range response.Data.Items {
			switch category.Type {
			case "token":
				rss := string(category.Result)
				err := json.Unmarshal(category.Result, &tokens)
				if err != nil {
					fmt.Println("解析 Token 失败:", err)
					continue
				}
				fmt.Println("解析到的 Token 数据:", tokens, rss)
			case "market":
				err := json.Unmarshal(category.Result, &markets)
				if err != nil {
					fmt.Println("解析 Market 失败:", err)
					continue
				}
				fmt.Println("解析到的 Market 数据:", markets)
			}
		}

		return tokens, markets, nil
	}

	return nil, nil, fmt.Errorf("request failed, try again")
}

func (c *BirdeydClient) TokenMetas(chain, cas string) (map[string]TokenMeta, error) {
	const maxRetries = 3
	endpoint := "/defi/v3/token/meta-data/multiple"
	params := url.Values{}
	params.Set("list_address", cas)

	queryString := "?" + params.Encode()
	headers := c.getHeaders(chain)
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
			log.Infof("[Birdeye]请求失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errMsg bytes.Buffer
			errMsg.ReadFrom(resp.Body)
			log.Infof("[Birdeye]HTTP 状态码错误 (尝试 %d/%d): %d, 响应: %s\n", attempt, maxRetries, resp.StatusCode, errMsg.String())
			time.Sleep(time.Second * 2)
			continue
		}

		var response BirdResponse[map[string]TokenMeta]

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Infof("[Birdeye]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}

		return response.Data, nil
	}

	return nil, fmt.Errorf("request failed, try again")
}

func (c *BirdeydClient) WalletPortfolio(chain, wallet string) (*WalletInfo, error) {
	const maxRetries = 3
	endpoint := "/v1/wallet/token_list"
	params := url.Values{}
	params.Set("wallet", wallet)

	queryString := "?" + params.Encode()
	headers := c.getHeaders(chain)
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
			log.Infof("[Birdeye]请求失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errMsg bytes.Buffer
			errMsg.ReadFrom(resp.Body)
			log.Infof("[Birdeye]HTTP 状态码错误 (尝试 %d/%d): %d, 响应: %s\n", attempt, maxRetries, resp.StatusCode, errMsg.String())
			time.Sleep(time.Second * 2)
			continue
		}

		var response BirdResponse[WalletInfo]

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Infof("[Birdeye]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}

		return &response.Data, nil
	}

	return nil, fmt.Errorf("request failed, try again")
}

func (c *BirdeydClient) TokenNewList(chain string, until_time uint64, meme_platform_enable bool, ps int) ([]NewToken, error) {
	const maxRetries = 3
	endpoint := "/defi/v2/tokens/new_listing"
	params := url.Values{}
	params.Set("time_to", fmt.Sprintf("%d", until_time))
	params.Set("limit", fmt.Sprintf("%d", ps))
	params.Set("meme_platform_enabled", fmt.Sprintf("%t", meme_platform_enable))

	queryString := "?" + params.Encode()
	headers := c.getHeaders(chain)
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
			log.Infof("[Birdeye]请求失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errMsg bytes.Buffer
			errMsg.ReadFrom(resp.Body)
			log.Infof("[Birdeye]HTTP 状态码错误 (尝试 %d/%d): %d, 响应: %s\n", attempt, maxRetries, resp.StatusCode, errMsg.String())
			time.Sleep(time.Second * 2)
			continue
		}

		type itemStruct struct {
			Items []NewToken `json:"items"`
		}
		var response BirdResponse[itemStruct]

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Infof("[Birdeye]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}

		return response.Data.Items, nil
	}

	return nil, fmt.Errorf("request failed, try again")
}
