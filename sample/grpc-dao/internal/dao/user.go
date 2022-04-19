package dao

import (
	"github.com/skema-dev/skema-go/database"
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

func NewUser() database.DaoModel {
	return &User{}
}
