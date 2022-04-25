package data_test

import (
	"os"
	"testing"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/data"
	db "github.com/skema-dev/skema-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SampleModel struct {
	data.Model
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
	s.testCreateAndUpdateDAO()
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

func (s *daoTestSuite) testCreateAndUpdateDAO() {
	yaml := `
type: sqlite
filepath: './test1.db'
`
	dbInstance, _ := db.NewSqliteDatabase(config.NewConfigWithString(yaml))
	dao := db.NewDAO(dbInstance, &SampleModel{})
	dao.Automigrate()
	assert.NotNil(s.T(), dao)

	dao.Create(&SampleModel{Name: "user1", Sex: "female", Nation: "china", City: "shenzhen"})
	dao.Create(&SampleModel{Name: "user2", Sex: "female", Nation: "japan", City: "tokyo"})
	dao.Create(&SampleModel{Name: "user3", Sex: "female", Nation: "france", City: "paris"})
	dao.Create(&SampleModel{Name: "user4", Sex: "male", Nation: "china", City: "shanghai"})
	dao.Create(&SampleModel{Name: "user5", Sex: "male", Nation: "china", City: "shenzhen"})

	rs := []SampleModel{}
	dao.Query(&db.QueryParams{}, &rs)
	assert.Equal(s.T(), 5, len(rs))

	err := dao.Create(&SampleModel{Name: "user4", Sex: "male", Nation: "us", City: "seattle"})
	assert.NotNil(s.T(), err)

	err = dao.Update(&db.QueryParams{"name": "user4", "sex": "male"}, &SampleModel{Sex: "female"})
	assert.Nil(s.T(), err)

	rs = []SampleModel{}
	dao.Query(&db.QueryParams{"name": "user4"}, &rs)
	assert.Equal(s.T(), 1, len(rs))
	assert.Equal(s.T(), "female", rs[0].Sex)

	err = dao.Update(&db.QueryParams{"name": "user10", "sex": "male"}, &SampleModel{Sex: "female"})
	assert.NotNil(s.T(), err)

	os.RemoveAll("./test1.db")
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
