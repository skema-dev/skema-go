package model

import (
	"github.com/skema-dev/skema-go/data"
)

func Register() {
	data.Manager().RegisterDaoModels([]data.DaoModel{
		User{},
		Paycheck{},
		Address{},
	})
}
