package thirdpart

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"chaos/api/log"
)

type Extensions struct {
	CoingeckoID string `json:"coingecko_id"`
	SerumV3Usdc string `json:"serumV3Usdc"`
	SerumV3Usdt string `json:"serumV3Usdt"`
	Website     string `json:"website"`
	Telegram    string `json:"telegram"`
	Description string `json:"description"`
	Twitter     string `json:"twitter"`
	Discord     string `json:"discord"`
	Medium      string `json:"medium"`
}

type BeToken struct {
	Address    string     `json:"address"`
	Symbol     string     `json:"symbol"`
	Name       string     `json:"name"`
	Decimals   int        `json:"decimals"`
	Extensions Extensions `json:"extensions"`
	LogoURI    string     `json:"logo_uri"`
}
type BeTokenOverview struct {
	Address                      string          `json:"address"`
	Decimals                     int             `json:"decimals"`
	Symbol                       string          `json:"symbol"`
	Name                         string          `json:"name"`
	Extensions                   Extensions      `json:"extensions"`
	LogoURI                      string          `json:"logoURI"`
	Liquidity                    decimal.Decimal `json:"liquidity"`
	LastTradeUnixTime            int64           `json:"lastTradeUnixTime"`
	LastTradeHumanTime           string          `json:"lastTradeHumanTime"`
	Price                        decimal.Decimal `json:"price"`
	History30mPrice              decimal.Decimal `json:"history30mPrice"`
	PriceChange30mPercent        decimal.Decimal `json:"priceChange30mPercent"`
	History1hPrice               decimal.Decimal `json:"history1hPrice"`
	PriceChange1hPercent         decimal.Decimal `json:"priceChange1hPercent"`
	History2hPrice               decimal.Decimal `json:"history2hPrice"`
	PriceChange2hPercent         decimal.Decimal `json:"priceChange2hPercent"`
	History4hPrice               decimal.Decimal `json:"history4hPrice"`
	PriceChange4hPercent         decimal.Decimal `json:"priceChange4hPercent"`
	History6hPrice               decimal.Decimal `json:"history6hPrice"`
	PriceChange6hPercent         decimal.Decimal `json:"priceChange6hPercent"`
	History8hPrice               decimal.Decimal `json:"history8hPrice"`
	PriceChange8hPercent         decimal.Decimal `json:"priceChange8hPercent"`
	History12hPrice              decimal.Decimal `json:"history12hPrice"`
	PriceChange12hPercent        decimal.Decimal `json:"priceChange12hPercent"`
	History24hPrice              decimal.Decimal `json:"history24hPrice"`
	PriceChange24hPercent        decimal.Decimal `json:"priceChange24hPercent"`
	UniqueWallet30m              int             `json:"uniqueWallet30m"`
	UniqueWalletHistory30m       int             `json:"uniqueWalletHistory30m"`
	UniqueWallet30mChangePercent decimal.Decimal `json:"uniqueWallet30mChangePercent"`
	UniqueWallet1h               int             `json:"uniqueWallet1h"`
	UniqueWalletHistory1h        int             `json:"uniqueWalletHistory1h"`
	UniqueWallet1hChangePercent  decimal.Decimal `json:"uniqueWallet1hChangePercent"`
	UniqueWallet2h               int             `json:"uniqueWallet2h"`
	UniqueWalletHistory2h        int             `json:"uniqueWalletHistory2h"`
	UniqueWallet2hChangePercent  decimal.Decimal `json:"uniqueWallet2hChangePercent"`
	UniqueWallet4h               int             `json:"uniqueWallet4h"`
	UniqueWalletHistory4h        int             `json:"uniqueWalletHistory4h"`
	UniqueWallet4hChangePercent  decimal.Decimal `json:"uniqueWallet4hChangePercent"`
	UniqueWallet8h               int             `json:"uniqueWallet8h"`
	UniqueWalletHistory8h        int             `json:"uniqueWalletHistory8h"`
	UniqueWallet8hChangePercent  decimal.Decimal `json:"uniqueWallet8hChangePercent"`
	UniqueWallet24h              int             `json:"uniqueWallet24h"`
	UniqueWalletHistory24h       int             `json:"uniqueWalletHistory24h"`
	UniqueWallet24hChangePercent decimal.Decimal `json:"uniqueWallet24hChangePercent"`
	TotalSupply                  decimal.Decimal `json:"totalSupply"`
	Fdv                          decimal.Decimal `json:"fdv"`
	CirculatingSupply            decimal.Decimal `json:"circulatingSupply"`
	MarketCap                    decimal.Decimal `json:"marketCap"`
	Holder                       int             `json:"holder"`
	Trade30m                     int             `json:"trade30m"`
	TradeHistory30m              int             `json:"tradeHistory30m"`
	Trade30mChangePercent        decimal.Decimal `json:"trade30mChangePercent"`
	Sell30m                      int             `json:"sell30m"`
	SellHistory30m               int             `json:"sellHistory30m"`
	Sell30mChangePercent         decimal.Decimal `json:"sell30mChangePercent"`
	Buy30m                       int             `json:"buy30m"`
	BuyHistory30m                int             `json:"buyHistory30m"`
	Buy30mChangePercent          decimal.Decimal `json:"buy30mChangePercent"`
	V30m                         decimal.Decimal `json:"v30m"`
	V30mUSD                      decimal.Decimal `json:"v30mUSD"`
	VHistory30m                  decimal.Decimal `json:"vHistory30m"`
	VHistory30mUSD               decimal.Decimal `json:"vHistory30mUSD"`
	V30mChangePercent            decimal.Decimal `json:"v30mChangePercent"`
	VBuy30m                      decimal.Decimal `json:"vBuy30m"`
	VBuy30mUSD                   decimal.Decimal `json:"vBuy30mUSD"`
	VBuyHistory30m               decimal.Decimal `json:"vBuyHistory30m"`
	VBuyHistory30mUSD            decimal.Decimal `json:"vBuyHistory30mUSD"`
	VBuy30mChangePercent         decimal.Decimal `json:"vBuy30mChangePercent"`
	VSell30m                     decimal.Decimal `json:"vSell30m"`
	VSell30mUSD                  decimal.Decimal `json:"vSell30mUSD"`
	VSellHistory30m              decimal.Decimal `json:"vSellHistory30m"`
	VSellHistory30mUSD           decimal.Decimal `json:"vSellHistory30mUSD"`
	VSell30mChangePercent        decimal.Decimal `json:"vSell30mChangePercent"`
	Trade1h                      int             `json:"trade1h"`
	TradeHistory1h               int             `json:"tradeHistory1h"`
	Trade1hChangePercent         decimal.Decimal `json:"trade1hChangePercent"`
	Sell1h                       int             `json:"sell1h"`
	SellHistory1h                int             `json:"sellHistory1h"`
	Sell1hChangePercent          decimal.Decimal `json:"sell1hChangePercent"`
	Buy1h                        int             `json:"buy1h"`
	BuyHistory1h                 int             `json:"buyHistory1h"`
	Buy1hChangePercent           decimal.Decimal `json:"buy1hChangePercent"`
	V1h                          decimal.Decimal `json:"v1h"`
	V1hUSD                       decimal.Decimal `json:"v1hUSD"`
	VHistory1h                   decimal.Decimal `json:"vHistory1h"`
	VHistory1hUSD                decimal.Decimal `json:"vHistory1hUSD"`
	V1hChangePercent             decimal.Decimal `json:"v1hChangePercent"`
	VBuy1h                       decimal.Decimal `json:"vBuy1h"`
	VBuy1hUSD                    decimal.Decimal `json:"vBuy1hUSD"`
	VBuyHistory1h                decimal.Decimal `json:"vBuyHistory1h"`
	VBuyHistory1hUSD             decimal.Decimal `json:"vBuyHistory1hUSD"`
	VBuy1hChangePercent          decimal.Decimal `json:"vBuy1hChangePercent"`
	VSell1h                      decimal.Decimal `json:"vSell1h"`
	VSell1hUSD                   decimal.Decimal `json:"vSell1hUSD"`
	VSellHistory1h               decimal.Decimal `json:"vSellHistory1h"`
	VSellHistory1hUSD            decimal.Decimal `json:"vSellHistory1hUSD"`
	VSell1hChangePercent         decimal.Decimal `json:"vSell1hChangePercent"`
	Trade2h                      int             `json:"trade2h"`
	TradeHistory2h               int             `json:"tradeHistory2h"`
	Trade2hChangePercent         decimal.Decimal `json:"trade2hChangePercent"`
	Sell2h                       int             `json:"sell2h"`
	SellHistory2h                int             `json:"sellHistory2h"`
	Sell2hChangePercent          decimal.Decimal `json:"sell2hChangePercent"`
	Buy2h                        int             `json:"buy2h"`
	BuyHistory2h                 int             `json:"buyHistory2h"`
	Buy2hChangePercent           decimal.Decimal `json:"buy2hChangePercent"`
	V2h                          decimal.Decimal `json:"v2h"`
	V2hUSD                       decimal.Decimal `json:"v2hUSD"`
	VHistory2h                   decimal.Decimal `json:"vHistory2h"`
	VHistory2hUSD                decimal.Decimal `json:"vHistory2hUSD"`
	V2hChangePercent             decimal.Decimal `json:"v2hChangePercent"`
	VBuy2h                       decimal.Decimal `json:"vBuy2h"`
	VBuy2hUSD                    decimal.Decimal `json:"vBuy2hUSD"`
	VBuyHistory2h                decimal.Decimal `json:"vBuyHistory2h"`
	VBuyHistory2hUSD             decimal.Decimal `json:"vBuyHistory2hUSD"`
	VBuy2hChangePercent          decimal.Decimal `json:"vBuy2hChangePercent"`
	VSell2h                      decimal.Decimal `json:"vSell2h"`
	VSell2hUSD                   decimal.Decimal `json:"vSell2hUSD"`
	VSellHistory2h               decimal.Decimal `json:"vSellHistory2h"`
	VSellHistory2hUSD            decimal.Decimal `json:"vSellHistory2hUSD"`
	VSell2hChangePercent         decimal.Decimal `json:"vSell2hChangePercent"`
	Trade4h                      int             `json:"trade4h"`
	TradeHistory4h               int             `json:"tradeHistory4h"`
	Trade4hChangePercent         decimal.Decimal `json:"trade4hChangePercent"`
	Sell4h                       int             `json:"sell4h"`
	SellHistory4h                int             `json:"sellHistory4h"`
	Sell4hChangePercent          decimal.Decimal `json:"sell4hChangePercent"`
	Buy4h                        int             `json:"buy4h"`
	BuyHistory4h                 int             `json:"buyHistory4h"`
	Buy4hChangePercent           decimal.Decimal `json:"buy4hChangePercent"`
	V4h                          decimal.Decimal `json:"v4h"`
	V4hUSD                       decimal.Decimal `json:"v4hUSD"`
	VHistory4h                   decimal.Decimal `json:"vHistory4h"`
	VHistory4hUSD                decimal.Decimal `json:"vHistory4hUSD"`
	V4hChangePercent             decimal.Decimal `json:"v4hChangePercent"`
	VBuy4h                       decimal.Decimal `json:"vBuy4h"`
	VBuy4hUSD                    decimal.Decimal `json:"vBuy4hUSD"`
	VBuyHistory4h                decimal.Decimal `json:"vBuyHistory4h"`
	VBuyHistory4hUSD             decimal.Decimal `json:"vBuyHistory4hUSD"`
	VBuy4hChangePercent          decimal.Decimal `json:"vBuy4hChangePercent"`
	VSell4h                      decimal.Decimal `json:"vSell4h"`
	VSell4hUSD                   decimal.Decimal `json:"vSell4hUSD"`
	VSellHistory4h               decimal.Decimal `json:"vSellHistory4h"`
	VSellHistory4hUSD            decimal.Decimal `json:"vSellHistory4hUSD"`
	VSell4hChangePercent         decimal.Decimal `json:"vSell4hChangePercent"`
	Trade8h                      int             `json:"trade8h"`
	TradeHistory8h               int             `json:"tradeHistory8h"`
	Trade8hChangePercent         decimal.Decimal `json:"trade8hChangePercent"`
	Sell8h                       int             `json:"sell8h"`
	SellHistory8h                int             `json:"sellHistory8h"`
	Sell8hChangePercent          decimal.Decimal `json:"sell8hChangePercent"`
	Buy8h                        int             `json:"buy8h"`
	BuyHistory8h                 int             `json:"buyHistory8h"`
	Buy8hChangePercent           decimal.Decimal `json:"buy8hChangePercent"`
	V8h                          decimal.Decimal `json:"v8h"`
	V8hUSD                       decimal.Decimal `json:"v8hUSD"`
	VHistory8h                   decimal.Decimal `json:"vHistory8h"`
	VHistory8hUSD                decimal.Decimal `json:"vHistory8hUSD"`
	V8hChangePercent             decimal.Decimal `json:"v8hChangePercent"`
	VBuy8h                       decimal.Decimal `json:"vBuy8h"`
	VBuy8hUSD                    decimal.Decimal `json:"vBuy8hUSD"`
	VBuyHistory8h                decimal.Decimal `json:"vBuyHistory8h"`
	VBuyHistory8hUSD             decimal.Decimal `json:"vBuyHistory8hUSD"`
	VBuy8hChangePercent          decimal.Decimal `json:"vBuy8hChangePercent"`
	VSell8h                      decimal.Decimal `json:"vSell8h"`
	VSell8hUSD                   decimal.Decimal `json:"vSell8hUSD"`
	VSellHistory8h               decimal.Decimal `json:"vSellHistory8h"`
	VSellHistory8hUSD            decimal.Decimal `json:"vSellHistory8hUSD"`
	VSell8hChangePercent         decimal.Decimal `json:"vSell8hChangePercent"`
	Trade24h                     int             `json:"trade24h"`
	TradeHistory24h              int             `json:"tradeHistory24h"`
	Trade24hChangePercent        decimal.Decimal `json:"trade24hChangePercent"`
	Sell24h                      int             `json:"sell24h"`
	SellHistory24h               int             `json:"sellHistory24h"`
	Sell24hChangePercent         decimal.Decimal `json:"sell24hChangePercent"`
	Buy24h                       int             `json:"buy24h"`
	BuyHistory24h                int             `json:"buyHistory24h"`
	Buy24hChangePercent          decimal.Decimal `json:"buy24hChangePercent"`
	V24h                         decimal.Decimal `json:"v24h"`
	V24hUSD                      decimal.Decimal `json:"v24hUSD"`
	VHistory24h                  decimal.Decimal `json:"vHistory24h"`
	VHistory24hUSD               decimal.Decimal `json:"vHistory24hUSD"`
	V24hChangePercent            decimal.Decimal `json:"v24hChangePercent"`
	VBuy24h                      decimal.Decimal `json:"vBuy24h"`
	VBuy24hUSD                   decimal.Decimal `json:"vBuy24hUSD"`
	VBuyHistory24h               decimal.Decimal `json:"vBuyHistory24h"`
	VBuyHistory24hUSD            decimal.Decimal `json:"vBuyHistory24hUSD"`
	VBuy24hChangePercent         decimal.Decimal `json:"vBuy24hChangePercent"`
	VSell24h                     decimal.Decimal `json:"vSell24h"`
	VSell24hUSD                  decimal.Decimal `json:"vSell24hUSD"`
	VSellHistory24h              decimal.Decimal `json:"vSellHistory24h"`
	VSellHistory24hUSD           decimal.Decimal `json:"vSellHistory24hUSD"`
	VSell24hChangePercent        decimal.Decimal `json:"vSell24hChangePercent"`
	Watch                        *string         `json:"watch"` // 可为空，使用指针支持 null
	NumberMarkets                int             `json:"numberMarkets"`
}

func (c *BirdeydClient) Metadata(cas []string, chain string) ([]BeToken, error) {
	const maxRetries = 3

	endpoint := "/defi/v3/token/meta-data/multiple"

	params := url.Values{}
	params.Add("list_address", strings.Join(cas, ","))

	queryString := "?" + params.Encode()
	headers := c.getHeaders(chain)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", c.baseURL+endpoint+queryString, nil)
		if err != nil {
			return []BeToken{}, fmt.Errorf("构造 HTTP 请求失败: %v", err)
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

		var response BirdResponse[map[string]BeToken]

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Infof("[OKX]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}
		data := response.Data
		tokens := make([]BeToken, 0)
		for _, ca := range cas {
			d, exists := data[ca]
			if exists {
				tokens = append(tokens, d)
			}
		}
		return tokens, nil
	}

	return nil, fmt.Errorf("request failed, try again")
}

func (c *BirdeydClient) TokenOverview(ca string, chain string) (*BeTokenOverview, error) {
	const maxRetries = 3

	endpoint := "/defi/token_overview"
	params := url.Values{}
	params.Add("address", ca)

	queryString := "?" + params.Encode()
	headers := c.getHeaders(chain)
	req, err := http.NewRequest("GET", c.baseURL+endpoint+queryString, nil)
	if err != nil {
		return nil, fmt.Errorf("构造 HTTP 请求失败: %v", err)
	}
	// 设置 Headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 发送 HTTP 请求
		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Infof("[OKX]请求失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2) // 退避策略
			continue
		}
		defer resp.Body.Close()

		// 处理错误状态码
		code := resp.StatusCode
		if code != http.StatusOK {
			var errMsg bytes.Buffer
			_, err := errMsg.ReadFrom(resp.Body)
			if err != nil {
				return nil, err
			}
			log.Infof("[Birdeyd]HTTP 状态码错误 (尝试 %d/%d): %d, 响应: %s\n", attempt, maxRetries, code, errMsg.String())
			time.Sleep(time.Second * 2)
			continue
		}

		var response BirdResponse[BeTokenOverview]

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Infof("[Birdeyd]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}
		data := response.Data

		return &data, nil
	}

	return nil, fmt.Errorf("request failed, try again")
}

type BeTokenSecurity struct {
	CreatorAddress                 string           `json:"creatorAddress"`
	CreatorOwnerAddress            string           `json:"creatorOwnerAddress"`
	OwnerAddress                   string           `json:"ownerAddress"`
	OwnerOfOwnerAddress            string           `json:"ownerOfOwnerAddress"`
	CreationTx                     string           `json:"creationTx"`
	CreationTime                   int64            `json:"creationTime"`
	CreationSlot                   int64            `json:"creationSlot"`
	MintTx                         string           `json:"mintTx"`
	MintTime                       int64            `json:"mintTime"`
	MintSlot                       int64            `json:"mintSlot"`
	CreatorBalance                 *decimal.Decimal `json:"creatorBalance"`
	OwnerBalance                   *decimal.Decimal `json:"ownerBalance"`
	OwnerPercentage                *decimal.Decimal `json:"ownerPercentage"`
	CreatorPercentage              *decimal.Decimal `json:"creatorPercentage"`
	MetaplexUpdateAuthority        string           `json:"metaplexUpdateAuthority"`
	MetaplexOwnerUpdateAuthority   string           `json:"metaplexOwnerUpdateAuthority"`
	MetaplexUpdateAuthorityBalance decimal.Decimal  `json:"metaplexUpdateAuthorityBalance"`
	MetaplexUpdateAuthorityPercent decimal.Decimal  `json:"metaplexUpdateAuthorityPercent"`
	MutableMetadata                *bool            `json:"mutableMetadata"`
	Top10HolderBalance             decimal.Decimal  `json:"top10HolderBalance"`
	Top10HolderPercent             decimal.Decimal  `json:"top10HolderPercent"`
	Top10UserBalance               decimal.Decimal  `json:"top10UserBalance"`
	Top10UserPercent               decimal.Decimal  `json:"top10UserPercent"`
	IsTrueToken                    *bool            `json:"isTrueToken"`
	TotalSupply                    decimal.Decimal  `json:"totalSupply"`
	PreMarketHolder                []interface{}    `json:"preMarketHolder"` // 空数组，暂用 interface{}
	LockInfo                       *interface{}     `json:"lockInfo"`        // null，暂用 *interface{}
	Freezeable                     *bool            `json:"freezeable"`
	FreezeAuthority                *string          `json:"freezeAuthority"`
	TransferFeeEnable              *bool            `json:"transferFeeEnable"`
	TransferFeeData                *interface{}     `json:"transferFeeData"` // null，暂用 *interface{}
	IsToken2022                    *bool            `json:"isToken2022"`
	NonTransferable                *bool            `json:"nonTransferable"`
}

func (c *BirdeydClient) TokenSecurity(ca string, chain string) (*BeTokenSecurity, error) {
	const maxRetries = 3

	endpoint := "/defi/token_security"
	params := url.Values{}
	params.Add("address", ca)

	queryString := "?" + params.Encode()
	headers := c.getHeaders(chain)
	req, err := http.NewRequest("GET", c.baseURL+endpoint+queryString, nil)
	if err != nil {
		return nil, fmt.Errorf("构造 HTTP 请求失败: %v", err)
	}
	// 设置 Headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 发送 HTTP 请求
		resp, err := c.httpClient.Do(req)
		if err != nil {
			fmt.Printf("[OKX]请求失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2) // 退避策略
			continue
		}
		defer resp.Body.Close()

		// 处理错误状态码
		code := resp.StatusCode
		if code != http.StatusOK {
			var errMsg bytes.Buffer
			_, err := errMsg.ReadFrom(resp.Body)
			if err != nil {
				return nil, err
			}
			fmt.Printf("[Birdeyd]HTTP 状态码错误 (尝试 %d/%d): %d, 响应: %s\n", attempt, maxRetries, code, errMsg.String())
			time.Sleep(time.Second * 2)
			continue
		}

		var response BirdResponse[BeTokenSecurity]

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			fmt.Printf("[Birdeyd]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}
		data := response.Data

		return &data, nil
	}

	return nil, fmt.Errorf("request failed, try again")
}

type BeTokenTxsItems struct {
	Items []BeTokenTrade `json:"items"`
}
type BeTokenTrade struct {
	TxType        string               `json:"tx_type"`
	TxHash        string               `json:"tx_hash"`
	BlockUnixTime int                  `json:"block_unix_time"`
	VolumeUsd     decimal.Decimal      `json:"volume_usd"`
	Volume        decimal.Decimal      `json:"volume"`
	Owner         string               `json:"owner"`
	Source        string               `json:"source"`
	Side          string               `json:"side"`
	Alias         interface{}          `json:"alias"`
	PricePair     decimal.Decimal      `json:"price_pair"`
	From          BeTokenTradeAmount   `json:"from"`
	To            BeTokenTradeAmount   `json:"to"`
	Tokens        []BeTokenTradeAmount `json:"tokens"`
	PoolId        string               `json:"pool_id"`
}
type BeTokenTradeAmount struct {
	Symbol         string          `json:"symbol"`
	Address        string          `json:"address"`
	Decimals       int             `json:"decimals"`
	Price          decimal.Decimal `json:"price"`
	Amount         decimal.Decimal `json:"amount"`
	UiAmount       decimal.Decimal `json:"ui_amount"`
	UiChangeAmount decimal.Decimal `json:"ui_change_amount"`
}

func (c *BirdeydClient) TokenTxsByTokenAddress(ca string, chain string, txType string,
	source string, owner string, poolId string,
	beforeTime int64, afterTime int64, page, size int) ([]BeTokenTrade, error) {
	const maxRetries = 3

	endpoint := "/defi/v3/token/txs"
	params := url.Values{}
	params.Add("address", ca)
	if len(txType) == 0 {
		txType = "all"
	}
	params.Add("tx_type", txType)
	if len(source) > 0 {
		params.Add("source", source)
	}
	if len(owner) > 0 {
		params.Add("owner", owner)
	}
	if len(poolId) > 0 {
		params.Add("pool_id", poolId)
	}
	if afterTime > 0 {
		params.Add("after_time", strconv.FormatInt(afterTime, 10))
	}
	if beforeTime > 0 {
		params.Add("before_time", strconv.FormatInt(beforeTime, 10))
	}

	queryString := "?" + params.Encode()
	headers := c.getHeaders(chain)
	req, err := http.NewRequest("GET", c.baseURL+endpoint+queryString, nil)
	if err != nil {
		return nil, fmt.Errorf("构造 HTTP 请求失败: %v", err)
	}
	// 设置 Headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 发送 HTTP 请求
		resp, err := c.httpClient.Do(req)
		if err != nil {
			fmt.Printf("[OKX]请求失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2) // 退避策略
			continue
		}
		defer resp.Body.Close()

		// 处理错误状态码
		code := resp.StatusCode
		if code != http.StatusOK {
			var errMsg bytes.Buffer
			_, err := errMsg.ReadFrom(resp.Body)
			if err != nil {
				return nil, err
			}
			fmt.Printf("[Birdeyd]HTTP 状态码错误 (尝试 %d/%d): %d, 响应: %s\n", attempt, maxRetries, code, errMsg.String())
			time.Sleep(time.Second * 2)
			continue
		}

		var response BirdResponse[BeTokenTxsItems]

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			fmt.Printf("[Birdeyd]JSON 解析失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}
		data := response.Data

		return data.Items, nil
	}

	return nil, fmt.Errorf("request failed, try again")
}
