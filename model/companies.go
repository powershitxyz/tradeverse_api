package model

import "chaos/api/tools"

// Company 对应表 "companies" 的 GORM 模型
type Company struct {
	Id          int         `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	Name        string      `gorm:"column:name;size:255;not null" json:"name"`
	ShortName   *string     `gorm:"column:short_name;size:100" json:"short_name"`
	IsPublic    *int        `gorm:"column:is_public" json:"is_public"`
	Symbol      *string     `gorm:"column:symbol;size:20" json:"symbol"`
	Exchange    *string     `gorm:"column:exchange;size:50" json:"exchange"`
	Sector      *string     `gorm:"column:sector;size:100" json:"sector"`
	Description *string     `gorm:"column:description" json:"description"`
	Website     *string     `gorm:"column:website;size:255" json:"website"`
	CreatedAt   *tools.Time `gorm:"column:created_at" json:"created_at"`
}

func (Company) TableName() string { return "companies" }
