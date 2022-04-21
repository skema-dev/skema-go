package data_test

import (
	"os"
	"testing"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
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
        dbname: hello3
        automigrate: true
    db2:
        type: sqlite
        filepath: hello4.db
        automigrate: true
`

type TestModel1 struct {
	gorm.Model
	Name string
}

func (TestModel1) TableName() string {
	return "test1"
}

type TestModel2 struct {
	gorm.Model
	Name string
}

func (TestModel2) TableName() string {
	return "test2"
}

type TestModel3 struct {
	gorm.Model
	Name string
}

func (TestModel3) TableName() string {
	return "test3"
}

type managerTestSuite struct {
	suite.Suite
}

func (s *managerTestSuite) SetupTest() {

}

func (s *managerTestSuite) TestAllInSequence() {
	s.testAddDbFromConfig()
	s.testCreatSqlitefileFromConfig()
	s.testCreatDAO()
}

func (s *managerTestSuite) testAddDbFromConfig() {
	dbConfig := config.NewConfigWithString(testConfig1)
	configs := dbConfig.GetMapConfig("database")

	dbManager := data.NewDataManager()

	for k, v := range configs {
		dbManager.AddDatabaseWithConfig(&v, k)
	}

	db1 := dbManager.GetDB("db1")
	db2 := dbManager.GetDB("db2")
	assert.NotNil(s.T(), db1)
	assert.NotNil(s.T(), db2)
}

func (s *managerTestSuite) testCreatSqlitefileFromConfig() {
	os.RemoveAll("hello4.db")

	dbConfig := config.NewConfigWithString(testConfig2)
	dbManager := data.NewDataManager().WithConfig(dbConfig, "database")

	db1 := dbManager.GetDB("db1")
	db2 := dbManager.GetDB("db2")
	assert.NotNil(s.T(), db1)
	assert.NotNil(s.T(), db2)

	_, err := os.Stat("hello4.db")
	assert.Nil(s.T(), err)

	os.RemoveAll("hello4.db")
}

func (s *managerTestSuite) testCreatDAO() {
	os.RemoveAll("hello4.db")

	dbConfig := config.NewConfigWithString(testConfig2)
	data.InitWithConfig(dbConfig, "database")

	data.Manager().RegisterDaoModelsForDb("db2", []data.DaoModel{TestModel1{}, TestModel2{}})

	dao := data.Manager().GetDaoForDb("db2", TestModel1{})
	dao.Upsert(&TestModel1{
		Name: "test1",
	}, nil, nil)
	dao.Upsert(&TestModel1{
		Name: "test2",
	}, nil, nil)

	result := []TestModel1{}
	dao.Query(&data.QueryParams{}, &result)
	assert.Equal(s.T(), 2, len(result))

	os.RemoveAll("hello4.db")
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, new(managerTestSuite))
}