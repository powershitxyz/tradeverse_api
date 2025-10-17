package chain

import (
	"context"
	"encoding/json"
	"testing"
)

// 手动本地测试：
// 运行前请设置环境变量：
//
//	CHAOS_SOL_RPC  - Solana HTTP RPC，例如 https://mainnet.helius-rpc.com/?api-key=xxxx
//	CHAOS_SOL_SIG  - 交易签名（hash）
//	CHAOS_SOL_MINT - 目标 SPL mint 地址

func TestAnalyzeTxMintDelta_Manual(t *testing.T) {
	rpc := "https://solana.publicnode.com"
	sig := "2RvLnfjAGFJ3tsKC8N6oktuLrhoUvx3A23oojLfv6Jw9UmTQPp34Kn2FMymyVhXRgKXYdd8dPV683VPkdf1LGqnD"
	sig = "32ueEjtQ6B5ih3YvNshPPAVwrHMy1n4wGhjCFqZwC26k3tNV3wRtY1smfzKYJTtCwAL6fRTqcGCTEPTM7Vfq5dGw"
	mint := "3ovR2CQczTA3T6t37UnyMj2pD3VgvJ6Dvec6rvFishot"
	if rpc == "" || sig == "" || mint == "" {
		t.Skip("set CHAOS_SOL_RPC, CHAOS_SOL_SIG, CHAOS_SOL_MINT to run")
	}

	ctx := context.Background()
	res, err := AnalyzeTxMintDelta(ctx, rpc, sig, mint)
	if err != nil {
		t.Fatalf("AnalyzeTxMintDelta error: %v", err)
	}
	b, _ := json.MarshalIndent(res, "", "  ")
	t.Logf("Result:\n%s", string(b))
}
