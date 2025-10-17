package thirdpart

import (
	"fmt"
	"testing"
	"time"

	"chaos/api/api/common"
	"chaos/api/tools"
)

func TestK(t *testing.T) {
	b, err := GetBirdClient()
	t.Log(b, err)

	result, _ := b.FetchKlineData("So11111111111111111111111111111111111111112", "solana", "15m", "usd", 1741606785, 1741693185)
	t.Log(result)

	tq := TokenListV3Request{
		SortBy:       Trade1HCount,
		SortType:     "desc",
		MinLiquidity: tools.Float64Ptr(100),
	}
	q := common.QueryParams[TokenListV3Request]{
		Data: tq,
	}

	s := q.BuildQueryString()
	fmt.Println(s)

	r, _ := b.TokenTrending("solana", tq, 1, 2)
	t.Log(r)
}

func TestHolder(t *testing.T) {
	b, err := GetBirdClient()
	t.Log(b, err)

	r, _ := b.TokenHolder("solana", "So11111111111111111111111111111111111111112", 0, 100)
	t.Log(r)
}

func TestSearch(t *testing.T) {
	b, err := GetBirdClient()
	t.Log(b, err)

	ts := TokenSearchRequest{
		Chain:       "solana",
		Keyword:     "stonks",
		Target:      "all",
		SearchMode:  "fuzzy",
		SearchBy:    "symbol",
		SortBy:      "marketcap",
		SortType:    "desc",
		VerifyToken: "true",
	}

	r, _, _ := b.TokenSearch(ts, 0, 20)
	t.Log(r)

	rs, _ := b.TokenMetas("solana", "So11111111111111111111111111111111111111112,mSoLzYCxHdYgdzU16g5QSh3i5K3z3KZK7ytfqcJm7So")
	t.Log(rs)

	rss, _ := b.TokenNewList("solana", uint64(time.Now().Unix()), true, 20)
	t.Log(rss)
}
