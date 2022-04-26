package model

import "github.com/skema-dev/skema-go/data"

func init() {
	data.R(&Paycheck{})
}

type Paycheck struct {
	data.Model

	UUID     string `gorm:"column:uuid;index;unique;"`
	UserUUID string
	Amount   int
	Month    int
}

func (Paycheck) TableName() string {
	return "paycheck"
}
