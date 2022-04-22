package model

import (
	"github.com/skema-dev/skema-go/data"
	"gorm.io/gorm"
)

func init() {
	data.RegisterModelType(&User{})
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
