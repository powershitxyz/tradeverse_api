package tools

import (
	"chaos/api/config"
	"chaos/api/log"
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// 全局客户端管理器
type ClientManager struct {
	clients map[string]*ethclient.Client
	n       *ERC20Token
	mutex   sync.RWMutex
}

var (
	globalClientManager = &ClientManager{
		clients: make(map[string]*ethclient.Client),
	}
)

func init() {
	n, err := newNToken()
	if err != nil {
		log.Errorf("failed to create N token: %v", err)
		// 不要设置 N 为 nil，让调用者检查
	} else {
		globalClientManager.n = n
	}
}

// GetNToken 安全获取 N 代币实例
func GetNToken() (*ERC20Token, error) {
	if globalClientManager.n == nil {
		return nil, fmt.Errorf("N token not initialized")
	}
	return globalClientManager.n, nil
}

func newNToken() (*ERC20Token, error) {
	conf := config.GetConfig()

	tokenAddr := conf.Contract.NAddress
	if tokenAddr == "" {
		return nil, fmt.Errorf("N token address is empty")
	}

	chainConfig := config.GetRpcConfig("BSC")
	if chainConfig == nil {
		return nil, fmt.Errorf("BSC chain config is nil")
	}

	var rpcURL string
	for _, chain := range chainConfig.GetRpc() {
		rpcURL = chain
		break // 只取第一个 RPC URL
	}

	if rpcURL == "" {
		return nil, fmt.Errorf("no RPC URL found for BSC")
	}

	return NewERC20Token(rpcURL, tokenAddr)
}

// GetClient 获取或创建指定网络的客户端
func (cm *ClientManager) GetClient(rpcURL string) (*ethclient.Client, error) {
	cm.mutex.RLock()
	client, exists := cm.clients[rpcURL]
	cm.mutex.RUnlock()

	if exists {
		// 检查连接是否健康
		if cm.isClientHealthy(client) {
			return client, nil
		}
		// 连接不健康，移除并重新创建
		cm.mutex.Lock()
		delete(cm.clients, rpcURL)
		client.Close()
		cm.mutex.Unlock()
	}

	// 创建新客户端
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 双重检查，避免并发创建
	if client, exists = cm.clients[rpcURL]; exists {
		return client, nil
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC: %w", err)
	}

	cm.clients[rpcURL] = client
	return client, nil
}

// isClientHealthy 检查客户端连接是否健康
func (cm *ClientManager) isClientHealthy(client *ethclient.Client) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.BlockNumber(ctx)
	return err == nil
}

// CloseAll 关闭所有客户端连接
func (cm *ClientManager) CloseAll() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for _, client := range cm.clients {
		client.Close()
	}
	cm.clients = make(map[string]*ethclient.Client)
}

func (cm *ClientManager) GetNToken() (*ERC20Token, error) {
	return globalClientManager.n, nil
}

// GetGlobalClient 获取全局客户端管理器
func GetGlobalClient() *ClientManager {
	return globalClientManager
}

const erc20ABI = `[
  {"constant":true,"inputs":[{"name":"account","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
  {"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"stateMutability":"view","type":"function"},
  {"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"stateMutability":"view","type":"function"}
]`

type ERC20Token struct {
	Address common.Address
	ABI     abi.ABI
	Client  *ethclient.Client
}

// NewERC20Token 创建新的 ERC20 代币实例，使用缓存的客户端
func NewERC20Token(rpcURL, tokenAddr string) (*ERC20Token, error) {
	client, err := globalClientManager.GetClient(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}
	addr := common.HexToAddress(tokenAddr)

	return &ERC20Token{
		Address: addr,
		ABI:     parsed,
		Client:  client,
	}, nil
}

// 查询代币余额
func (t *ERC20Token) BalanceOf(ctx context.Context, wallet string) (*big.Int, error) {
	addr := common.HexToAddress(wallet)
	data, err := t.ABI.Pack("balanceOf", addr)
	if err != nil {
		return nil, fmt.Errorf("pack balanceOf data failed: %w", err)
	}

	result, err := t.Client.CallContract(ctx, ethereum.CallMsg{
		To:   &t.Address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("call balanceOf contract failed: %w", err)
	}

	var balance *big.Int
	err = t.ABI.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return nil, fmt.Errorf("unpack balanceOf result failed: %w", err)
	}
	return balance, err
}

// 查询代币精度
func (t *ERC20Token) Decimals(ctx context.Context) (uint8, error) {
	data, err := t.ABI.Pack("decimals")
	if err != nil {
		return 0, fmt.Errorf("pack decimals data failed: %w", err)
	}

	result, err := t.Client.CallContract(ctx, ethereum.CallMsg{
		To:   &t.Address,
		Data: data,
	}, nil)
	if err != nil {
		return 0, fmt.Errorf("call decimals contract failed: %w", err)
	}

	var decimals uint8
	err = t.ABI.UnpackIntoInterface(&decimals, "decimals", result)
	if err != nil {
		return 0, fmt.Errorf("unpack decimals result failed: %w", err)
	}
	return decimals, err
}

// 查询代币符号
func (t *ERC20Token) Symbol(ctx context.Context) (string, error) {
	data, err := t.ABI.Pack("symbol")
	if err != nil {
		return "", fmt.Errorf("pack symbol data failed: %w", err)
	}

	result, err := t.Client.CallContract(ctx, ethereum.CallMsg{
		To:   &t.Address,
		Data: data,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("call symbol contract failed: %w", err)
	}

	var symbol string
	err = t.ABI.UnpackIntoInterface(&symbol, "symbol", result)
	if err != nil {
		return "", fmt.Errorf("unpack symbol result failed: %w", err)
	}
	return symbol, err
}

// 将原始余额转成可读格式
func FormatBalance(raw *big.Int, decimals uint8) string {
	if decimals == 0 {
		return raw.String()
	}
	scale := new(big.Float).SetFloat64(float64(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil).Int64()))
	value := new(big.Float).Quo(new(big.Float).SetInt(raw), scale)
	return value.Text('f', int(decimals)) // 保留 decimals 位
}
