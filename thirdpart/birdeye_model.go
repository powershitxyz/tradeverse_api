package thirdpart

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type BirdResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type OHLCVResp struct {
	O        decimal.Decimal `json:"o"`
	H        decimal.Decimal `json:"h"`
	C        decimal.Decimal `json:"c"`
	L        decimal.Decimal `json:"l"`
	V        decimal.Decimal `json:"v"`
	UnixTime uint64          `json:"unixTime"`
	Address  string          `json:"address"`
	Type     string          `json:"type"`
	Currency string          `json:"currency"`
}

type SortByType string

const (
	Liquidity          SortByType = "liquidity"
	MarketCap          SortByType = "market_cap"
	FDV                SortByType = "fdv"
	RecentListingTime  SortByType = "recent_listing_time"
	Holder             SortByType = "holder"
	Volume1HUsd        SortByType = "volume_1h_usd"
	Volume2HUsd        SortByType = "volume_2h_usd"
	Volume4HUsd        SortByType = "volume_4h_usd"
	Volume8HUsd        SortByType = "volume_8h_usd"
	Volume24HUsd       SortByType = "volume_24h_usd"
	Volume1HChangePct  SortByType = "volume_1h_change_percent"
	Volume2HChangePct  SortByType = "volume_2h_change_percent"
	Volume4HChangePct  SortByType = "volume_4h_change_percent"
	Volume8HChangePct  SortByType = "volume_8h_change_percent"
	Volume24HChangePct SortByType = "volume_24h_change_percent"
	PriceChange1HPct   SortByType = "price_change_1h_percent"
	PriceChange2HPct   SortByType = "price_change_2h_percent"
	PriceChange4HPct   SortByType = "price_change_4h_percent"
	PriceChange8HPct   SortByType = "price_change_8h_percent"
	PriceChange24HPct  SortByType = "price_change_24h_percent"
	Trade1HCount       SortByType = "trade_1h_count"
	Trade2HCount       SortByType = "trade_2h_count"
	Trade4HCount       SortByType = "trade_4h_count"
	Trade8HCount       SortByType = "trade_8h_count"
	Trade24HCount      SortByType = "trade_24h_count"
)

func (s SortByType) IsValid() bool {
	switch s {
	case Liquidity, MarketCap, FDV, RecentListingTime, Holder,
		Volume1HUsd, Volume2HUsd, Volume4HUsd, Volume8HUsd, Volume24HUsd,
		Volume1HChangePct, Volume2HChangePct, Volume4HChangePct, Volume8HChangePct, Volume24HChangePct,
		PriceChange1HPct, PriceChange2HPct, PriceChange4HPct, PriceChange8HPct, PriceChange24HPct,
		Trade1HCount, Trade2HCount, Trade4HCount, Trade8HCount, Trade24HCount:
		return true
	default:
		return false
	}
}

type TokenListV3Request struct {
	SortBy                SortByType `json:"sort_by"`
	SortType              string     `json:"sort_type"`
	MinLiquidity          *float64   `json:"min_liquidity,omitempty"`
	MaxLiquidity          *float64   `json:"max_liquidity,omitempty"`
	MinMarketCap          *float64   `json:"min_market_cap,omitempty"`
	MaxMarketCap          *float64   `json:"max_market_cap,omitempty"`
	MinFDV                *float64   `json:"min_fdv,omitempty"`
	MaxFDV                *float64   `json:"max_fdv,omitempty"`
	MinRecentListingTime  *int64     `json:"min_recent_listing_time,omitempty"`
	MaxRecentListingTime  *int64     `json:"max_recent_listing_time,omitempty"`
	MinHolder             *int64     `json:"min_holder,omitempty"`
	MinVolume1HChangePct  *float64   `json:"min_volume_1h_change_percent,omitempty"`
	MinVolume2HChangePct  *float64   `json:"min_volume_2h_change_percent,omitempty"`
	MinVolume4HChangePct  *float64   `json:"min_volume_4h_change_percent,omitempty"`
	MinVolume8HChangePct  *float64   `json:"min_volume_8h_change_percent,omitempty"`
	MinVolume24HChangePct *float64   `json:"min_volume_24h_change_percent,omitempty"`
	MinPriceChange1HPct   *float64   `json:"min_price_change_1h_percent,omitempty"`
	MinPriceChange2HPct   *float64   `json:"min_price_change_2h_percent,omitempty"`
	MinPriceChange4HPct   *float64   `json:"min_price_change_4h_percent,omitempty"`
	MinPriceChange8HPct   *float64   `json:"min_price_change_8h_percent,omitempty"`
	MinPriceChange24HPct  *float64   `json:"min_price_change_24h_percent,omitempty"`
	MinTrade1HCount       *int64     `json:"min_trade_1h_count,omitempty"`
	MinTrade2HCount       *int64     `json:"min_trade_2h_count,omitempty"`
	MinTrade4HCount       *int64     `json:"min_trade_4h_count,omitempty"`
	MinTrade8HCount       *int64     `json:"min_trade_8h_count,omitempty"`
	MinTrade24HCount      *int64     `json:"min_trade_24h_count,omitempty"`
}

func (r *TokenListV3Request) Validate() error {
	if !r.SortBy.IsValid() {
		return fmt.Errorf("invalid sort_by value: %s", r.SortBy)
	}
	if r.SortType != "asc" && r.SortType != "desc" {
		return errors.New("sort_type must be either 'asc' or 'desc'")
	}
	return nil
}

type TokenListV3Resp struct {
	Address                string                 `json:"address"`
	Name                   string                 `json:"name"`
	Symbol                 string                 `json:"symbol"`
	Decimals               uint8                  `json:"decimals"`
	LogoURI                string                 `json:"logo_uri"`
	MarketCap              float64                `json:"market_cap"`
	FDV                    float64                `json:"fdv"`
	Liquidity              float64                `json:"liquidity"`
	Price                  float64                `json:"price"`
	Holder                 int                    `json:"holder"`
	Volume1HUsd            float64                `json:"volume_1h_usd"`
	Volume1HChangePercent  float64                `json:"volume_1h_change_percent"`
	PriceChange1HPercent   float64                `json:"price_change_1h_percent"`
	Trade1HCount           int                    `json:"trade_1h_count"`
	Volume2HUsd            float64                `json:"volume_2h_usd"`
	Volume2HChangePercent  float64                `json:"volume_2h_change_percent"`
	PriceChange2HPercent   float64                `json:"price_change_2h_percent"`
	Trade2HCount           int                    `json:"trade_2h_count"`
	Volume4HUsd            float64                `json:"volume_4h_usd"`
	Volume4HChangePercent  float64                `json:"volume_4h_change_percent"`
	PriceChange4HPercent   float64                `json:"price_change_4h_percent"`
	Trade4HCount           int                    `json:"trade_4h_count"`
	Volume8HUsd            float64                `json:"volume_8h_usd"`
	Volume8HChangePercent  float64                `json:"volume_8h_change_percent"`
	PriceChange8HPercent   float64                `json:"price_change_8h_percent"`
	Trade8HCount           int                    `json:"trade_8h_count"`
	Volume24HUsd           float64                `json:"volume_24h_usd"`
	Volume24HChangePercent float64                `json:"volume_24h_change_percent"`
	PriceChange24HPercent  float64                `json:"price_change_24h_percent"`
	Trade24HCount          int                    `json:"trade_24h_count"`
	RecentListingTime      int64                  `json:"recent_listing_time"`
	LastTradeUnixTime      uint64                 `json:"last_trade_unix_time"`
	Extension              map[string]interface{} `json:"extensions"`
}

type TokenHolderResp struct {
	Amount       string          `json:"amount"`
	Decimals     uint8           `json:"decimals"`
	Mint         string          `json:"mint"`
	Owner        string          `json:"owner"`
	TokenAccount string          `json:"token_account"`
	UiAmount     decimal.Decimal `json:"ui_amount"`
}

type TokenSearchRequest struct {
	Chain       string `json:"chain"`
	Keyword     string `json:"keyword"`
	Target      string `json:"target"`
	SearchMode  string `json:"search_mode"`
	SearchBy    string `json:"search_by"`
	SortBy      string `json:"sort_by"`
	SortType    string `json:"sort_type"`
	VerifyToken string `json:"verify_token"`
	Markets     string `json:"markets"`
}

type TokenSearchTokenResp struct {
	Name                         string          `json:"name"`
	Symbol                       string          `json:"symbol"`
	Address                      string          `json:"address"`
	Network                      string          `json:"network"`
	Decimals                     int             `json:"decimals"`
	LogoURI                      string          `json:"logo_uri"`
	Verified                     bool            `json:"verified"`
	FDV                          decimal.Decimal `json:"fdv"`
	MarketCap                    decimal.Decimal `json:"market_cap"`
	Liquidity                    decimal.Decimal `json:"liquidity"`
	Price                        decimal.Decimal `json:"price"`
	PriceChange24hPercent        decimal.Decimal `json:"price_change_24h_percent"`
	Sell24h                      int             `json:"sell_24h"`
	Sell24hChangePercent         decimal.Decimal `json:"sell_24h_change_percent"`
	Buy24h                       int             `json:"buy_24h"`
	Buy24hChangePercent          decimal.Decimal `json:"buy_24h_change_percent"`
	UniqueWallet24h              int             `json:"unique_wallet_24h"`
	UniqueWallet24hChangePercent decimal.Decimal `json:"unique_wallet_24h_change_percent"`
	Trade24h                     int             `json:"trade_24h"`
	Trade24hChangePercent        decimal.Decimal `json:"trade_24h_change_percent"`
	Volume24hChangePercent       decimal.Decimal `json:"volume_24h_change_percent"`
	Volume24hUSD                 decimal.Decimal `json:"volume_24h_usd"`
	LastTradeUnixTime            int64           `json:"last_trade_unix_time"`
	LastTradeHumanTime           string          `json:"last_trade_human_time"`
	Supply                       decimal.Decimal `json:"supply"`
	UpdatedTime                  int64           `json:"updated_time"`
}

// TokenSearchMarketResp 结构体表示市场交易对的信息
type TokenSearchMarketResp struct {
	Name                         string          `json:"name"`
	Address                      string          `json:"address"`
	Network                      string          `json:"network"`
	Liquidity                    decimal.Decimal `json:"liquidity"`
	Source                       string          `json:"source"`
	Trade24h                     int             `json:"trade_24h"`
	Trade24hChangePercent        decimal.Decimal `json:"trade_24h_change_percent"`
	UniqueWallet24h              int             `json:"unique_wallet_24h"`
	UniqueWallet24hChangePercent decimal.Decimal `json:"unique_wallet_24h_change_percent"`
	LastTradeHumanTime           time.Time       `json:"last_trade_human_time"`
	LastTradeUnixTime            int64           `json:"last_trade_unix_time"`
	BaseMint                     string          `json:"base_mint"`
	QuoteMint                    string          `json:"quote_mint"`
	AmountBase                   decimal.Decimal `json:"amount_base"`
	AmountQuote                  decimal.Decimal `json:"amout_quote"`
	CreationTime                 time.Time       `json:"creation_time"`
	Volume24hUSD                 decimal.Decimal `json:"volume_24h_usd"`
}

// TokenSearchCategory 结构体表示 "type" 和对应的结果列表
type TokenSearchCategory struct {
	Type   string          `json:"type"`
	Result json.RawMessage `json:"result"`
}

type TokenMeta struct {
	Address    string                 `json:"address"`
	Symbol     string                 `json:"symbol"`
	Name       string                 `json:"name"`
	Decimals   uint8                  `json:"decimals"`
	LogoUri    string                 `json:"logo_uri"`
	Extensions map[string]interface{} `json:"extensions"`
}

type WalletInfo struct {
	Wallet   string        `json:"wallet"`
	TotalUsd float64       `json:"totalUsd"`
	Items    []WalletToken `json:"items"`
}

type WalletToken struct {
	Address  string          `json:"address"`
	Decimals int             `json:"decimals"`
	Balance  int64           `json:"balance"`
	UiAmount decimal.Decimal `json:"uiAmount"`
	ChainID  string          `json:"chainId"`
	Name     string          `json:"name,omitempty"`
	Symbol   string          `json:"symbol,omitempty"`
	Icon     string          `json:"icon,omitempty"`
	LogoURI  string          `json:"logoURI,omitempty"`
	PriceUsd decimal.Decimal `json:"priceUsd,omitempty"`
	ValueUsd decimal.Decimal `json:"valueUsd,omitempty"`
}

type NewToken struct {
	Address        string          `json:"address"`
	Symbol         string          `json:"symbol"`
	Name           string          `json:"name"`
	Decimals       uint8           `json:"decimals"`
	Source         string          `json:"source"`
	LiquidityAddAt string          `json:"liquidityAddedAt"`
	Logo           string          `json:"logoURI"`
	Liquidity      decimal.Decimal `json:"liquidity"`
}
