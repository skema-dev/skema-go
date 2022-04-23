package elastic

import (
	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
)

type Elastic interface {
	Index(index string, id string, value interface{}) error
	Search(index string, query map[string]interface{}) ([]map[string]interface{}, error)
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
