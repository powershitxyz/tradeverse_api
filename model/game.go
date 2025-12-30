package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	SeasonStatusDraft            = "00"
	SeasonStatusActive           = "10"
	SeasonStatusEnded            = "20"
	SeasonStatusSettled          = "30"
	SeasonIsVisible              = 1
	SeasonIsNotVisible           = 0
	SeasonAllowJoinAfterStart    = 1
	SeasonNotAllowJoinAfterStart = 0
	GameSessionStatusStart       = 1
	GameSessionStatusEnd         = 2
	GameSessionStatusReversed    = 3
	GameSessionStatusSettled     = 4
	GameSessionStatusAuditing    = 5
	GameStatusDraft              = "00"
	GameStatusTesting            = "05"
	GameStatusTested             = "06"
	GameStatusPendingApprove     = "07"
	GameStatusWaitingOnline      = "08"
	GameStatusActive             = "10"
	GameStatusInactive           = "20"
)

type GameDeveloper struct {
	ID          uint64    `gorm:"column:id;primary_key;auto_increment"`
	DevName     string    `gorm:"column:dev_name" json:"dev_name"`
	Email       string    `gorm:"column:email" json:"email"`
	Password    string    `gorm:"column:password" json:"-"`
	AddTime     time.Time `gorm:"column:add_time" json:"add_time"`
	Website     string    `gorm:"column:website" json:"website"`
	CountryCode string    `gorm:"column:country_code" json:"country_code"`
	Status      string    `gorm:"column:status" json:"status"`
}

func (GameDeveloper) TableName() string {
	return TB_GAME_DEVELOPER
}

type GameInfo struct {
	ID          uint64    `gorm:"column:id;primary_key;auto_increment"`
	DevID       uint64    `gorm:"column:dev_id" json:"-"`
	Name        string    `gorm:"column:name" json:"name"`
	Info        string    `gorm:"column:info" json:"info"`
	Description string    `gorm:"column:description" json:"description"`
	Avatar      string    `gorm:"column:avatar" json:"avatar"`
	Image       string    `gorm:"column:image" json:"image"`
	AddTime     time.Time `gorm:"column:add_time" json:"add_time"`
	PlayUrl     string    `gorm:"column:play_url" json:"play_url"`
	Status      string    `gorm:"column:status" json:"status"`
}

func (GameInfo) TableName() string {
	return TB_GAME_INFO
}

type GameApp struct {
	ID            uint64    `gorm:"column:id;primary_key;auto_increment"`
	GameID        uint64    `gorm:"column:game_id" json:"game_id"`
	AddTime       time.Time `gorm:"column:add_time" json:"add_time"`
	ClientID      string    `gorm:"column:client_id" json:"client_id"`
	ClientSecret  string    `gorm:"column:client_secret" json:"client_secret"`
	OauthCallback string    `gorm:"column:oauth_callback" json:"oauth_callback"`
}

func (GameApp) TableName() string {
	return TB_GAME_APP
}

type GameSetting struct {
	ID            uint64    `gorm:"column:id;primary_key;auto_increment"`
	GameID        uint64    `gorm:"column:game_id" json:"game_id"`
	Code          string    `gorm:"column:code" json:"code"`
	Catalog       string    `gorm:"column:catalog" json:"catalog"`
	AmountPerPlay uint64    `gorm:"column:amount_per_play" json:"amount_per_play"`
	AddTime       time.Time `gorm:"column:add_time" json:"add_time"`
	UpdateTime    time.Time `gorm:"column:update_time" json:"update_time"`
}

func (GameSetting) TableName() string {
	return TB_GAME_SETTING
}

type GameSession struct {
	SessionID       string          `gorm:"column:session_id;primary_key" json:"session_id"`
	GameID          uint64          `gorm:"column:game_id" json:"game_id"`
	MainID          uint64          `gorm:"column:main_id" json:"-"`
	StartTime       time.Time       `gorm:"column:start_time" json:"start_time"`
	EndTime         time.Time       `gorm:"column:end_time" json:"end_time"`
	Status          int             `gorm:"column:status" json:"status"`
	Score           decimal.Decimal `gorm:"column:score" json:"score"`
	UserReportScore decimal.Decimal `gorm:"column:user_report_score" json:"user_report_score"`
	SpendAmountN    uint64          `gorm:"column:spend_amount_n" json:"spend_amount_n"`
	Testing         int             `gorm:"column:testing" json:"testing"`
}

func (GameSession) TableName() string {
	return TB_GAME_SESSION
}

// ////////////////////////////////////////////////////////////////////////////////////////
type SeasonInfo struct {
	ID                  uint64     `gorm:"column:id;primary_key;auto_increment"`
	Title               string     `gorm:"column:title" json:"title"`
	Info                string     `gorm:"column:info" json:"info"`
	Description         string     `gorm:"column:description" json:"description"`
	Code                string     `gorm:"column:code" json:"code"`
	BannerImage         string     `gorm:"column:banner_image" json:"banner_image"`
	StartTime           time.Time  `gorm:"column:start_time" json:"start_time"`
	EndTime             time.Time  `gorm:"column:end_time" json:"end_time"`
	Status              string     `gorm:"column:status" json:"status"`
	Timezone            int        `gorm:"column:timezone" json:"timezone"`
	BasePrizeN          uint64     `gorm:"column:base_prize_n" json:"-"`
	BasePrizeU          uint64     `gorm:"column:base_prize_u" json:"base_prize_u"`
	DailySpendCapN      uint64     `gorm:"column:daily_spend_cap_n" json:"-"`
	IsVisible           int        `gorm:"column:is_visible" json:"-"`
	AllowJoinAfterStart int        `gorm:"column:allow_join_after_start" json:"allow_join_after_start"`
	AddTime             time.Time  `gorm:"column:add_time" json:"-"`
	SettledAt           time.Time  `gorm:"column:settled_at" json:"settled_at"`
	IncludeGames        []GameInfo `gorm:"-" json:"include_games"`
	LimitUserCount      uint64     `gorm:"column:limit_user_count" json:"limit_user_count"`
	JoinCount           uint64     `gorm:"-" json:"join_count"`
}

func (SeasonInfo) TableName() string {
	return TB_SEASON_INFO
}

type SeasonUser struct {
	ID       uint64    `gorm:"column:id;primary_key;auto_increment"`
	SeasonID uint64    `gorm:"column:season_id" json:"season_id"`
	MainID   uint64    `gorm:"column:main_id" json:"main_id"`
	JoinTime time.Time `gorm:"column:join_time" json:"join_time"`
}

func (SeasonUser) TableName() string {
	return TB_SEASON_USER
}

type SeasonSessionBoard struct {
	ID                uint64    `gorm:"column:id;primary_key;auto_increment"`
	GameID            uint64    `gorm:"column:game_id" json:"game_id"`
	SeasonID          uint64    `gorm:"column:season_id" json:"season_id"`
	MainID            uint64    `gorm:"column:main_id" json:"main_id"`
	AccumulatedScore  uint64    `gorm:"column:accumulated_score" json:"accumulated_score"`
	AccumulatedAmount uint64    `gorm:"column:accumulated_amount" json:"accumulated_amount"`
	UpdateTime        time.Time `gorm:"column:update_time" json:"update_time"`
}

func (SeasonSessionBoard) TableName() string {
	return TB_SEASON_SESSION_BOARD
}

type SeasonGame struct {
	ID           uint64 `gorm:"column:id;primary_key;auto_increment"`
	GameID       uint64 `gorm:"column:game_id" json:"game_id"`
	SeasonID     uint64 `gorm:"column:season_id" json:"season_id"`
	WeightScore  uint64 `gorm:"column:weight_score" json:"weight_score"`
	WeightAmount uint64 `gorm:"column:weight_amount" json:"weight_amount"`
}

func (SeasonGame) TableName() string {
	return TB_SEASON_GAME
}

/*********************** TODO ***********************/
type GameWithTags struct {
	GameInfo
	TagsRaw string   `gorm:"column:tags" json:"-"` // 数据库里的 JSON 原始字符串
	Tags    []string `gorm:"-" json:"tags"`        // 解析后真正返回给前端
}

type GameWithSeason struct {
	GameInfo
	SeasonID uint64 `gorm:"column:season_id" json:"season_id"`
}

func GameSessionStatus(statusStr string) int {
	switch statusStr {
	case "start":
		return GameSessionStatusStart
	case "end":
		return GameSessionStatusEnd
	case "reversed":
		return GameSessionStatusReversed
	case "settled":
		return GameSessionStatusSettled
	case "auditing":
		return GameSessionStatusAuditing
	default:
		return 0
	}
}
