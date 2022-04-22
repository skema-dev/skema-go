package model

import (
	"gorm.io/gorm"
)

func init() {
	// data.RegisterModelType(&User{})
}

type User struct {
	gorm.Model

	UUID   string `gorm:"column:uuid;index;unique;"`
	Name   string
	Nation string
	City   string
}

func (User) TableName() string {
	return "user"
}
