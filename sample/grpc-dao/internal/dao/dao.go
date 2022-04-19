package dao

import (
	"github.com/skema-dev/skema-go/database"
)

func init() {
	database.Manager().RegisterDaoModels("", []database.DaoModel{Address{}})
}
