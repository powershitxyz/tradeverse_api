package tools

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ExampleUsage 展示如何在生产环境中使用客户端管理器
func ExampleUsage() {
	// 1. 初始化时获取 RPC URL
	rpcURL := "https://bsc-dataseed1.binance.org/"

	// 2. 创建多个代币实例，它们会复用同一个客户端连接
	token1, err := NewERC20Token(rpcURL, "0xb82582bf335bc4f57ec3c536e67019e1fa263f81") // N 代币
	if err != nil {
		log.Fatal("创建 N 代币实例失败:", err)
	}

	token2, err := NewERC20Token(rpcURL, "0x55d398326f99059fF775485246999027B3197955") // USDT
	if err != nil {
		log.Fatal("创建 USDT 代币实例失败:", err)
	}

	// 3. 并发查询多个代币余额
	walletAddress := "0xe38533e11B680eAf4C9519Ea99B633BD3ef5c2F8"
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 并发查询
	type result struct {
		symbol   string
		balance  string
		decimals uint8
		err      error
	}

	results := make(chan result, 2)

	// 查询 N 代币
	go func() {
		balance, err := token1.BalanceOf(ctx, walletAddress)
		if err != nil {
			results <- result{err: err}
			return
		}

		decimals, err := token1.Decimals(ctx)
		if err != nil {
			results <- result{err: err}
			return
		}

		symbol, _ := token1.Symbol(ctx)
		results <- result{
			symbol:   symbol,
			balance:  FormatBalance(balance, decimals),
			decimals: decimals,
		}
	}()

	// 查询 USDT
	go func() {
		balance, err := token2.BalanceOf(ctx, walletAddress)
		if err != nil {
			results <- result{err: err}
			return
		}

		decimals, err := token2.Decimals(ctx)
		if err != nil {
			results <- result{err: err}
			return
		}

		symbol, _ := token2.Symbol(ctx)
		results <- result{
			symbol:   symbol,
			balance:  FormatBalance(balance, decimals),
			decimals: decimals,
		}
	}()

	// 收集结果
	for i := 0; i < 2; i++ {
		r := <-results
		if r.err != nil {
			fmt.Printf("查询失败: %v\n", r.err)
		} else {
			fmt.Printf("%s 余额: %s (精度: %d)\n", r.symbol, r.balance, r.decimals)
		}
	}

	// 4. 应用关闭时清理连接
	// 在 main 函数或应用关闭时调用
	// GetGlobalClient().CloseAll()
}

// GetTokenBalance 便捷函数：获取指定代币余额
func GetTokenBalance(rpcURL, tokenAddress, walletAddress string) (string, error) {
	token, err := NewERC20Token(rpcURL, tokenAddress)
	if err != nil {
		return "", fmt.Errorf("创建代币实例失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	balance, err := token.BalanceOf(ctx, walletAddress)
	if err != nil {
		return "", fmt.Errorf("查询余额失败: %w", err)
	}

	decimals, err := token.Decimals(ctx)
	if err != nil {
		return "", fmt.Errorf("查询精度失败: %w", err)
	}

	return FormatBalance(balance, decimals), nil
}
