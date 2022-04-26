package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/data"
	"github.com/skema-dev/skema-go/logging"
)

type TestData struct {
	data.Model
	Name string
}

func (TestData) TableName() string {
	return "test_data"
}

func main() {
	yamlV7 := `
databases:
  db1:
    type: sqlite
    filepath: default.db
    dbname: test
    automigrate: true
    cqrs:
       type: elastic
       name: elastic-search

elastic-search:
    version: v7
    addresses:
        - http://localhost:9200
`
	yamlV8 := `
databases:
  db1:
    type: sqlite
    filepath: default.db
    dbname: test
    automigrate: true
    cqrs:
       type: elastic
       name: elastic-search

elastic-search:
    version: v7
    addresses:
        - http://localhost:9200
`

	version := flag.String("version", "v7", "specify elastic version: v7 or v8")
	flag.Parse()

	s := yamlV7
	if *version == "v8" {
		s = yamlV8
	}

	os.RemoveAll("./default.db")
	data.InitWithConfig(config.NewConfigWithString(s), "databases")
	dao := data.Manager().GetDAO(&TestData{})
	indexName := dao.GetDB().Name() + "_" + TestData{}.TableName()
	dao.DeleteFromElastic([]string{indexName})

	// no es enabled
	data1 := TestData{Name: "user1"}
	data1.UUID = "10-aaaaaa-bbbbbb-1"
	// var err error
	result := []TestData{}

	dao.Create(&data1)

	fmt.Printf("******************Start Update database*******************\n")
	result = []TestData{}
	dao.Query(&data.QueryParams{"uuid": data1.UUID}, &result)
	fmt.Printf("%v\n", result)

	data1.Name = "user1_3"
	dao.Update(&data.QueryParams{"uuid": data1.UUID}, &data1)

	fmt.Printf("******************Update Done!!*******************\n")

	dao.Query(&data.QueryParams{"uuid": data1.UUID}, &result)
	stringEquals("user1_3", result[0].Name)

	// fmt.Printf("******************With Elasticsearch*******************\n")
	data1.Name = "user1_2"
	dao.Update(&data.QueryParams{"uuid": data1.UUID}, &data1)
	dao.Query(&data.QueryParams{"uuid": data1.UUID}, &result)
	stringEquals("user1_2", result[0].Name)

	fmt.Printf("All Done!\n")

}

func stringEquals(expected string, actual string) {
	if expected != actual {
		logging.Fatalf("Expected: %s, Actual: %s", expected, actual)
	}
}

func intEquals(expected int, actual int) {
	if expected != actual {
		logging.Fatalf("Expected: %d, Actual: %d", expected, actual)
	}
}
