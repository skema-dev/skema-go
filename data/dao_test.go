package data_test

import (
	"os"
	"testing"

	"github.com/skema-dev/skema-go/config"
	db "github.com/skema-dev/skema-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type SampleModel struct {
	gorm.Model
	Name   string `gorm:"type:varchar(100);uniqueIndex:unique_name_sex;"`
	Sex    string `gorm:"type:varchar(32);uniqueIndex:unique_name_sex;"`
	Nation string
	City   string
}

func (SampleModel) TableName() string {
	return "sample"
}

type daoTestSuite struct {
	suite.Suite
}

func (s *daoTestSuite) SetupTest() {

}

func (s *daoTestSuite) TestDAOInSequence() {
	s.testSampleDAO()
	s.testDeleteDAO()
}

func (s *daoTestSuite) testSampleDAO() {
	dbInstance, _ := db.NewMemoryDatabase(nil)
	dao := db.NewDAO(dbInstance, &SampleModel{})
	dao.Automigrate()
	assert.NotNil(s.T(), dao)

	dao.Upsert(&SampleModel{Name: "user1", Sex: "male", Nation: "china", City: "shanghai"}, nil, nil)
	dao.Upsert(&SampleModel{Name: "user2", Sex: "female", Nation: "england", City: "london"}, nil, nil)

	var results []SampleModel

	err := dao.Query(&db.QueryParams{"name": "user1"}, &results)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, len(results))
	assert.Equal(s.T(), "china", results[0].Nation)

	err = dao.Query(&db.QueryParams{"nation": "england"}, &results)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, len(results))
	assert.Equal(s.T(), "user2", results[0].Name)

	err = dao.Query(&db.QueryParams{}, &results)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 2, len(results))

	err = dao.Upsert(
		&SampleModel{Name: "user2", Sex: "female", Nation: "japan", City: "tokyo"},
		[]string{"name", "sex"},
		[]string{"nation"},
	)
	assert.Nil(s.T(), err)

	err = dao.Query(&db.QueryParams{"name": "user2"}, &results)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, len(results))
	assert.Equal(s.T(), "japan", results[0].Nation)
	assert.Equal(s.T(), "london", results[0].City)

	err = dao.Upsert(
		&SampleModel{Name: "user2", Sex: "female", Nation: "usa", City: "san francisco"},
		[]string{"name", "sex"},
		[]string{"nation", "city"},
	)
	assert.Nil(s.T(), err)

	err = dao.Query(&db.QueryParams{"name": "user2"}, &results)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, len(results))
	assert.Equal(s.T(), "usa", results[0].Nation)
	assert.Equal(s.T(), "san francisco", results[0].City)
}

func (s *daoTestSuite) testDeleteDAO() {
	yaml := `
type: sqlite
filepath: './test.db'
`
	dbInstance, _ := db.NewSqliteDatabase(config.NewConfigWithString(yaml))
	dao := db.NewDAO(dbInstance, &SampleModel{})
	dao.Automigrate()
	assert.NotNil(s.T(), dao)

	dao.Upsert(&SampleModel{Name: "user1", Sex: "male", Nation: "china", City: "shanghai"}, nil, nil)
	dao.Upsert(&SampleModel{Name: "user2", Sex: "female", Nation: "england", City: "london"}, nil, nil)
	dao.Upsert(&SampleModel{Name: "user3", Sex: "female", Nation: "france", City: "paris"}, nil, nil)
	dao.Upsert(&SampleModel{Name: "user4", Sex: "female", Nation: "brazil", City: "san paulo"}, nil, nil)

	samples := []SampleModel{}
	dao.Query(&db.QueryParams{"name": "user1"}, &samples)
	assert.Equal(s.T(), 1, len(samples))

	dao.Delete(&samples[0])
	dao.Query(&db.QueryParams{"name": "user1"}, &samples)
	assert.Equal(s.T(), 0, len(samples))

	dao.Query(&db.QueryParams{}, &samples)
	assert.Equal(s.T(), 3, len(samples))

	err := dao.Delete("name like 'user%'")
	assert.Nil(s.T(), err)
	dao.Query(&db.QueryParams{}, &samples)
	assert.Equal(s.T(), 0, len(samples))

	os.RemoveAll("./test.db")
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, new(daoTestSuite))
}
