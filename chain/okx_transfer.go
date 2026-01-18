package chain

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var chainIndexMap = map[string]string{
	"BTC":                 "0",
	"BITCOIN":             "0",
	"ETH":                 "1",
	"ETHEREUM":            "1",
	"ETHEREUM_ETH":        "1",
	"UNICHAIN":            "130",
	"UNICHAIN_ETH":        "130",
	"BASE":                "8453",
	"BASE_ETH":            "8453",
	"CFX":                 "1030",
	"CONFLUX":             "1030",
	"CONFLUX_ESPACE":      "1030",
	"LINEA":               "59144",
	"LINEA_ETH":           "59144",
	"MANTLE":              "5000",
	"MNT":                 "5000",
	"BOB":                 "60808",
	"BOB_ETH":             "60808",
	"BOB_MAINNET":         "60808",
	"POLYGON":             "137",
	"POLYGON_MATIC":       "137",
	"MATIC":               "137",
	"OPTIMISM":            "10",
	"OP":                  "10",
	"MODE":                "34443",
	"MODE_ETH":            "34443",
	"MONAD":               "143",
	"MON":                 "143",
	"SUI":                 "784",
	"ETHW":                "10001",
	"ETHEREUMPOW":         "10001",
	"ETHW_POW":            "10001",
	"PLASMA":              "9745",
	"XPL":                 "9745",
	"SONIC":               "146",
	"S":                   "146",
	"SONIC_MAINNET":       "146",
	"CRONOS":              "25",
	"CRO":                 "25",
	"MANTA":               "169",
	"MANTA_ETH":           "169",
	"MANTA_PACIFIC":       "169",
	"STARKNET":            "9004",
	"STARKNET_ETH":        "9004",
	"BLAST":               "81457",
	"ARBITRUM":            "42161",
	"ARB":                 "42161",
	"ARBITRUM_ONE":        "42161",
	"SEI":                 "70000029",
	"SEIEVM":              "1329",
	"EVM_SEI":             "1329",
	"BSC":                 "56",
	"BNB":                 "56",
	"BNB_CHAIN":           "56",
	"BNBCHAIN":            "56",
	"BINANCE":             "56",
	"BINANCE_SMART_CHAIN": "56",
	"BINANCE_CHAIN":       "56",
	"CHILIZ":              "88888",
	"CHZ":                 "88888",
	"CHILIZ_CHAIN":        "88888",
	"IMMUTABLE":           "13371",
	"IMX":                 "13371",
	"IMMUTABLE_ZKEVM":     "13371",
	"METIS":               "1088",
	"METIS_ANDROMEDA":     "1088",
	"OSMOSIS":             "706",
	"OSMO":                "706",
	"TRON":                "195",
	"TRX":                 "195",
	"ZKSYNC":              "324",
	"ERA_ETH":             "324",
	"ZK_SYNC_ERA":         "324",
	"ZKSYNC_ERA":          "324",
	"XLAYER":              "196",
	"OKB":                 "196",
	"OKX_CHAIN":           "196",
	"BITLAYER":            "200901",
	"BITLAYER_BTC":        "200901",
	"OPBNB":               "204",
	"OP_BNB":              "204",
	"OPBNB_CHAIN":         "204",
	"POLYGON_ZKEVM":       "1101",
	"POLYGON_ETH":         "1101",
	"POLYGON_ZKEVM_ETH":   "1101",
	"IOTEX":               "4689",
	"IOTX":                "4689",
	"IOTEX_NETWORK":       "4689",
	"SCROLL":              "534352",
	"SCROLL_ETH":          "534352",
	"SCROLL_L2":           "534352",
	"ZETA":                "7000",
	"ZETACHAIN":           "7000",
	"TAIKO":               "167000",
	"TAIKO_ETH":           "167000",
	"TAIKO_MAINNET":       "167000",
	"STABLE":              "988",
	"GUSDT":               "988",
	"STABLE_COIN":         "988",
	"TON":                 "607",
	"TONCHAIN":            "607",
	"THE_OPEN_NETWORK":    "607",
	"B2":                  "223",
	"B2_BTC":              "223",
	"B2_NETWORK":          "223",
	"BERACHAIN":           "80094",
	"BERA":                "80094",
	"GNOSIS":              "100",
	"XDAI":                "100",
	"GNOSIS_CHAIN":        "100",
	"MERLIN":              "4200",
	"MERLIN_BTC":          "4200",
	"MERLIN_CHAIN":        "4200",
	"AVAX":                "43114",
	"AVALANCHE":           "43114",
	"AVALANCHE_C":         "43114",
	"AVALANCHE_C_CHAIN":   "43114",
	"PULSECHAIN":          "369",
	"PLS":                 "369",
	"PULSE":               "369",
	"APE":                 "33139",
	"APECHAIN":            "33139",
	"APE_CHAIN":           "33139",
	"SOLANA":              "501",
	"SOL":                 "501",
	"FANTOM":              "250",
	"FTM":                 "250",
	"FANTOM_OPERA":        "250",
	"APTOS":               "637",
	"APT":                 "637",
	"STACKS":              "5757",
	"STX":                 "5757",
	"STACKS_BTC":          "5757",
}

type OKXTransferClient struct {
	apiKey     string
	secretKey  string
	passphrase string
	baseURL    string
	httpClient *http.Client
}

func NewOKXTransferClient() (*OKXTransferClient, error) {
	apiKey := os.Getenv("OKX_API_KEY")
	secretKey := os.Getenv("OKX_SECRET_KEY")
	passphrase := os.Getenv("OKX_API_PASSPHRASE")

	if apiKey == "" || secretKey == "" || passphrase == "" {
		return nil, fmt.Errorf("缺少必要的 OKX API 环境变量: OKX_API_KEY, OKX_SECRET_KEY, OKX_API_PASSPHRASE")
	}

	return &OKXTransferClient{
		apiKey:     apiKey,
		secretKey:  secretKey,
		passphrase: passphrase,
		baseURL:    "https://web3.okx.com",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *OKXTransferClient) createSignature(timestamp, method, requestPath, body string) string {
	message := timestamp + method + requestPath + body
	h := hmac.New(sha256.New, []byte(c.secretKey))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (c *OKXTransferClient) doRequest(method, endpoint string, params url.Values, body []byte) ([]byte, error) {
	var requestPath string
	var bodyStr string
	var queryString string

	if method == "GET" {
		requestPath = endpoint
		if params != nil && len(params) > 0 {
			queryString = "?" + params.Encode()
		} else {
			queryString = ""
		}
		bodyStr = ""
	} else {
		requestPath = endpoint
		queryString = ""
		if body != nil {
			bodyStr = string(body)
		} else {
			bodyStr = ""
		}
	}

	var req *http.Request
	var err error
	if method == "POST" && body != nil {
		req, err = http.NewRequest(method, c.baseURL+requestPath+queryString, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, c.baseURL+requestPath+queryString, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("构造 HTTP 请求失败: %v", err)
	}

	now := time.Now().UTC()
	timestamp := now.Format("2006-01-02T15:04:05.000Z")

	signature := c.createSignature(timestamp, method, requestPath+queryString, bodyStr)

	req.Header.Set("OK-ACCESS-KEY", c.apiKey)
	req.Header.Set("OK-ACCESS-SIGN", signature)
	req.Header.Set("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("OK-ACCESS-PASSPHRASE", c.passphrase)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 状态码错误: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

type NormalTransferInfo struct {
	TxHash         string
	From           string
	To             string
	Amount         string
	BlockNumber    string
	BlockTime      string
	Status         string
	ChainIndex     string
	ChainName      string
	Symbol         string
	IsFromContract bool
	IsToContract   bool
	MethodID       string
	TxFee          string
	Nonce          string
	TxValue        string
	TxValueSymbol  string
	TokenAddress   string
}

var chainIndexToNameMap = map[string]string{
	"0":        "BTC",
	"1":        "ETH",
	"130":      "UNICHAIN",
	"8453":     "BASE",
	"1030":     "CFX",
	"59144":    "LINEA",
	"5000":     "MANTLE",
	"60808":    "BOB",
	"137":      "POLYGON",
	"10":       "OPTIMISM",
	"34443":    "MODE",
	"143":      "MONAD",
	"784":      "SUI",
	"10001":    "ETHW",
	"9745":     "PLASMA",
	"146":      "SONIC",
	"25":       "CRONOS",
	"169":      "MANTA",
	"9004":     "STARKNET",
	"81457":    "BLAST",
	"42161":    "ARBITRUM",
	"70000029": "SEI",
	"1329":     "SEIEVM",
	"56":       "BSC",
	"88888":    "CHILIZ",
	"13371":    "IMMUTABLE",
	"1088":     "METIS",
	"706":      "OSMOSIS",
	"195":      "TRON",
	"324":      "ZKSYNC",
	"196":      "XLAYER",
	"200901":   "BITLAYER",
	"204":      "OPBNB",
	"1101":     "POLYGON_ZKEVM",
	"4689":     "IOTEX",
	"534352":   "SCROLL",
	"7000":     "ZETA",
	"167000":   "TAIKO",
	"988":      "STABLE",
	"607":      "TON",
	"223":      "B2",
	"80094":    "BERACHAIN",
	"100":      "GNOSIS",
	"4200":     "MERLIN",
	"43114":    "AVAX",
	"369":      "PULSECHAIN",
	"33139":    "APE",
	"501":      "SOLANA",
	"250":      "FANTOM",
	"637":      "APTOS",
	"5757":     "STACKS",
}

type TransactionDetailResponseData struct {
	ChainIndex  string `json:"chainIndex"`
	Height      string `json:"height"`
	TxTime      string `json:"txTime"`
	Txhash      string `json:"txhash"`
	TxStatus    string `json:"txStatus"`
	MethodID    string `json:"methodId"`
	Nonce       string `json:"nonce"`
	TxFee       string `json:"txFee"`
	FromDetails []struct {
		Address    string `json:"address"`
		VoutIndex  string `json:"voutIndex"`
		IsContract bool   `json:"isContract"`
		Amount     string `json:"amount"`
	} `json:"fromDetails"`
	ToDetails []struct {
		Address    string `json:"address"`
		VoutIndex  string `json:"voutIndex"`
		IsContract bool   `json:"isContract"`
		Amount     string `json:"amount"`
	} `json:"toDetails"`
	InternalTransactionDetails []struct {
		From           string `json:"from"`
		To             string `json:"to"`
		IsFromContract bool   `json:"isFromContract"`
		IsToContract   bool   `json:"isToContract"`
		Amount         string `json:"amount"`
		State          string `json:"state"`
	} `json:"internalTransactionDetails"`
	TokenTransferDetails []struct {
		From          string `json:"from"`
		To            string `json:"to"`
		TokenAddress  string `json:"tokenAddress"`
		TokenSymbol   string `json:"tokenSymbol"`
		TokenDecimals string `json:"tokenDecimals"`
		Amount        string `json:"amount"`
		State         string `json:"state"`
	} `json:"tokenTransferDetails"`
}

type TransactionDetailResponse struct {
	Code string                          `json:"code"`
	Msg  string                          `json:"msg"`
	Data []TransactionDetailResponseData `json:"data"`
}

func parseTransactionDetail(tx TransactionDetailResponseData, chainIndex string) *NormalTransferInfo {
	chainName := getChainNameByIndex(chainIndex)
	info := &NormalTransferInfo{
		TxHash:      tx.Txhash,
		BlockNumber: tx.Height,
		BlockTime:   tx.TxTime,
		Status:      tx.TxStatus,
		ChainIndex:  tx.ChainIndex,
		ChainName:   chainName,
		MethodID:    tx.MethodID,
		Nonce:       tx.Nonce,
		TxFee:       tx.TxFee,
	}

	if len(tx.FromDetails) > 0 {
		info.From = tx.FromDetails[0].Address
		info.IsFromContract = tx.FromDetails[0].IsContract
	}

	if len(tx.ToDetails) > 0 {
		info.To = tx.ToDetails[0].Address
		info.IsToContract = tx.ToDetails[0].IsContract
		if tx.ToDetails[0].Amount != "" {
			amount, err := strconv.ParseFloat(tx.ToDetails[0].Amount, 64)
			if err == nil {
				info.Amount = fmt.Sprintf("%.0f", amount)
				info.TxValue = fmt.Sprintf("%.0f", amount)
			} else {
				info.Amount = tx.ToDetails[0].Amount
				info.TxValue = tx.ToDetails[0].Amount
			}
		}
	}

	if len(tx.InternalTransactionDetails) > 0 {
		internalTx := tx.InternalTransactionDetails[0]
		if info.To == "" {
			info.To = internalTx.To
		}
		if !info.IsFromContract {
			info.IsFromContract = internalTx.IsFromContract
		}
		if !info.IsToContract {
			info.IsToContract = internalTx.IsToContract
		}
		if info.Amount == "" && internalTx.Amount != "" {
			amount, err := strconv.ParseFloat(internalTx.Amount, 64)
			if err == nil {
				info.Amount = fmt.Sprintf("%.0f", amount)
				info.TxValue = fmt.Sprintf("%.0f", amount)
			} else {
				info.Amount = internalTx.Amount
				info.TxValue = internalTx.Amount
			}
		}
	}

	if len(tx.TokenTransferDetails) > 0 {
		tokenTx := tx.TokenTransferDetails[0]
		info.TokenAddress = tokenTx.TokenAddress
		if info.Symbol == "" {
			if chainIndex == "501" && tokenTx.TokenAddress == "So11111111111111111111111111111111111111111" && tokenTx.TokenSymbol == "" {
				info.Symbol = "SOLANA"
			} else {
				info.Symbol = tokenTx.TokenSymbol
			}
		}
		if info.TxValueSymbol == "" {
			if chainIndex == "501" && tokenTx.TokenAddress == "So11111111111111111111111111111111111111111" && tokenTx.TokenSymbol == "" {
				info.TxValueSymbol = "SOLANA"
			} else {
				info.TxValueSymbol = tokenTx.TokenSymbol
			}
		}
		if info.Amount == "" && tokenTx.Amount != "" {
			info.Amount = tokenTx.Amount
			info.TxValue = tokenTx.Amount
		}
		if info.From == "" {
			info.From = tokenTx.From
		}
		if info.To == "" {
			info.To = tokenTx.To
		}
	} else {
		info.TokenAddress = "0x0000000000000000000000000000000000000000"
		nativeSymbol := getNativeSymbol(chainIndex)
		if info.Symbol == "" {
			info.Symbol = nativeSymbol
		}
		if info.TxValueSymbol == "" {
			info.TxValueSymbol = nativeSymbol
		}

		if info.Amount == "" && info.TxValue == "" {
			if len(tx.ToDetails) > 0 && tx.ToDetails[0].Amount != "" {
				info.Amount = tx.ToDetails[0].Amount
				info.TxValue = tx.ToDetails[0].Amount
			} else if len(tx.InternalTransactionDetails) > 0 && tx.InternalTransactionDetails[0].Amount != "" {
				info.Amount = tx.InternalTransactionDetails[0].Amount
				info.TxValue = tx.InternalTransactionDetails[0].Amount
			}
		}

		if info.From == "" {
			if len(tx.FromDetails) > 0 {
				info.From = tx.FromDetails[0].Address
			} else if len(tx.InternalTransactionDetails) > 0 {
				info.From = tx.InternalTransactionDetails[0].From
			}
		}

		if info.To == "" {
			if len(tx.ToDetails) > 0 {
				info.To = tx.ToDetails[0].Address
			} else if len(tx.InternalTransactionDetails) > 0 {
				info.To = tx.InternalTransactionDetails[0].To
			}
		}
	}

	normalizeAddresses(info, chainIndex)

	return info
}

func isEVMChain(chainIndex string) bool {
	nonEVMChains := map[string]bool{
		"0":        true, // BTC
		"501":      true, // SOLANA
		"195":      true, // TRON
		"637":      true, // APTOS
		"5757":     true, // STACKS
		"70000029": true, // SEI
		"607":      true, // TON
		"706":      true, // OSMOSIS
		"784":      true, // SUI
		"988":      true, // STABLE
		"223":      true, // B2
		"4200":     true, // MERLIN
		"200901":   true, // BITLAYER
	}
	return !nonEVMChains[chainIndex]
}

func normalizeAddress(address string, chainIndex string) string {
	if address == "" {
		return address
	}
	if isEVMChain(chainIndex) {
		return strings.ToLower(address)
	}
	return address
}

func normalizeAddresses(info *NormalTransferInfo, chainIndex string) {
	info.From = normalizeAddress(info.From, chainIndex)
	info.To = normalizeAddress(info.To, chainIndex)
	info.TokenAddress = normalizeAddress(info.TokenAddress, chainIndex)
}

func GetNormalTransferByOKX(chainIndex, txHash string) ([]*NormalTransferInfo, error) {
	client, err := NewOKXTransferClient()
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("chainIndex", chainIndex)
	params.Set("txHash", txHash)

	respBody, err := client.doRequest("GET", "/api/v6/dex/post-transaction/transaction-detail-by-txhash", params, nil)
	if err != nil {
		return nil, err
	}

	var response TransactionDetailResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %v", err)
	}

	if response.Code != "0" {
		return nil, fmt.Errorf("API 返回错误: %s - %s", response.Code, response.Msg)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("未找到交易数据")
	}

	var result []*NormalTransferInfo
	for _, tx := range response.Data {
		info := parseTransactionDetail(tx, chainIndex)
		result = append(result, info)
	}

	return result, nil
}

func getNativeSymbol(chainIndex string) string {
	symbolMap := map[string]string{
		"0":        "BTC",
		"1":        "ETH",
		"56":       "BNB",
		"137":      "MATIC",
		"10":       "ETH",
		"42161":    "ETH",
		"43114":    "AVAX",
		"250":      "FTM",
		"8453":     "ETH",
		"59144":    "ETH",
		"5000":     "MNT",
		"60808":    "ETH",
		"34443":    "ETH",
		"143":      "MON",
		"784":      "SUI",
		"10001":    "ETHW",
		"9745":     "XPL",
		"146":      "S",
		"25":       "CRO",
		"169":      "ETH",
		"9004":     "ETH",
		"81457":    "ETH",
		"70000029": "SEI",
		"1329":     "ETH",
		"88888":    "CHZ",
		"13371":    "ETH",
		"1088":     "METIS",
		"706":      "OSMO",
		"195":      "TRX",
		"324":      "ETH",
		"196":      "OKB",
		"200901":   "BTC",
		"204":      "BNB",
		"1101":     "ETH",
		"4689":     "IOTX",
		"534352":   "ETH",
		"7000":     "ZETA",
		"167000":   "ETH",
		"988":      "USDT",
		"607":      "TON",
		"223":      "BTC",
		"80094":    "BERA",
		"100":      "XDAI",
		"4200":     "BTC",
		"369":      "PLS",
		"33139":    "APE",
		"501":      "SOL",
		"637":      "APT",
		"5757":     "STX",
		"130":      "ETH",
	}
	if symbol, ok := symbolMap[chainIndex]; ok {
		return symbol
	}
	return "ETH"
}

func GetNormalTransfer(chain, txHash string) ([]*NormalTransferInfo, error) {
	chainIndex := getChainIndex(chain)
	if chainIndex == "" {
		return nil, fmt.Errorf("不支持的链: %s", chain)
	}
	return GetNormalTransferByOKX(chainIndex, txHash)
}

func getChainNameByIndex(chainIndex string) string {
	if name, ok := chainIndexToNameMap[chainIndex]; ok {
		return name
	}
	return chainIndex
}

func getChainIndex(chain string) string {
	chain = strings.TrimSpace(chain)
	chainUpper := strings.ToUpper(chain)

	chainUpper = strings.ReplaceAll(chainUpper, " ", "_")
	chainUpper = strings.ReplaceAll(chainUpper, "-", "_")

	if index, ok := chainIndexMap[chainUpper]; ok {
		return index
	}

	normalizedMap := map[string]string{
		"ETHEREUM":            "ETH",
		"BITCOIN":             "BTC",
		"BNB":                 "BSC",
		"BNB_CHAIN":           "BSC",
		"BNBCHAIN":            "BSC",
		"BINANCE":             "BSC",
		"BINANCE_SMART_CHAIN": "BSC",
		"BINANCE_CHAIN":       "BSC",
		"AVALANCHE":           "AVAX",
		"AVALANCHE_C":         "AVAX",
		"AVALANCHE_C_CHAIN":   "AVAX",
		"FANTOM":              "FTM",
		"FANTOM_OPERA":        "FTM",
		"BASE_ETH":            "BASE",
		"LINEA_ETH":           "LINEA",
		"MODE_ETH":            "MODE",
		"BOB_ETH":             "BOB",
		"BOB_MAINNET":         "BOB",
		"MANTA_ETH":           "MANTA",
		"MANTA_PACIFIC":       "MANTA",
		"STARKNET_ETH":        "STARKNET",
		"EVM_SEI":             "SEIEVM",
		"ERA_ETH":             "ZKSYNC",
		"ZK_SYNC_ERA":         "ZKSYNC",
		"ZKSYNC_ERA":          "ZKSYNC",
		"POLYGON_ETH":         "POLYGON_ZKEVM",
		"POLYGON_ZKEVM_ETH":   "POLYGON_ZKEVM",
		"SCROLL_ETH":          "SCROLL",
		"SCROLL_L2":           "SCROLL",
		"TAIKO_ETH":           "TAIKO",
		"TAIKO_MAINNET":       "TAIKO",
		"BITLAYER_BTC":        "BITLAYER",
		"B2_BTC":              "B2",
		"B2_NETWORK":          "B2",
		"MERLIN_BTC":          "MERLIN",
		"MERLIN_CHAIN":        "MERLIN",
		"OP_BNB":              "OPBNB",
		"OPBNB_CHAIN":         "OPBNB",
		"UNICHAIN_ETH":        "UNICHAIN",
		"ETHEREUMPOW":         "ETHW",
		"ETHW_POW":            "ETHW",
		"SOL":                 "SOLANA",
		"APT":                 "APTOS",
		"STX":                 "STACKS",
		"STACKS_BTC":          "STACKS",
		"OSMO":                "OSMOSIS",
		"CHZ":                 "CHILIZ",
		"CHILIZ_CHAIN":        "CHILIZ",
		"IMX":                 "IMMUTABLE",
		"IMMUTABLE_ZKEVM":     "IMMUTABLE",
		"CRO":                 "CRONOS",
		"MNT":                 "MANTLE",
		"MON":                 "MONAD",
		"XPL":                 "PLASMA",
		"S":                   "SONIC",
		"SONIC_MAINNET":       "SONIC",
		"PLS":                 "PULSECHAIN",
		"PULSE":               "PULSECHAIN",
		"GUSDT":               "STABLE",
		"STABLE_COIN":         "STABLE",
		"XDAI":                "GNOSIS",
		"GNOSIS_CHAIN":        "GNOSIS",
		"IOTX":                "IOTEX",
		"IOTEX_NETWORK":       "IOTEX",
		"CFX":                 "CONFLUX",
		"CONFLUX_ESPACE":      "CONFLUX",
		"OKB":                 "XLAYER",
		"OKX_CHAIN":           "XLAYER",
		"BERA":                "BERACHAIN",
		"APECHAIN":            "APE",
		"APE_CHAIN":           "APE",
		"ZETACHAIN":           "ZETA",
		"TONCHAIN":            "TON",
		"THE_OPEN_NETWORK":    "TON",
		"TRX":                 "TRON",
		"ARB":                 "ARBITRUM",
		"ARBITRUM_ONE":        "ARBITRUM",
		"OP":                  "OPTIMISM",
		"MATIC":               "POLYGON",
		"POLYGON_MATIC":       "POLYGON",
	}

	if normalized, ok := normalizedMap[chainUpper]; ok {
		if index, ok := chainIndexMap[normalized]; ok {
			return index
		}
	}

	return ""
}
