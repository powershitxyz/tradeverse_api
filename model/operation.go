package model

import "time"

type PagePush struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"-"`
	SortCode    string    `gorm:"column:sort_code;type:varchar(255);not null" json:"sort_code"`
	ImageUrl    string    `gorm:"column:image_url;type:varchar(255);not null" json:"image_url"`
	Title       string    `gorm:"column:title;type:varchar(255);not null" json:"title"`
	AddTime     time.Time `gorm:"column:add_time;type:datetime;not null" json:"add_time"`
	ObjectID    string    `gorm:"column:object_id;type:varchar(255);not null" json:"object_id"`
	ObjectType  string    `gorm:"column:object_type;type:varchar(255);not null" json:"object_type"`
	HomeVisible int       `gorm:"column:home_visible;type:int(11);not null" json:"home_visible"`
}

func (PagePush) TableName() string {
	return "n_page_push"
}
