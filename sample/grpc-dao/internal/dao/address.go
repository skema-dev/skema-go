package dao

import (
	"gorm.io/gorm"
)

type Address struct {
	gorm.Model

	UserUUID string `gorm:"column:user_uuid;index;unique;"`
	Country  string
	State    string
	City     string
	Street   string
	Building string
	Location string
}

func (Address) TableName() string {
	return "address"
}
