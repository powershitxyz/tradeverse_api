package mycache

import (
	"time"

	"chaos/api/thirdpart"

	"github.com/dgraph-io/ristretto/v2"
)

const chartCacheTTL = 15 * time.Minute

var ChartCache *ristretto.Cache[string, *thirdpart.ChartResult]

func init() {
	cache, err := ristretto.NewCache[string, *thirdpart.ChartResult](&ristretto.Config[string, *thirdpart.ChartResult]{
		NumCounters: 10000,
		MaxCost:     50 * 1024 * 1024, // 50MB 约可存数百条 K 线结果
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	ChartCache = cache
}

func chartCacheKey(symbol, interval, rangeParam string) string {
	return symbol + "|" + interval + "|" + rangeParam
}

// GetChart 从缓存读取 K 线结果，ok 表示命中
func GetChart(symbol, interval, rangeParam string) (*thirdpart.ChartResult, bool) {
	ChartCache.Wait()
	return ChartCache.Get(chartCacheKey(symbol, interval, rangeParam))
}

// SetChart 写入 K 线结果到缓存，TTL 15 分钟
func SetChart(symbol, interval, rangeParam string, chart *thirdpart.ChartResult) {
	if chart == nil {
		return
	}
	key := chartCacheKey(symbol, interval, rangeParam)
	cost := int64(1)
	if len(chart.Data) > 0 {
		cost = int64(len(chart.Data))
	}
	ChartCache.SetWithTTL(key, chart, cost, chartCacheTTL)
	ChartCache.Wait()
}
