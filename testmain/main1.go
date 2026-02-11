package main

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
	"time"
)

type OKXWalletClient struct {
	apiKey     string
	secretKey  string
	passphrase string
	baseURL    string
	httpClient *http.Client
}

func NewOKXWalletClient() (*OKXWalletClient, error) {
	apiKey := os.Getenv("OKX_API_KEY")
	secretKey := os.Getenv("OKX_SECRET_KEY")
	passphrase := os.Getenv("OKX_API_PASSPHRASE")

	if apiKey == "" || secretKey == "" || passphrase == "" {
		return nil, fmt.Errorf("缺少必要的 OKX API 环境变量: OKX_API_KEY, OKX_SECRET_KEY, OKX_API_PASSPHRASE")
	}

	return &OKXWalletClient{
		apiKey:     apiKey,
		secretKey:  secretKey,
		passphrase: passphrase,
		baseURL:    "https://web3.okx.com",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *OKXWalletClient) createSignature(timestamp, method, requestPath, body string) string {
	message := timestamp + method + requestPath + body
	h := hmac.New(sha256.New, []byte(c.secretKey))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (c *OKXWalletClient) getHeaders(timestamp, method, requestPath, body string) map[string]string {
	signature := c.createSignature(timestamp, method, requestPath, body)
	return map[string]string{
		"OK-ACCESS-KEY":        c.apiKey,
		"OK-ACCESS-SIGN":       signature,
		"OK-ACCESS-TIMESTAMP":  timestamp,
		"OK-ACCESS-PASSPHRASE": c.passphrase,
		"Content-Type":         "application/json",
	}
}

func (c *OKXWalletClient) doRequest(method, endpoint string, params url.Values, body []byte) ([]byte, error) {
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

	signMessage := timestamp + method + requestPath + queryString + bodyStr
	fmt.Printf("[DEBUG] 时间戳: %s\n", timestamp)
	fmt.Printf("[DEBUG] 请求方法: %s\n", method)
	fmt.Printf("[DEBUG] 请求路径: %s\n", requestPath)
	fmt.Printf("[DEBUG] 查询字符串: %s\n", queryString)
	fmt.Printf("[DEBUG] Body字符串: %s\n", bodyStr)
	fmt.Printf("[DEBUG] 完整请求URL: %s\n", c.baseURL+requestPath+queryString)
	fmt.Printf("[DEBUG] 签名字符串: %s\n", signMessage)
	fmt.Printf("[DEBUG] 签名: %s\n", signature)

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

type SupportedChainResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		ChainIndex   string `json:"chainIndex"`
		ChainName    string `json:"chainName"`
		ChainLogoUrl string `json:"chainLogoUrl"`
		ChainSymbol  string `json:"chainSymbol"`
	} `json:"data"`
}

func (c *OKXWalletClient) GetSupportedChains(chainIndex string) (*SupportedChainResponse, error) {
	params := url.Values{}
	if chainIndex != "" {
		params.Set("chainIndex", chainIndex)
	}

	respBody, err := c.doRequest("GET", "/api/v6/dex/market/supported/chain", params, nil)
	if err != nil {
		return nil, err
	}

	var response SupportedChainResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %v", err)
	}

	if response.Code != "0" {
		return nil, fmt.Errorf("API 返回错误: %s - %s", response.Code, response.Msg)
	}

	return &response, nil
}

type TransactionDetailResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		ChainIndex  string `json:"chainIndex"`
		Height      string `json:"height"`
		TxTime      string `json:"txTime"`
		Txhash      string `json:"txhash"`
		TxStatus    string `json:"txStatus"`
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
	} `json:"data"`
}

func (c *OKXWalletClient) GetTransactionDetail(chainIndex, txHash, itype string) (*TransactionDetailResponse, error) {
	params := url.Values{}
	params.Set("chainIndex", chainIndex)
	params.Set("txHash", txHash)
	if itype != "" {
		params.Set("itype", itype)
	}

	respBody, err := c.doRequest("GET", "/api/v6/dex/post-transaction/transaction-detail-by-txhash", params, nil)
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

	return &response, nil
}

func main1() {
	os.Setenv("OKX_API_KEY", "41a7e5f0-a650-4a49-bbb1-fc28d7489047")
	os.Setenv("OKX_SECRET_KEY", "DA2DBC087E43F33B55ADFCDBBF5ABEF5")
	os.Setenv("OKX_API_PASSPHRASE", "@Aa147258")

	client, err := NewOKXWalletClient()
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== 1. 获取所有支持的链 ===")
	chainsResp, err := client.GetSupportedChains("")
	if err != nil {
		fmt.Printf("获取支持的链失败: %v\n", err)
	} else {
		fmt.Printf("响应码: %s\n", chainsResp.Code)
		fmt.Printf("支持的链数量: %d\n", len(chainsResp.Data))
		for _, chain := range chainsResp.Data {
			fmt.Printf("  - %s (%s): chainIndex=%s\n", chain.ChainName, chain.ChainSymbol, chain.ChainIndex)
		}
	}

	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n=== 2. 获取第一个交易详情 ===")
	time.Sleep(1 * time.Second)
	txHash1 := "0x655a7d748c5fd206742d95b7da11720ad3a2f94795e036acea9e4ba469e1f9f0"
	txDetail1, err := client.GetTransactionDetail("1", txHash1, "")
	if err != nil {
		fmt.Printf("获取交易详情失败: %v\n", err)
	} else {
		fmt.Printf("响应码: %s\n", txDetail1.Code)
		if len(txDetail1.Data) > 0 {
			tx := txDetail1.Data[0]
			fmt.Printf("交易哈希: %s\n", tx.Txhash)
			fmt.Printf("链标识: %s\n", tx.ChainIndex)
			fmt.Printf("区块高度: %s\n", tx.Height)
			fmt.Printf("交易时间: %s\n", tx.TxTime)
			fmt.Printf("交易状态: %s\n", tx.TxStatus)
			fmt.Printf("发送方数量: %d\n", len(tx.FromDetails))
			fmt.Printf("接收方数量: %d\n", len(tx.ToDetails))
			fmt.Printf("内部交易数量: %d\n", len(tx.InternalTransactionDetails))
			fmt.Printf("代币转移数量: %d\n", len(tx.TokenTransferDetails))
		}
	}

	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n=== 3. 获取第二个交易详情 ===")
	time.Sleep(1 * time.Second)
	txHash2 := "0x4c8f255ec964e62c44f4ef4fb148f712185b36fd220bca21a6600e5d55811709"
	txDetail2, err := client.GetTransactionDetail("1", txHash2, "")
	if err != nil {
		fmt.Printf("获取交易详情失败: %v\n", err)
	} else {
		fmt.Printf("响应码: %s\n", txDetail2.Code)
		if len(txDetail2.Data) > 0 {
			tx := txDetail2.Data[0]
			fmt.Printf("交易哈希: %s\n", tx.Txhash)
			fmt.Printf("链标识: %s\n", tx.ChainIndex)
			fmt.Printf("区块高度: %s\n", tx.Height)
			fmt.Printf("交易时间: %s\n", tx.TxTime)
			fmt.Printf("交易状态: %s\n", tx.TxStatus)
			fmt.Printf("发送方数量: %d\n", len(tx.FromDetails))
			fmt.Printf("接收方数量: %d\n", len(tx.ToDetails))
			fmt.Printf("内部交易数量: %d\n", len(tx.InternalTransactionDetails))
			fmt.Printf("代币转移数量: %d\n", len(tx.TokenTransferDetails))
		}
	}
}
