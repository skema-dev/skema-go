package database_test

import (
	"os"
	"testing"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const testConfig1 = `
database:
    db1:
        type: memory
        dbname: hello1
    db2:
        type: memory
        dbname: hello2
`

const testConfig2 = `
database:
    db1:
        type: memory
        dbname: hello1
    db2:
        type: sqlite
        filepath: hello2.db
`

type managerTestSuite struct {
	suite.Suite
}

func (s *managerTestSuite) SetupTest() {

}

func (s *managerTestSuite) TestAddDbFromConfig() {
	dbConfig := config.NewConfigWithString(testConfig1)
	configs := dbConfig.GetMapConfig("database")

	dbManager := database.NewDatabaseManager()

	for k, v := range configs {
		dbManager.AddDatabaseWithConfig(&v, k)
	}

	db1 := dbManager.GetDB("db1")
	db2 := dbManager.GetDB("db2")
	assert.NotNil(s.T(), db1)
	assert.NotNil(s.T(), db2)
}

func (s *managerTestSuite) TestCreatSqlitefileFromConfig() {
	os.RemoveAll("hello2.db")

	dbConfig := config.NewConfigWithString(testConfig2)
	dbManager := database.NewDatabaseManager().WithConfig(dbConfig, "database")

	db1 := dbManager.GetDB("db1")
	db2 := dbManager.GetDB("db2")
	assert.NotNil(s.T(), db1)
	assert.NotNil(s.T(), db2)

	_, err := os.Stat("hello2.db")
	assert.Nil(s.T(), err)

	os.RemoveAll("hello2.db")
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, new(managerTestSuite))
}
