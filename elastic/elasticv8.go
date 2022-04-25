package elastic

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strings"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
)

type elasticClientV8 struct {
	client *es.Client
}

func newElasticClientV8(conf *config.Config) *elasticClientV8 {
	var err error
	addresses := conf.GetStringArray("addresses")
	cfg := es.Config{
		Addresses: addresses,
	}

	username := conf.GetString("username")
	password := conf.GetString("password")
	if username != "" && password != "" {
		cfg.Username = username
		cfg.Password = password
		logging.Debugf("Elastic Account  %s:%s", username, password)
	}

	var cert []byte
	if certFile := conf.GetString("cert"); certFile != "" {
		cert, err = ioutil.ReadFile("./http_ca.crt")
		if err != nil {
			logging.Fatalf("ERROR: Unable to read CA from %q: %s", certFile, err)
		}
		cfg.CACert = cert
	}

	esclient, err := es.NewClient(cfg)
	if err != nil {
		logging.Errorf("Failed creating es client: %s", err.Error())
		return nil
	}

	if info, err := esclient.Info(); err == nil {
		logging.Infof(info.String())
	} else {
		logging.Errorf(err.Error())
	}

	return &elasticClientV8{
		client: esclient,
	}
}

func (e *elasticClientV8) Index(index string, id string, value interface{}) error {
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

func (e *elasticClientV8) Search(index string, termQueryType string, query map[string]interface{}) ([]map[string]interface{}, error) {
	searchQuery, err := buildTermQuery(termQueryType, query)
	if err != nil {
		return nil, logging.Errorf(err.Error())
	}

	logging.Debugf("Search Query: %s", searchQuery)

	res, err := e.client.Search(
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(index),
		e.client.Search.WithBody(strings.NewReader(searchQuery)),
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

	return processSearchResult(resMap)
}

func (e *elasticClientV8) Delete(index string, id string) {
	_, err := e.client.Delete(index, id)
	if err != nil {
		logging.Errorf("failded to delete %s_%s: %s", index, id, err.Error())
		return
	}
}
