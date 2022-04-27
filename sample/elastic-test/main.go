package main

import (
	"flag"
	"fmt"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/elastic"
	"github.com/skema-dev/skema-go/logging"
)

type TestData struct {
	Id   int
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

	err := client.Index("test1", "aaaaa1", &TestData{Id: 100, UUID: "aaaaaa-bbbbbb", Name: "user1"})
	if err != nil {
		logging.Fatalf(err.Error())
	}

	err = client.Index("test1", "aaaaa2", &TestData{Id: 200, UUID: "aaaaaa-bbbbbb-2", Name: "user2"})
	err = client.Index("test1", "aaaaa3", &TestData{Id: 300, UUID: "aaaaaa-bbbbbb-2", Name: "user3"})

	result, _ := client.Search("test1", "match", map[string]interface{}{"Name": "user1"}, nil)
	stringEquals("aaaaaa-bbbbbb", result[0]["UUID"].(string))
	stringEquals("user1", result[0]["Name"].(string))

	result, _ = client.Search("test1", "wildcard", map[string]interface{}{"Name": "use*"}, nil)
	intEquals(3, len(result))

	result, _ = client.Search("test1", "terms", map[string]interface{}{"Name": []string{"user1", "user3"}}, nil)
	intEquals(2, len(result))

	result, _ = client.Search("test1", "match", map[string]interface{}{"Name": "1234"}, nil)
	intEquals(0, len(result))

	result, _ = client.Search("test1", "wildcard", map[string]interface{}{"Name": "use*"}, &elastic.SearchOption{Size: 1})
	intEquals(1, len(result))

	// result should be nil since name cannot be sorted (not keyword), but not breaking the test
	result, _ = client.Search("test1", "wildcard", map[string]interface{}{"Name": "use*"}, &elastic.SearchOption{Sort: "Name", Size: 1})
	if result != nil {
		panic("should return nil")
	}

	// ok to index on numeric fields
	result, _ = client.Search("test1", "wildcard", map[string]interface{}{"Name": "use*"}, &elastic.SearchOption{Sort: "Id desc", Size: 1})
	stringEquals("user3", result[0]["Name"].(string))

	result, _ = client.Search("test1", "wildcard", map[string]interface{}{"Name": "use*"}, &elastic.SearchOption{From: 1, Size: 1})
	intEquals(1, len(result))
	stringEquals("user2", result[0]["Name"].(string))

	fmt.Printf("ALL TESTS DONE\n")
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
