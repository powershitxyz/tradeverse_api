package thirdpart

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const yahooFinanceChartBase = "https://query1.finance.yahoo.com/v8/finance/chart"

// YahooChartResponse 对应 Yahoo Finance Chart API 的 JSON 结构
type YahooChartResponse struct {
	Chart *struct {
		Result []*YahooChartResult `json:"result"`
		Error  interface{}         `json:"error"`
	} `json:"chart"`
}

// YahooChartResult 单只标的的 K 线结果
type YahooChartResult struct {
	Meta struct {
		Currency           string   `json:"currency"`
		Symbol             string   `json:"symbol"`
		ExchangeName       string   `json:"exchangeName"`
		RegularMarketPrice float64  `json:"regularMarketPrice"`
		ChartPreviousClose float64  `json:"chartPreviousClose"`
		DataGranularity    string   `json:"dataGranularity"`
		ValidRanges        []string `json:"validRanges"`
	} `json:"meta"`
	Timestamp  []int64 `json:"timestamp"`
	Indicators *struct {
		Quote []*struct {
			Open   []*float64 `json:"open"`
			High   []*float64 `json:"high"`
			Low    []*float64 `json:"low"`
			Close  []*float64 `json:"close"`
			Volume []*int64   `json:"volume"`
		} `json:"quote"`
	} `json:"indicators"`
}

// KlineItem 对外返回的单根 K 线
type KlineItem struct {
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    int64   `json:"volume"`
}

// ChartMeta 图表元信息
type ChartMeta struct {
	Symbol             string   `json:"symbol"`
	Currency           string   `json:"currency"`
	ExchangeName       string   `json:"exchangeName"`
	RegularMarketPrice float64  `json:"regularMarketPrice"`
	DataGranularity    string   `json:"dataGranularity"`
	ValidRanges        []string `json:"validRanges"`
}

// ChartResult 对外统一的 K 线结果
type ChartResult struct {
	Meta ChartMeta   `json:"meta"`
	Data []KlineItem `json:"data"`
}

var yahooHTTPClient = &http.Client{Timeout: 15 * time.Second}

// GetChart 请求 Yahoo Finance Chart API，获取 K 线
// interval: 如 1m, 5m, 15m, 1h, 1d 等；range: 如 1d, 5d, 1mo, 3mo, 6mo, 1y 等
func GetChart(symbol, interval, rangeParam string) (*ChartResult, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if interval == "" {
		interval = "1h"
	}
	if rangeParam == "" {
		rangeParam = "1mo"
	}

	u, err := url.Parse(yahooFinanceChartBase + "/" + url.PathEscape(symbol))
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("interval", interval)
	q.Set("range", rangeParam)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Tradeverse/1.0)")

	resp, err := yahooHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo chart api status: %d", resp.StatusCode)
	}

	var raw YahooChartResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if raw.Chart == nil || len(raw.Chart.Result) == 0 {
		return nil, fmt.Errorf("no chart result for symbol %s", symbol)
	}

	r := raw.Chart.Result[0]
	out := &ChartResult{
		Meta: ChartMeta{
			Symbol:             r.Meta.Symbol,
			Currency:           r.Meta.Currency,
			ExchangeName:       r.Meta.ExchangeName,
			RegularMarketPrice: r.Meta.RegularMarketPrice,
			DataGranularity:    r.Meta.DataGranularity,
			ValidRanges:        r.Meta.ValidRanges,
		},
		Data: nil,
	}

	if r.Indicators == nil || len(r.Indicators.Quote) == 0 {
		return out, nil
	}

	quote := r.Indicators.Quote[0]
	ts := r.Timestamp
	open := quote.Open
	high := quote.High
	low := quote.Low
	close := quote.Close
	vol := quote.Volume

	n := len(ts)
	if n == 0 {
		return out, nil
	}
	out.Data = make([]KlineItem, 0, n)

	for i := 0; i < n; i++ {
		item := KlineItem{Timestamp: ts[i]}
		if i < len(open) && open[i] != nil {
			item.Open = *open[i]
		}
		if i < len(high) && high[i] != nil {
			item.High = *high[i]
		}
		if i < len(low) && low[i] != nil {
			item.Low = *low[i]
		}
		if i < len(close) && close[i] != nil {
			item.Close = *close[i]
		}
		if i < len(vol) && vol[i] != nil {
			item.Volume = *vol[i]
		}
		out.Data = append(out.Data, item)
	}

	return out, nil
}
