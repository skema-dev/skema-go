package model

import (
	"github.com/skema-dev/skema-go/data"
)

func init() {
	data.R(&User{})
}

type User struct {
	data.Model

	UUID   string `gorm:"column:uuid;index;unique;"`
	Name   string
	Nation string
	City   string
}

func (User) TableName() string {
	return "user"
}
