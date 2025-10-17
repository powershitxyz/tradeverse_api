package oauth

import "github.com/shopspring/decimal"

type FreezeReq struct {
	Amount     uint64 `json:"amount"`
	ExternalID string `json:"external_id"`
	Remark     string `json:"remark"`
}

type SpendOrUnfreezeReq struct {
	FreezeID uint64 `json:"freeze_id"`
}

type GameStartReq struct {
	SessionID       string `json:"session_id"`
	PlaySettingCode string `json:"play_setting_code"`
	ExternalID      string `json:"external_id"`
	Remark          string `json:"remark"`
}

type GameEndReq struct {
	SessionID   string          `json:"session_id"`
	Score       decimal.Decimal `json:"score"`
	GameStartID uint64          `json:"operation_id"`
}
