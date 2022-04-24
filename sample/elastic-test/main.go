package main

import (
	"flag"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/elastic"
	"github.com/skema-dev/skema-go/logging"
)

type TestData struct {
	UUID string
	Name string
}

func main() {
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

	switch *version {
	case "v8":
		client = elastic.NewElasticClient(config.NewConfigWithString(yamlWithCert).GetSubConfig("elastic"))
	case "v7":
		client = elastic.NewElasticClient(config.NewConfigWithString(yamlDefault).GetSubConfig("elastic"))
	default:
		logging.Fatalf("version must be v7 or v8")
	}

	err := client.Index("test1", "aaaaa1", &TestData{UUID: "aaaaaa-bbbbbb", Name: "user1"})
	if err != nil {
		logging.Fatalf(err.Error())
	}

	err = client.Index("test1", "aaaaa2", &TestData{UUID: "aaaaaa-bbbbbb-2", Name: "user2"})
	err = client.Index("test1", "aaaaa3", &TestData{UUID: "aaaaaa-bbbbbb-2", Name: "user3"})

	result, _ := client.Search("test1", "match", map[string]interface{}{"Name": "user1"})
	stringEquals("aaaaaa-bbbbbb", result[0]["UUID"].(string))
	stringEquals("user1", result[0]["Name"].(string))

	result, _ = client.Search("test1", "wildcard", map[string]interface{}{"Name": "use*"})
	intEquals(3, len(result))

	result, _ = client.Search("test1", "terms", map[string]interface{}{"Name": []string{"user1", "user3"}})
	intEquals(2, len(result))
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
