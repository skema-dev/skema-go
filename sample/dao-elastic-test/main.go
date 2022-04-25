package main

import (
	"flag"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/data"
	"github.com/skema-dev/skema-go/elastic"
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
	dbYaml := `
databases:
  db1:
    type: sqlite
    filepath: default.db
    dbname: test
    automigrate: true
`
	yamlWithCert := `
elastic:
    version: v8
    addresses:
        - https://localhost:9200
    username: elastic
    password: sN89iAwxjbyaj=ptTeaP
    cert: ./http_ca.crt
`
	yamlDefault := `
elastic:
    version: v7
    addresses:
        - http://localhost:9200
`
	version := flag.String("version", "v7", "specify elasticsearch version: v7 or v8")
	var client elastic.Elastic

	flag.Parse()

	data.InitWithConfig(config.NewConfigWithString(dbYaml), "databases")
	dao := data.Manager().GetDAO(TestData{}, true)

	switch *version {
	case "v8":
		client = elastic.NewElasticClient(config.NewConfigWithString(yamlWithCert).GetSubConfig("elastic"))
	case "v7":
		client = elastic.NewElasticClient(config.NewConfigWithString(yamlDefault).GetSubConfig("elastic"))
	default:
		logging.Fatalf("version must be v7 or v8")
	}
	dao.SetElasticClient(client)

	data1 := TestData{Model: data.Model{UUID: "aaaaaa-bbbbbb"}, Name: "user1"}
	data2 := TestData{Model: data.Model{UUID: "aaaaaa-bbbbbb-2"}, Name: "user2"}
	data3 := TestData{Model: data.Model{UUID: "aaaaaa-bbbbbb-3"}, Name: "user3"}

	indexName := dao.GetDB().Name() + "_" + TestData{}.TableName()
	err := client.Index(indexName, "aaaaa1", &data1)
	if err != nil {
		logging.Fatalf(err.Error())
	}

	err = client.Index(indexName, "aaaaa2", &data2)
	err = client.Index(indexName, "aaaaa3", &data3)

	result := []TestData{}
	err = dao.Query(&data.QueryParams{"name": "user2"}, &result)
	if err != nil {
		logging.Fatalf(err.Error())
	}

	intEquals(1, len(result))
	stringEquals("user2", result[0].Name)
	stringEquals("aaaaaa-bbbbbb-2", result[0].UUID)

	result = []TestData{}

	dao.Update(&data.QueryParams{"uuid": data1.UUID}, data1)
	err = dao.Query(&data.QueryParams{"uuid": data1.UUID}, &result)
	stringEquals("user1", result[0].Name)

	data1.Name = "user1_1"
	dao.Update(&data.QueryParams{"uuid": data1.UUID}, data1)
	err = dao.Query(&data.QueryParams{"uuid": data1.UUID}, &result)
	stringEquals("user1_1", result[0].Name)
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
