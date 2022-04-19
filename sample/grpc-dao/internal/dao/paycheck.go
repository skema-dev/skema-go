package dao

import (
	"gorm.io/gorm"
)

type Paycheck struct {
	gorm.Model

	UUID     string `gorm:"column:uuid;index;unique;"`
	UserUUID string
	Amount   int
	Month    int
}

func (Paycheck) TableName() string {
	return "paycheck"
}
