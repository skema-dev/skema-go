package model

import (
	"gorm.io/gorm"
)

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
