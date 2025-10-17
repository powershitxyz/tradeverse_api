package model

import (
	"time"
)

const (
	FlowRecharge = 0
	FlowFreeze   = 1
	FlowUnfreeze = 2
	FlowSpend    = 3
	FlowWithdraw = 4
	FlowRefund   = 5
)

const (
	FlowStatusPending  = 0
	FlowStatusDone     = 1
	FlowStatusFailed   = 2
	FlowStatusReversed = 3
)

const (
	DirectionNone = 0
	DirectionIn   = 1
	DirectionOut  = 2
)

const (
	BalanceFlowOpRecharge = 0
	BalanceFlowOpFreeze   = 1
	BalanceFlowOpWithdraw = 2
	BalanceFlowOpUnfreeze = 3

	BalanceFlowStatusPending         = 0
	BalanceFlowStatusPendingWithDraw = 4
	BalanceFlowStatusSuccess         = 1
	BalanceFlowStatusFailed          = 2
	BalanceFlowStatusCanceled        = 3
)

func AvailFlowType(typeStr string) int {
	switch typeStr {
	case "recharge":
		return FlowRecharge
	case "freeze":
		return FlowFreeze
	case "unfreeze":
		return FlowUnfreeze
	case "spend":
		return 3
	case "withdraw":
		return FlowSpend
	case "refund":
		return FlowRefund
	}
	return -1
}

type AccountFlow struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MainID         uint64    `gorm:"column:main_id;type:int(11);not null" json:"main_id"`
	AssetID        uint64    `gorm:"column:asset_id;type:int(11);not null" json:"asset_id"`
	BizType        int       `gorm:"column:biz_type;type:varchar(255);not null" json:"biz_type"`
	Amount         uint64    `gorm:"column:amount;type:int(11);not null" json:"amount"`
	Direction      int       `gorm:"column:direction;type:int(11);not null" json:"direction"`
	ClientID       string    `gorm:"column:client_id;type:varchar(255);not null" json:"client_id"`
	GameID         uint64    `gorm:"column:game_id;type:int(11);not null" json:"game_id"`
	ExternalID     string    `gorm:"column:external_id;type:varchar(255);not null" json:"external_id"`
	ExternalRemark string    `gorm:"column:external_remark;type:varchar(255);not null" json:"external_remark"`
	RefFlowID      uint64    `gorm:"column:ref_flow_id;type:int(11);not null" json:"ref_flow_id"`
	Status         int       `gorm:"column:status;type:int(11);not null" json:"status"`
	AddTime        time.Time `gorm:"column:add_time;type:datetime;not null" json:"add_time"`
	UpdateTime     time.Time `gorm:"column:update_time;type:datetime;not null" json:"update_time"`
	SessionID      string    `gorm:"column:session_id;type:varchar(255);not null" json:"session_id"`
}

func (AccountFlow) TableName() string {
	return "n_account_flow"
}

type AccountBalance struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MainID     uint64    `gorm:"column:main_id;type:int(11);not null" json:"main_id"`
	AssetID    uint64    `gorm:"column:asset_id;type:int(11);not null" json:"asset_id"`
	Available  uint64    `gorm:"column:available;type:int(11);not null" json:"available"`
	Frozen     uint64    `gorm:"column:frozen;type:int(11);not null" json:"frozen"`
	UpdateTime time.Time `gorm:"column:update_time;type:datetime;not null" json:"update_time"`
	Withdrawal uint64    `gorm:"column:withdrawal;type:int(11);not null" json:"withdrawal"`
}

func (AccountBalance) TableName() string {
	return "n_account_balance"
}

type AccountBalanceFlow struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MainID         uint64     `gorm:"column:main_id;type:int(11);not null" json:"main_id"`
	AssetID        uint64     `gorm:"column:asset_id;type:int(11);not null" json:"asset_id"`
	Op             int        `gorm:"column:op;type:int(11);not null" json:"op"`
	Status         int        `gorm:"column:status;type:int(11);not null" json:"status"`
	Amount         uint64     `gorm:"column:amount;type:int(11);not null" json:"amount"`
	RealAmount     uint64     `gorm:"column:real_amount;type:int(11);not null" json:"real_amount"`
	AvailableDelta uint64     `gorm:"column:available_delta;type:int(11);not null" json:"available_delta"`
	FrozenDelta    uint64     `gorm:"column:frozen_delta;type:int(11);not null" json:"frozen_delta"`
	ChainID        string     `gorm:"column:chain_id;type:varchar(255);not null" json:"chain_id"`
	TxHash         string     `gorm:"column:tx_hash;type:varchar(255);not null" json:"tx_hash"`
	BlockHeight    int        `gorm:"column:block_height;type:int(11);not null" json:"block_height"`
	LogIndex       int        `gorm:"column:log_index;type:int(11);not null" json:"log_index"`
	BlockTimestamp time.Time  `gorm:"column:block_timestamp;type:datetime;not null" json:"block_timestamp"`
	AddTime        time.Time  `gorm:"column:add_time;type:datetime;not null" json:"add_time"`
	UpdateTime     time.Time  `gorm:"column:update_time;type:datetime;not null" json:"update_time"`
	RefFlowID      uint64     `gorm:"column:ref_flow_id;type:int(11);not null" json:"ref_flow_id"`
	FromAddr       string     `gorm:"column:from_addr;type:varchar(255);not null" json:"from_addr"`
	ToAddr         string     `gorm:"column:to_addr;type:varchar(255);not null" json:"to_addr"`
	LockExpiry     *time.Time `gorm:"column:lock_expiry;type:datetime;not null" json:"lock_expiry"`
	LockNonce      string     `gorm:"column:lock_nonce;type:varchar(255);not null" json:"lock_nonce"`
	LockID         string     `gorm:"column:lock_id;type:varchar(255);not null" json:"lock_id"`
	LockSig        string     `gorm:"column:lock_sig;type:varchar(255);not null" json:"lock_sig"`
	LockAddr       string     `gorm:"column:lock_addr;type:varchar(255);not null" json:"lock_addr"`
}

func (AccountBalanceFlow) TableName() string {
	return "n_account_balance_flow"
}
