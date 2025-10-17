package home

import (
	"time"

	"github.com/shopspring/decimal"
)

var seasonLeaderboardSql = `
	WITH per_game AS (
	SELECT
		s.main_id,
		s.game_id,
		SUM(s.accumulated_score)  AS sum_score,
		SUM(s.accumulated_amount) AS sum_amount
	FROM n_season_session_board s
	JOIN n_season_game sg
		ON sg.game_id = s.game_id
	AND sg.season_id = ?
	GROUP BY s.main_id, s.game_id
	),
	mx AS (
	SELECT
		game_id,
		MAX(sum_score)  AS max_score,
		MAX(sum_amount) AS max_amount
	FROM per_game
	GROUP BY game_id
	),

	ranked AS (
	SELECT
	p.main_id,
	SUM(
		%s
		+ %s
	) AS weighted_score,
	SUM(p.sum_score)  AS season_sum_score,
	SUM(p.sum_amount) AS season_sum_amount
	FROM per_game p
	GROUP BY p.main_id
    )
	SELECT
	u.user_no,
	u.email,
	r.weighted_score,
	r.season_sum_score,
	r.season_sum_amount
	FROM ranked r
	JOIN n_user_main u ON u.id = r.main_id
	ORDER BY r.weighted_score DESC
	LIMIT 100;
	`

type SeasonLeaderboard struct {
	UserNo          string          `json:"user_no"`
	Email           string          `json:"email"`
	WeightedScore   decimal.Decimal `json:"weighted_score"`
	SeasonSumScore  int64           `json:"season_sum_score"`
	SeasonSumAmount int64           `json:"season_sum_amount"`
	EstimateReward  decimal.Decimal `json:"estimate_reward"`
}

type GameSessionWithUser struct {
	StartTime   time.Time       `gorm:"column:start_time" json:"start_time"`
	EndTime     time.Time       `gorm:"column:end_time" json:"end_time"`
	MaxScore    decimal.Decimal `gorm:"column:max_score" json:"max_score"`
	UserNo      string          `gorm:"column:user_no" json:"user_no"`
	Email       string          `gorm:"column:email" json:"email"`
	Name        string          `gorm:"column:name" json:"name"`
	Avatar      string          `gorm:"column:avatar" json:"avatar"`
	CountryCode string          `gorm:"column:country_code" json:"country_code"`
}
