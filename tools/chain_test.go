package tools

import (
	"chaos/api/config"
	"context"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestERC20Token(t *testing.T) {

	chainConfig := config.GetRpcConfig("BSC")

	var rpcURL string
	for _, chain := range chainConfig.GetRpc() {
		rpcURL = chain
	}

	// 示例：BSC 上 N 代币合约
	tokenAddress := "0xb82582bf335bc4f57ec3c536e67019e1fa263f81"
	// 替换成你要查询的钱包地址
	walletAddress := "0xe38533e11B680eAf4C9519Ea99B633BD3ef5c2F8"

	token, err := NewERC20Token(rpcURL, tokenAddress)
	if err != nil {
		log.Fatal("代币初始化失败:", err)
	}

	// 增加超时时间到 30 秒
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 查询余额
	rawBal, err := token.BalanceOf(ctx, walletAddress)
	if err != nil {
		log.Fatal("查询余额失败:", err)
	}

	// 查询 decimals，添加重试机制
	var decimals uint8
	for i := 0; i < 3; i++ {
		decimals, err = token.Decimals(ctx)
		if err == nil {
			break
		}
		log.Printf("第 %d 次查询 decimals 失败: %v", i+1, err)
		if i < 2 {
			time.Sleep(2 * time.Second) // 等待 2 秒后重试
		}
	}
	if err != nil {
		log.Fatal("查询 decimals 失败，已重试 3 次:", err)
	}

	// 查询 symbol，添加重试机制
	var symbol string
	for i := 0; i < 3; i++ {
		symbol, err = token.Symbol(ctx)
		if err == nil {
			break
		}
		log.Printf("第 %d 次查询 symbol 失败: %v", i+1, err)
		if i < 2 {
			time.Sleep(2 * time.Second) // 等待 2 秒后重试
		}
	}
	if err != nil {
		// 有些代币 symbol 会报错，这里容错
		symbol = ""
		log.Printf("查询 symbol 失败，使用空字符串: %v", err)
	}

	fmt.Printf("地址: %s\n", walletAddress)
	fmt.Printf("代币: %s (%s)\n", symbol, tokenAddress)
	fmt.Printf("原始余额: %s\n", rawBal.String())
	fmt.Printf("可读余额: %s %s\n", FormatBalance(rawBal, decimals), symbol)
}
