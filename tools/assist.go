package tools

import (
	"crypto/rand"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func Float64Ptr(v float64) *float64 {
	return &v
}

func Deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func NewLockID32() ([32]byte, string, error) {
	var id32 [32]byte
	if _, err := rand.Read(id32[:]); err != nil {
		return [32]byte{}, "", err
	}
	// hex 字符串（用于返回/存库）
	hex := hexutil.Encode(id32[:]) // "0x..."
	return id32, hex, nil
}

func BigIntToBytes32(n *big.Int) [32]byte {
	var out [32]byte
	if n == nil {
		return out // 默认全 0
	}
	b := n.Bytes()
	if len(b) > 32 {
		b = b[len(b)-32:]
	}
	copy(out[32-len(b):], b)
	return out
}
