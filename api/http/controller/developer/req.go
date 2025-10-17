package developer

import "github.com/shopspring/decimal"

type UpdateProfileReq struct {
	DevName     string `json:"dev_name"`
	CountryCode string `json:"country_code"`
	Website     string `json:"website"`
}

type UpdateGameReq struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Info          string `json:"info"`
	Description   string `json:"description"`
	Avatar        string `json:"avatar"`
	Image         string `json:"image"`
	PlayUrl       string `json:"play_url"`
	AmountPerPlay uint64 `json:"amount_per_play"`
}

type GameKeyReq struct {
	ClientID string `json:"client_id"`
	GameID   uint64 `json:"game_id"`
}

type GameTestingStartReq struct {
	GameID uint64 `json:"game_id"`
}

type GameSettingSaveReq struct {
	GameID        uint64          `json:"game_id"`
	Code          string          `json:"code"`
	Catalog       string          `json:"catalog"`
	AmountPerPlay decimal.Decimal `json:"amount_per_play"`
}

type GameSettingDeleteReq struct {
	GameID uint64 `json:"game_id"`
	Code   string `json:"code"`
}

type GameSessionListReq struct {
	Page      int    `json:"pn"`
	PageSize  int    `json:"ps"`
	GameID    int    `json:"game_id"`
	Status    int    `json:"status"`
	Testing   *int   `json:"testing"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}
