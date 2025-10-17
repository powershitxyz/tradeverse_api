package model

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/mr-tron/base58"
)

type UserWallet struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Wallet     string    `gorm:"column:wallet;type:varchar(255);not null" json:"wallet"`
	Chain      string    `gorm:"column:chain" json:"chain"`
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
	MainID     uint64    `gorm:"column:main_id" json:"main_id"`
	RefID      uint64    `gorm:"column:ref_id" json:"ref_id"`
}

func (UserWallet) TableName() string {
	return "user_wallet"
}

type AuthMessage struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	AuthKey    string    `gorm:"column:auth_key;type:varchar(255);not null" json:"auth_key"`
	AuthMsg    string    `gorm:"column:auth_msg" json:"auth_msg"`
	Nonce      string    `gorm:"type:varchar(255);not null" json:"nonce"`
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
	ExpireTime time.Time `gorm:"column:expire_time" json:"expire_time"`
	Type       int       `gorm:"column:type" json:"type"`
	Wallet     string    `gorm:"column:wallet" json:"wallet"`
}

func (AuthMessage) TableName() string {
	return "user_auth_msg"
}

func (auth AuthMessage) ComputeAuthDigest(sigInput string) bool {
	msg := auth.Format()

	// 若是 EVM 地址：先按 personal_sign 校验，失败则回退 eth_sign
	if isEvmAddress(auth.AuthKey) {
		sig, ok := decodeSigFlexible(sigInput)
		if !ok || !normalizeV(sig) {
			return false
		}
		// personal_sign
		if pub, err := gethcrypto.SigToPub(personalMessageHash([]byte(msg)), sig); err == nil {
			if strings.EqualFold(gethcrypto.PubkeyToAddress(*pub).Hex(), auth.AuthKey) {
				return true
			}
		}
		// eth_sign（无前缀）
		if pub, err := gethcrypto.SigToPub(gethcrypto.Keccak256Hash([]byte(msg)).Bytes(), sig); err == nil {
			if strings.EqualFold(gethcrypto.PubkeyToAddress(*pub).Hex(), auth.AuthKey) {
				return true
			}
		}
		return false
	}

	// Solana ed25519（公钥 base58，签名 base64）
	publicKey, err := base58.Decode(auth.AuthKey)
	if err != nil {
		log.Println(err)
		return false
	}
	signature, err := base64.StdEncoding.DecodeString(sigInput)
	if err != nil {
		log.Println(err)
		return false
	}
	return ed25519.Verify(publicKey, []byte(msg), signature)
}

func (auth AuthMessage) Format() string {
	data := fmt.Sprintf("Wallet:%s\nMessage:%s\nNonce:%s\n",
		auth.AuthKey,
		auth.AuthMsg,
		auth.Nonce,
	)
	return data
}

type DailyCheck struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UwID      int64     `gorm:"column:uw_id;type:varchar(255);not null" json:"uw_id"`
	CheckDate string    `gorm:"column:check_date" json:"check_date"`
	CheckTime time.Time `gorm:"column:check_time" json:"check_time"`
}

func (DailyCheck) TableName() string {
	return "daily_checkin"
}

type UserAttention struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UwID       int64     `gorm:"column:uw_id;type:varchar(255);not null" json:"uw_id"`
	Chain      string    `gorm:"column:chain" json:"chain"`
	Ca         string    `gorm:"column:ca" json:"ca"`
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time" json:"update_time"`
	Flag       int       `gorm:"column:flag" json:"flag"`
}

func (UserAttention) TableName() string {
	return "user_attention"
}

type ChaosHolder struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	WalletAddress string    `gorm:"column:wallet_address" json:"wallet_address"`
	SnapAmount    int64     `gorm:"column:snap_amount" json:"snap_amount"`
	Slot          uint64    `gorm:"column:slot" json:"slot"`
	AfterTx       int64     `gorm:"column:after_tx" json:"after_tx"`
	BscWallet     string    `gorm:"column:bsc_wallet" json:"bsc_wallet"`
	Step          int       `gorm:"column:step" json:"step"`
	UpdateTime    time.Time `gorm:"column:update_time" json:"update_time"`
}

func (ChaosHolder) TableName() string {
	return "chaos_holder_357547209"
}

type ChaosTrans struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserWallet string    `gorm:"column:user_wallet" json:"user_wallet"`
	TxHash     string    `gorm:"column:tx_hash" json:"tx_hash"`
	AddTime    time.Time `gorm:"column:add_time" json:"add_time"`
	Amount     uint64    `gorm:"column:amount" json:"amount"`
	Status     string    `gorm:"column:status" json:"status"`
	Flag       int       `gorm:"column:flag" json:"flag"`
}

func (ChaosTrans) TableName() string {
	return "chaos_holder_trans"
}

type NAirdropClaim struct {
	ID                uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	SolWallet         string `gorm:"column:sol_wallet" json:"sol_wallet"`
	BscWallet         string `gorm:"column:bsc_wallet" json:"bsc_wallet"`
	SolSnapAmount     uint64 `gorm:"column:sol_snap_amount" json:"sol_snap_amount"`
	SolTransferAmount uint64 `gorm:"column:sol_transfer_amount" json:"sol_transfer_amount"`
}

func (NAirdropClaim) TableName() string {
	return "n_airdrop_claim"
}

// helper: EVM 地址判断
func isEvmAddress(s string) bool {
	if len(s) != 42 {
		return false
	}
	if !strings.HasPrefix(s, "0x") && !strings.HasPrefix(s, "0X") {
		return false
	}
	_, err := hex.DecodeString(s[2:])
	return err == nil
}

// helper: EIP-191 personal_sign 哈希
func personalMessageHash(message []byte) []byte {
	prefix := []byte("\x19Ethereum Signed Message:\n" + strconv.Itoa(len(message)))
	data := append(prefix, message...)
	h := gethcrypto.Keccak256Hash(data)
	return h.Bytes()
}

// helper: 签名解码（优先 hex，再 base64，最后兜底）
func decodeSigFlexible(s string) ([]byte, bool) {
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	// 优先判断并尝试 hex
	if strings.HasPrefix(lower, "0x") || looksLikeHex(s) {
		ss := strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X")
		if b, err := hex.DecodeString(ss); err == nil {
			return b, true
		}
	}
	// 再尝试 base64
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return b, true
	}
	// 兜底：再试一次纯 hex
	if b, err := hex.DecodeString(s); err == nil {
		return b, true
	}
	return nil, false
}

func looksLikeHex(s string) bool {
	ss := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(s), "0x"), "0X")
	if len(ss)%2 != 0 || len(ss) < 2 {
		return false
	}
	for i := 0; i < len(ss); i++ {
		c := ss[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// helper: 归一化 v，支持 27/28 与 EIP-155 (>=35)
func normalizeV(sig []byte) bool {
	if len(sig) != 65 {
		return false
	}
	v := sig[64]
	switch {
	case v == 27 || v == 28:
		sig[64] = v - 27
	case v >= 35:
		// v = 35 + 2*chainId + {0,1}
		sig[64] = (v - 35) % 2
	case v <= 1:
		// already normalized
	default:
		return false
	}
	return sig[64] <= 1
}
