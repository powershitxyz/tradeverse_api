package auth

import "github.com/shopspring/decimal"

type UpdateProfileReq struct {
	Avatar      string `json:"avatar"`
	Name        string `json:"name"`
	Bio         string `json:"bio"`
	Birthday    string `json:"birthday"`
	CountryCode string `json:"country_code"`
	Timezone    int    `json:"timezone"`
	XUri        string `json:"x_uri"`
}

type SendEmailReq struct {
	Email string `json:"email"`
}

type VerifyEmailReq struct {
	SendEmailReq
	Code string `json:"code"`
}

type ConfirmGameSessionReq struct {
	SessionID string          `json:"session_id"`
	Score     decimal.Decimal `json:"score"`
}

type JoinSeasonReq struct {
	SeasonID   uint64 `json:"season_id"`
	SeasonCode string `json:"season_code"`
}

type TopupReq struct {
	TxHash  string          `json:"tx_hash"`
	ChainID uint64          `json:"chain_id"`
	Amount  decimal.Decimal `json:"amount"`
	Type    string          `json:"type"` // recharge/withdraw/lock/cancel
	RefID   uint64          `json:"ref_id"`
}

type WithdrawReq struct {
	Amount  decimal.Decimal `json:"amount"`
	ChainID uint64          `json:"chain_id"`
	ID      uint64          `json:"id"`
}
