package elastic_test

import (
	"testing"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/elastic"
	"github.com/stretchr/testify/assert"
)

type TestData struct {
	UUID string
	Name string
}

// You need to have elasticsearch runnig first, preferrably v7 since it's not cert enabled by default
func TestElasticSearch(t *testing.T) {
	yaml := `
elastic:
    version: v7
    addresses:
        - http://localhost:9200
`
	client := elastic.NewElasticClient(config.NewConfigWithString(yaml).GetSubConfig("elastic"))

	err := client.Index("test1", "aaaaa1", &TestData{UUID: "aaaaaa-bbbbbb", Name: "user1"})
	assert.Nil(t, err)

	err = client.Index("test1", "aaaaa2", &TestData{UUID: "aaaaaa-bbbbbb-2", Name: "user2"})

	result, _ := client.Search("test1", map[string]interface{}{"Name": "user1"})
	assert.Equal(t, "aaaaaa-bbbbbb", result[0]["UUID"].(string))
	assert.Equal(t, "user1", result[0]["Name"].(string))
}
