package tools

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Time 自定义时间类型，基于 time.Time，用于 GORM 与 JSON 序列化
type Time time.Time

// Scan 实现 sql.Scanner，供 GORM 从数据库读取
func (t *Time) Scan(value interface{}) error {
	if value == nil {
		*t = Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*t = Time(v)
		return nil
	case []byte:
		parsed, err := time.ParseInLocation("2006-01-02 15:04:05", string(v), time.Local)
		if err != nil {
			parsed, err = time.Parse(time.RFC3339, string(v))
			if err != nil {
				return err
			}
		}
		*t = Time(parsed)
		return nil
	case string:
		parsed, err := time.ParseInLocation("2006-01-02 15:04:05", v, time.Local)
		if err != nil {
			parsed, err = time.Parse(time.RFC3339, v)
			if err != nil {
				return err
			}
		}
		*t = Time(parsed)
		return nil
	default:
		return nil
	}
}

// Value 实现 driver.Valuer，供 GORM 写入数据库
func (t Time) Value() (driver.Value, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return nil, nil
	}
	return tt, nil
}

// MarshalJSON 实现 json.Marshaler，统一输出为 RFC3339
func (t Time) MarshalJSON() ([]byte, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(tt.Format(time.RFC3339))
}

// UnmarshalJSON 实现 json.Unmarshaler
func (t *Time) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" || s == "null" {
		*t = Time{}
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		parsed, err = time.Parse("2006-01-02 15:04:05", s)
		if err != nil {
			return err
		}
	}
	*t = Time(parsed)
	return nil
}

// ToTime 转为标准库 time.Time
func (t Time) ToTime() time.Time { return time.Time(t) }

// FromTime 从标准库 time.Time 构造
func FromTime(tt time.Time) Time { return Time(tt) }
