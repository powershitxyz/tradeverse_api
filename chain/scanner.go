package chain

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/*************** 配置与状态 ***************/
type TailScannerConfig struct {
	RPCURL     string   // 必填：HTTP RPC
	Wallets    []string // 必填：要监控的钱包地址集合
	TargetMint string   // 必填：要检测买/卖的 SPL mint
	StartSlot  uint64   // 从这个 slot 之后开始扫描（含 startSlot+1）
	StatePath  string   // 必填：进度文件，例如 "state.json"
	CSVPath    string   // 必填：结果 CSV，例如 "trades_tail.csv"

	FlushEvery   time.Duration // 可选：CSV flush 周期，默认 2s
	SaveEveryN   uint64        // 可选：每处理 N 个 slot 也保存一次进度，默认 100
	HeadIdleWait time.Duration // 可选：追到最新时的等待时间，默认 1s
}

type tailState struct {
	LastProcessed uint64 `json:"last_processed_slot"`
}

/*************** 运行入口（可用 goroutine 启动） ***************/
func RunTailScanner(ctx context.Context, cfg TailScannerConfig) error {
	if cfg.RPCURL == "" || len(cfg.Wallets) == 0 || cfg.TargetMint == "" ||
		cfg.StatePath == "" || cfg.CSVPath == "" {
		return fmt.Errorf("invalid config")
	}
	if cfg.FlushEvery == 0 {
		cfg.FlushEvery = 2 * time.Second
	}
	if cfg.SaveEveryN == 0 {
		cfg.SaveEveryN = 100
	}
	if cfg.HeadIdleWait == 0 {
		cfg.HeadIdleWait = 1 * time.Second
	}

	// 钱包集合
	wSet := make(map[string]struct{}, len(cfg.Wallets))
	for _, w := range cfg.Wallets {
		wSet[w] = struct{}{}
	}

	// 进度
	st, _ := loadState(cfg.StatePath)
	if st.LastProcessed < cfg.StartSlot {
		st.LastProcessed = cfg.StartSlot
	}
	fmt.Printf("[scanner] resume from slot %d\n", st.LastProcessed)

	// CSV append
	writer, file, err := openCSVAppend(cfg.CSVPath)
	if err != nil {
		return err
	}
	defer file.Close()

	flushTicker := time.NewTicker(cfg.FlushEvery)
	defer flushTicker.Stop()

	// 主循环（可长期运行）
	for {
		select {
		case <-ctx.Done():
			writer.Flush()
			_ = writer.Error()
			_ = saveState(cfg.StatePath, st)
			return ctx.Err()
		default:
		}

		// 最新 head
		head, err := getSlot(cfg.RPCURL)
		if err != nil {
			fmt.Println("[scanner] getSlot err:", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if st.LastProcessed >= head {
			time.Sleep(cfg.HeadIdleWait)
			continue
		}

		// 逐 slot 扫描
		begin := st.LastProcessed + 1
		for slot := begin; slot <= head; slot++ {
			select {
			case <-ctx.Done():
				writer.Flush()
				_ = writer.Error()
				_ = saveState(cfg.StatePath, st)
				return ctx.Err()
			default:
			}

			br, err := getBlock(cfg.RPCURL, slot)
			if err != nil {
				fmt.Printf("[scanner] getBlock(%d) err: %v\n", slot, err)
				st.LastProcessed = slot
				if slot%cfg.SaveEveryN == 0 {
					_ = saveState(cfg.StatePath, st)
				}
				continue
			}
			if br == nil || len(br.Transactions) == 0 {
				// skipped / 空块
				st.LastProcessed = slot
				if slot%cfg.SaveEveryN == 0 {
					_ = saveState(cfg.StatePath, st)
				}
				continue
			}

			var ts string
			if br.BlockTime != nil {
				ts = time.Unix(*br.BlockTime, 0).UTC().Format(time.RFC3339)
			}

			for _, tx := range br.Transactions {
				if tx.Meta == nil {
					continue
				}
				feeSOL := fmt.Sprintf("%.9f", float64(tx.Meta.Fee)/1e9)
				sig := ""
				if len(tx.Transaction.Signatures) > 0 {
					sig = tx.Transaction.Signatures[0]
				}

				// 先找这笔交易中，跟我们目标 mint 且 owner 在钱包集合的账户
				owners := ownersHit(tx.Meta, cfg.TargetMint, wSet)
				if len(owners) == 0 {
					continue
				}
				for owner := range owners {
					delta, dec, ok := calcDeltaFor(owner, cfg.TargetMint, tx.Meta)
					if !ok || delta.Sign() == 0 {
						continue
					}
					dir := "BUY"
					if delta.Sign() < 0 {
						dir = "SELL"
					}
					deltaAbs := new(big.Rat).Abs(delta)
					// 用代币 decimals 打印，避免精度损失
					deltaStr := deltaAbs.FloatString(dec)
					okStr := "true"
					if tx.Meta.Err != nil {
						okStr = "false"
					}

					_ = writer.Write([]string{
						fmt.Sprint(slot), ts, sig, owner, dir, deltaStr, cfg.TargetMint, feeSOL, okStr,
					})
				}
			}

			// 定期 flush + 保存进度
			st.LastProcessed = slot
			select {
			case <-flushTicker.C:
				writer.Flush()
				_ = writer.Error()
				_ = saveState(cfg.StatePath, st)
			default:
			}
		}

		// 扫到最新再落一次
		writer.Flush()
		_ = writer.Error()
		_ = saveState(cfg.StatePath, st)
	}
}

/*************** RPC 与数据类型 ***************/
type rpcReq struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}
type rpcErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type rpcResp[T any] struct {
	Jsonrpc string  `json:"jsonrpc"`
	ID      int     `json:"id"`
	Result  T       `json:"result"`
	Error   *rpcErr `json:"error,omitempty"`
}

func rpcCall[T any](rpcURL, method string, params []interface{}, out *T) error {
	body, _ := json.Marshal(rpcReq{"2.0", 1, method, params})
	resp, err := http.Post(rpcURL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var wrap rpcResp[T]
	if err := json.NewDecoder(resp.Body).Decode(&wrap); err != nil {
		return err
	}
	if wrap.Error != nil {
		return fmt.Errorf("rpc %s error %d: %s", method, wrap.Error.Code, wrap.Error.Message)
	}
	*out = wrap.Result
	return nil
}

func getSlot(rpcURL string) (uint64, error) {
	var out uint64
	return out, rpcCall(rpcURL, "getSlot", []interface{}{}, &out)
}

type tokenAmt struct {
	Amount   string `json:"amount"`
	Decimals int    `json:"decimals"`
}
type tokBal struct {
	Owner        string    `json:"owner"`
	Mint         string    `json:"mint"`
	UIAmountInfo *tokenAmt `json:"uiTokenAmount"`
}
type txMeta struct {
	PreTokenBalances  []tokBal `json:"preTokenBalances"`
	PostTokenBalances []tokBal `json:"postTokenBalances"`
	Fee               uint64   `json:"fee"`
	Err               any      `json:"err"`
}
type txEnc struct {
	Signatures []string `json:"signatures"`
}
type blockTx struct {
	Transaction txEnc   `json:"transaction"`
	Meta        *txMeta `json:"meta"`
}
type blockResult struct {
	BlockTime    *int64    `json:"blockTime"`
	Transactions []blockTx `json:"transactions"`
}

func getBlock(rpcURL string, slot uint64) (*blockResult, error) {
	var out *blockResult
	err := rpcCall(rpcURL, "getBlock", []interface{}{
		slot,
		map[string]any{
			"encoding":                       "jsonParsed",
			"maxSupportedTransactionVersion": 0,
			"transactionDetails":             "full",
			"rewards":                        false,
		},
	}, &out)
	// 某些 slot 返回 null；按 nil 处理
	return out, err
}

/*************** 余额差与命中判断 ***************/
func amtToRat(a string, dec int) *big.Rat {
	n := new(big.Int)
	n.SetString(a, 10)
	den := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(dec)), nil)
	return new(big.Rat).SetFrac(n, den)
}
func calcDeltaFor(owner, mint string, m *txMeta) (*big.Rat, int, bool) {
	if m == nil {
		return nil, 0, false
	}
	var pre, post *big.Rat
	var dp, dq int
	found := false

	for _, b := range m.PreTokenBalances {
		if b.Owner == owner && b.Mint == mint && b.UIAmountInfo != nil {
			pre = amtToRat(b.UIAmountInfo.Amount, b.UIAmountInfo.Decimals)
			dp = b.UIAmountInfo.Decimals
			found = true
			break
		}
	}
	for _, b := range m.PostTokenBalances {
		if b.Owner == owner && b.Mint == mint && b.UIAmountInfo != nil {
			post = amtToRat(b.UIAmountInfo.Amount, b.UIAmountInfo.Decimals)
			dq = b.UIAmountInfo.Decimals
			found = true
			break
		}
	}
	dec := dp
	if post != nil {
		dec = dq
	}
	if !found {
		return nil, 0, false
	}
	if pre == nil {
		pre = new(big.Rat) // 0
	}
	if post == nil {
		post = new(big.Rat)
	}
	return new(big.Rat).Sub(post, pre), dec, true
}

func ownersHit(m *txMeta, mint string, wset map[string]struct{}) map[string]struct{} {
	hit := make(map[string]struct{})
	for _, b := range m.PostTokenBalances {
		if b.Mint == mint {
			if _, ok := wset[b.Owner]; ok {
				hit[b.Owner] = struct{}{}
			}
		}
	}
	for _, b := range m.PreTokenBalances {
		if b.Mint == mint {
			if _, ok := wset[b.Owner]; ok {
				hit[b.Owner] = struct{}{}
			}
		}
	}
	return hit
}

/*************** CSV 与状态落盘 ***************/
func openCSVAppend(path string) (*csv.Writer, *os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil && !os.IsExist(err) {
		return nil, nil, err
	}
	_, statErr := os.Stat(path)
	isNew := os.IsNotExist(statErr)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, err
	}
	w := csv.NewWriter(f)
	if isNew {
		_ = w.Write([]string{"slot", "time(utc)", "signature", "wallet", "direction", "delta", "mint", "fee(SOL)", "ok"})
		w.Flush()
	}
	return w, f, nil
}

func loadState(path string) (tailState, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return tailState{}, err
	}
	var s tailState
	if err := json.Unmarshal(b, &s); err != nil {
		return tailState{}, err
	}
	return s, nil
}

func saveState(path string, s tailState) error {
	tmp := path + ".tmp"
	b, _ := json.MarshalIndent(s, "", "  ")
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path) // 原子替换
}
