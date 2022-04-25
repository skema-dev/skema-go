package model

import "github.com/skema-dev/skema-go/data"

func init() {
	data.R(&Address{})
}

type Address struct {
	data.Model

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
