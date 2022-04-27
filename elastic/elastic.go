package elastic

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
)

type SearchOption struct {
	Sort string
	Size int
	From int
}
type Elastic interface {
	Index(index string, id string, value interface{}) error
	Search(index string, termQueryType string, query map[string]interface{}, option *SearchOption) ([]map[string]interface{}, error)
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

func createSortCondition(order string) []map[string]string {
	result := []map[string]string{}

	conds := strings.Split(order, ",")
	for _, con := range conds {
		con = strings.Trim(con, " ")
		ss := strings.Split(con, " ")
		i := 0
		k := ""
		m := map[string]string{}

		for _, s := range ss {
			s = strings.Trim(s, " ")
			if len(s) == 0 {
				continue
			}

			switch i {
			case 0:
				k = s
				i++
			case 1:
				m[k] = s
				break
			}
		}
		if _, ok := m[k]; !ok {
			// key is specified but no further order description. Using default desc order
			m[k] = "desc"
		}

		result = append(result, m)
	}
	return result
}

func buildTermQuery(queryType string, query map[string]interface{}, option *SearchOption) (string, error) {
	var buf bytes.Buffer

	// 	"sort" : [
	//     { "post_date" : {"order" : "asc", "format": "strict_date_optional_time_nanos"}},
	//     "user",
	//     { "name" : "desc" },
	//     { "age" : "desc" },
	//     "_score"
	//   ],
	condition := map[string]interface{}{
		"query": map[string]interface{}{
			queryType: query,
		},
	}

	if option != nil {
		if option.Sort != "" {
			sortCondition := createSortCondition(option.Sort)
			if len(sortCondition) > 0 {
				condition["sort"] = sortCondition
			}
		}

		if option.From > 0 {
			condition["from"] = option.From
		}

		if option.Size > 0 {
			condition["size"] = option.Size
		}
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
