package model

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"log"
	"testing"

	"github.com/mr-tron/base58"
)

func TestAuthSign(t *testing.T) {
	msg := "Wallet:CB3LXZ28gWcU9cc9Tkv4QKGnbnJjFUgBfPXgXAN3tq98\n|Message:Welcome to Stonks\n|Nonce:suiqMXoyCw\n"
	hash := sha256.Sum256([]byte(msg))
	result := hex.EncodeToString(hash[:])
	log.Println(result)

	// 解码钱包公钥（Base58 转字节）
	publicKey, err := base58.Decode("CB3LXZ28gWcU9cc9Tkv4QKGnbnJjFUgBfPXgXAN3tq98")
	if err != nil {
		log.Println(err)
	}

	// 解码签名（Base64 转字节）
	signature, err := base64.StdEncoding.DecodeString("yty5l4sGrg7ptVYy92+osc0zj4YeCcfpzxGopm0PsfLnVfntnUvujeMfVaw41txHHbjEz78BDsrUPvcJzJYyCg==")
	if err != nil {
		log.Println(err)
	}

	// 进行签名验证
	if !ed25519.Verify(publicKey, []byte(msg), signature) {
		log.Println("un processed")
	}
}
