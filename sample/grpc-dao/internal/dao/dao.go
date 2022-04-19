package dao

import (
	"github.com/skema-dev/skema-go/database"
)

func Register() {
	database.Manager().RegisterDaoModels([]database.DaoModel{User{}, Paycheck{}, Address{}})
}
