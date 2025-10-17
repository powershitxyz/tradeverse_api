package model

import (
	"time"
)

type BaseM struct {
	Id uint64 `gorm:"AUTO_INCREMENT;PRIMARY_KEY"`
}

type WithTime struct {
	BaseM
	AddTime    time.Time
	UpdateTime time.Time
}
