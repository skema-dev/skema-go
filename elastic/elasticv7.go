package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
)

type elasticClientV7 struct {
	client *es.Client
}

func newElasticClientV7(conf *config.Config) *elasticClientV7 {
	addresses := conf.GetStringArray("addresses")
	cfg := es.Config{
		Addresses: addresses,
	}

	username := conf.GetString("username")
	password := conf.GetString("password")
	if username != "" && password != "" {
		cfg.Username = username
		cfg.Password = password
	}

	esclient, err := es.NewClient(cfg)
	if err != nil {
		logging.Errorf(err.Error())
		return nil
	}
	esclient.Info()
	if info, err := esclient.Info(); err == nil {
		logging.Infof(info.String())
	} else {
		logging.Errorf(err.Error())
	}

	return &elasticClientV7{
		client: esclient,
	}
}

func (e *elasticClientV7) Index(index string, id string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return logging.Errorf(err.Error())
	}

	s := string(data)
	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       strings.NewReader(s),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), e.client)
	if err != nil {
		return logging.Errorf(err.Error())
	}
	defer res.Body.Close()

	if res.IsError() {
		return logging.Errorf("Elasticsearch indexing error for document id=%s: %s", id, res.Status())
	}

	return nil
}

func (e *elasticClientV7) Search(index string, query map[string]interface{}) ([]map[string]interface{}, error) {
	var buf bytes.Buffer
	condition := map[string]interface{}{
		"query": map[string]interface{}{
			"match": query,
		},
	}
	if err := json.NewEncoder(&buf).Encode(condition); err != nil {
		return nil, logging.Errorf(err.Error())
	}

	fmt.Printf("Query: %s\n", buf.String())

	res, err := e.client.Search(
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(index),
		e.client.Search.WithBody(&buf),
		e.client.Search.WithTrackTotalHits(true),
		e.client.Search.WithPretty(),
	)
	if err != nil {
		return nil, logging.Errorf(err.Error())
	}

	resMap := map[string]interface{}{}
	err = json.NewDecoder(res.Body).Decode(&resMap)
	if err != nil {
		return nil, logging.Errorf(err.Error())
	}

	h := resMap["hits"].(map[string]interface{})
	hits := h["hits"].([]interface{})

	result := []map[string]interface{}{}
	for _, hit := range hits {
		hitData := hit.(map[string]interface{})
		result = append(result, hitData["_source"].(map[string]interface{}))
	}

	return result, nil
}
