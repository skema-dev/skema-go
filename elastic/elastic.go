package elastic

import (
	"bytes"
	"encoding/json"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
)

type Elastic interface {
	Index(index string, id string, value interface{}) error
	Search(index string, termQueryType string, query map[string]interface{}) ([]map[string]interface{}, error)
	Delete(index string, ids []string)
	DeleteIndex(indexes []string)
}

func NewElasticClient(conf *config.Config) Elastic {
	var result Elastic
	version := conf.GetString("version", "v8")
	switch version {
	case "v8":
		result = newElasticClientV8(conf)
	case "v7":
		result = newElasticClientV7(conf)
	default:
		logging.Fatalf("unsupported elastic version %s", version)
	}

	return result
}

func buildTermQuery(queryType string, query map[string]interface{}) (string, error) {
	var buf bytes.Buffer
	condition := map[string]interface{}{
		"query": map[string]interface{}{
			queryType: query,
		},
	}

	if err := json.NewEncoder(&buf).Encode(condition); err != nil {
		return "", logging.Errorf(err.Error())
	}

	return buf.String(), nil
}

func processSearchResult(res map[string]interface{}) ([]map[string]interface{}, error) {
	h := res["hits"].(map[string]interface{})
	hits := h["hits"].([]interface{})

	result := []map[string]interface{}{}
	for _, hit := range hits {
		hitData := hit.(map[string]interface{})
		result = append(result, hitData["_source"].(map[string]interface{}))
	}
	return result, nil
}

func ConvertMapToStruct(value map[string]interface{}, target interface{}) error {
	jsonBody, err := json.Marshal(value)
	if err != nil {
		return logging.Errorf(err.Error())
	}

	return json.Unmarshal(jsonBody, target)
}
