package chain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// TxMintDelta 表示在一笔交易中，指定 mint 对某个 Owner 的数量变动（正为增加，负为减少），
// Delta 已按 token 的 decimals 转为十进制字符串表示，便于直接展示或记录。
type TxMintDelta struct {
	Owner    string `json:"owner"`
	Delta    string `json:"delta"` // 十进制字符串，带符号，按 decimals 格式化
	Decimals int    `json:"decimals"`
}

// TxMintAnalysis 是对某笔交易中某个 mint 的解析结果。
type TxMintAnalysis struct {
	Signature  string        `json:"signature"`
	Success    bool          `json:"success"` // true 表示 meta.err == nil
	Slot       uint64        `json:"slot"`
	BlockTime  *time.Time    `json:"block_time,omitempty"`
	Mint       string        `json:"mint"`
	Deltas     []TxMintDelta `json:"deltas"` // 仅包含 delta != 0 的账户
	FeeLamport uint64        `json:"fee_lamport"`
}

// getTransactionResult 只取我们关心的字段（encoding=jsonParsed）。
type getTransactionResult struct {
	Slot      uint64  `json:"slot"`
	BlockTime *int64  `json:"blockTime"`
	Meta      *txMeta `json:"meta"`
}

// getSignatureStatuses 精简返回
type signatureStatuses struct {
	Value []struct {
		Slot               uint64  `json:"slot"`
		Err                any     `json:"err"`
		ConfirmationStatus *string `json:"confirmationStatus"`
	} `json:"value"`
}

// AnalyzeTxMintDelta 解析指定交易（signature）中某个 SPL mint 的数量变化，并给出交易状态。
// - rpcURL: Solana HTTP RPC 地址
// - signature: 交易签名（hash）
// - mint: 目标 SPL mint 地址
func AnalyzeTxMintDelta(ctx context.Context, rpcURL, signature, mint string) (*TxMintAnalysis, error) {
	if rpcURL == "" || signature == "" || mint == "" {
		return nil, fmt.Errorf("invalid input")
	}

	// 通过 getTransaction 拉取明细（使用 jsonParsed 以便直接读取 tokenBalances）
	var out *getTransactionResult
	err := rpcCall(rpcURL, "getTransaction", []interface{}{
		signature,
		map[string]any{
			"encoding":                       "jsonParsed",
			"maxSupportedTransactionVersion": 0,
			"commitment":                     "confirmed",
		},
	}, &out)
	if err != nil {
		return nil, errors.New("rpc_call_error:" + err.Error())
	}
	if out == nil || out.Meta == nil {
		// 非归档节点会返回 result=null 且无 error。回退到 getSignatureStatuses 给出至少的状态信息。
		var st signatureStatuses
		_ = rpcCall(rpcURL, "getSignatureStatuses", []interface{}{[]string{signature}, map[string]any{"searchTransactionHistory": true}}, &st)

		var slot uint64
		var success bool
		if len(st.Value) > 0 {
			slot = st.Value[0].Slot
			success = (st.Value[0].Err == nil)
		}
		return &TxMintAnalysis{
			Signature:  signature,
			Success:    success,
			Slot:       slot,
			BlockTime:  nil,
			Mint:       mint,
			Deltas:     nil,
			FeeLamport: 0,
		}, fmt.Errorf("data_pruned_error")
	}

	// 收集该 mint 相关的 owner 集合
	ownerSet := make(map[string]struct{})
	for _, b := range out.Meta.PostTokenBalances {
		if b.Mint == mint && b.Owner != "" {
			ownerSet[b.Owner] = struct{}{}
		}
	}
	for _, b := range out.Meta.PreTokenBalances {
		if b.Mint == mint && b.Owner != "" {
			ownerSet[b.Owner] = struct{}{}
		}
	}

	var deltas []TxMintDelta
	for owner := range ownerSet {
		rat, dec, ok := calcDeltaFor(owner, mint, out.Meta)
		if !ok || rat == nil || rat.Sign() == 0 {
			continue
		}
		deltas = append(deltas, TxMintDelta{
			Owner:    owner,
			Delta:    rat.FloatString(dec),
			Decimals: dec,
		})
	}

	var tt *time.Time
	if out.BlockTime != nil {
		t := time.Unix(*out.BlockTime, 0).UTC()
		tt = &t
	}

	analysis := &TxMintAnalysis{
		Signature:  signature,
		Success:    out.Meta.Err == nil,
		Slot:       out.Slot,
		BlockTime:  tt,
		Mint:       mint,
		Deltas:     deltas,
		FeeLamport: out.Meta.Fee,
	}
	return analysis, nil
}
