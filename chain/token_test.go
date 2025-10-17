package chain

import (
	"context"
	"encoding/binary"
	"fmt"
	"sort"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// TokenHolder 结构体
type TokenHolder struct {
	Owner    string `json:"owner"`
	Amount   uint64 `json:"amount"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

func parseAccountData(data []byte) (string, uint64) {
	ownerPubkey := solana.PublicKeyFromBytes(data[32:64]).String()
	amount := binary.LittleEndian.Uint64(data[64:72])
	return ownerPubkey, amount
}

// 获取 Solana 代币持有者列表
func getTokenHoldersFromRPC(tokenMintAddress string) ([]TokenHolder, error) {
	client := rpc.New("https://mainnet.helius-rpc.com/?api-key=f03a4129-4b9c-4a95-a0be-32a99ed9a36c") // 替换为你的 Solana RPC 地址

	mintPubkey, err := solana.PublicKeyFromBase58(tokenMintAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid token mint address: %v", err)
	}

	// RPC 请求
	res, err := client.GetProgramAccountsWithOpts(
		context.Background(),
		solana.TokenProgramID,
		&rpc.GetProgramAccountsOpts{
			Filters: []rpc.RPCFilter{
				{DataSize: 165}, // 确保数据大小匹配 SPL 账户
				{Memcmp: &rpc.RPCFilterMemcmp{Offset: 0, Bytes: mintPubkey[:]}}, // 过滤指定的代币 mint 地址
			},
			Encoding: solana.EncodingBase64,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}

	// 解析返回的账户数据
	var holders []TokenHolder
	for _, account := range res {
		_, amount := parseAccountData(account.Account.Data.GetBinary())

		holder := account.Pubkey.String()
		holders = append(holders, TokenHolder{Owner: holder, Amount: amount})
	}

	// 按照持仓数量从多到少排序
	sort.Slice(holders, func(i, j int) bool {
		return holders[i].Amount > holders[j].Amount
	})

	return holders, nil
}

func TestGetHolders(t *testing.T) {
	getTokenHoldersFromRPC("Buoj8HCZMnLRwzDmjzaswhkVhLZD58PG4pZ7rnYp6pCr")
}
