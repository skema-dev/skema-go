package data_test

import (
	"fmt"
	"testing"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type Sample struct {
	gorm.Model
	Name   string
	Nation string
}

type databaseTestSuite struct {
	suite.Suite
}

func (s *databaseTestSuite) SetupTest() {

}

func (s *databaseTestSuite) TestMemoryDb() {
	db, err := data.NewMemoryDatabase(nil)
	db.AutoMigrate(&Sample{})

	assert.NotNil(s.T(), db)
	assert.Nil(s.T(), err)

	db.Create(&Sample{Name: "testuser1", Nation: "china"})
	db.Create(&Sample{Name: "testuser2", Nation: "england"})

	var samples []Sample
	result := db.Find(&samples)
	assert.Nil(s.T(), result.Error)
	assert.Equal(s.T(), 2, len(samples))

	fmt.Printf("%v\n%v\n", samples[0], samples[1])

	sample := Sample{}
	db.Where(&Sample{Nation: "china"}).First(&sample)
	assert.Equal(s.T(), "testuser1", sample.Name)

	sample = Sample{}
	db.Where(&Sample{Name: "testuser2"}).First(&sample)
	assert.Equal(s.T(), "england", sample.Nation)
}

func (s *databaseTestSuite) TestMysqlDb() {
	mysqlConfigStr := `
db1:
  type: mysql
  username: root
  password: abcd
  dbname: test
  host: localhost
  port: 3306
  retry: 1
`
	dbConfig := config.NewConfigWithString(mysqlConfigStr)
	db, err := data.NewMysqlDatabase(dbConfig.GetSubConfig("db1"))
	assert.Nil(s.T(), db)
	assert.NotNil(s.T(), err)
}

func TestDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(databaseTestSuite))
}
