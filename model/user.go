package model

import "time"

const (
	PROVIDER_TYPE_WALLET = "wallet"
	PROVIDER_TYPE_GOOGLE = "google"
)

type UserMain struct {
	ID     uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserNo string `gorm:"column:user_no;type:varchar(255);not null" json:"user_no"`
	// Email    string    `gorm:"column:email;type:varchar(255);not null" json:"email"`
	Email    *string   `gorm:"column:email;type:varchar(255);uniqueIndex" json:"email"`
	Password string    `gorm:"column:password;type:varchar(255);not null" json:"-"`
	AddTime  time.Time `gorm:"column:add_time" json:"add_time"`
	Status   string    `gorm:"column:status" json:"status"`
	RefID    uint64    `gorm:"column:ref_id;type:int(11);not null" json:"ref_id"`
}

func (UserMain) TableName() string {
	return "n_user_main"
}

type UserProvider struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MainID        uint64    `gorm:"column:main_id;type:int(11);not null" json:"main_id"`
	ProviderType  string    `gorm:"column:provider_type;type:varchar(255);not null" json:"provider_type"`
	ProviderID    string    `gorm:"column:provider_id;type:varchar(255);not null" json:"provider_id"`
	AddTime       time.Time `gorm:"column:add_time" json:"add_time"`
	ProviderLabel string    `gorm:"column:provider_label;type:varchar(255);not null" json:"provider_label"`
}

func (UserProvider) TableName() string {
	return "n_user_provider"
}

type UserProfile struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement" json:"-"`
	MainID      uint64 `gorm:"column:main_id;type:int(11);not null" json:"-"`
	Name        string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Avatar      string `gorm:"column:avatar;type:varchar(255);not null" json:"avatar"`
	Bio         string `gorm:"column:bio;type:varchar(255);not null" json:"bio"`
	Birthday    string `gorm:"column:birthday;type:datetime;not null" json:"birthday"`
	CountryCode string `gorm:"column:country_code;type:varchar(255);not null" json:"country_code"`
	Timezone    int    `gorm:"column:timezone;type:int(11);not null" json:"timezone"`
	XUri        string `gorm:"column:x_uri;type:varchar(255);not null" json:"x_uri"`
}

func (UserProfile) TableName() string {
	return "n_user_profile"
}

type VerificationProcess struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Target         string    `gorm:"column:target" json:"target"`
	Type           string    `gorm:"column:type" json:"type"`
	Code           string    `gorm:"column:code" json:"code"`
	AddTime        time.Time `gorm:"column:add_time" json:"add_time"`
	ValidatePeriod int64     `gorm:"column:validate_period" json:"validate_period"`
	Sort           string    `gorm:"column:sort" json:"sort"`
	Status         string    `gorm:"column:status" json:"status"`
	MainID         uint64    `gorm:"column:main_id" json:"main_id"`
}

func (VerificationProcess) TableName() string {
	return "n_verification_process"
}

type UserRef struct {
	ID      uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MainID  uint64    `gorm:"column:main_id;type:int(11);not null" json:"main_id"`
	RefCode string    `gorm:"column:ref_code;type:varchar(255);not null" json:"ref_code"`
	AddTime time.Time `gorm:"column:add_time" json:"add_time"`
}

func (UserRef) TableName() string {
	return "n_user_ref"
}
