package elastic

import (
	"context"
	"encoding/json"
	"io/ioutil"
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

	return &elasticClientV7{
		client: esclient,
	}
}

func (e *elasticClientV7) Index(index string, id string, value interface{}) error {
	if index == "" || id == "" {
		return logging.Errorf("index and id should not be empty. index: %s, id: %s", index, id)
	}
	data, err := json.Marshal(value)
	if err != nil {
		return logging.Errorf(err.Error())
	}

	s := string(data)
	logging.Debugw("elastic index request", "index", index, "id", id, "body", s)

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

func (e *elasticClientV7) Search(index string, termQueryType string, query map[string]interface{}, option *SearchOption) ([]map[string]interface{}, error) {
	searchQuery, err := buildTermQuery(termQueryType, query, option)
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
	if res.IsError() {
		return nil, logging.Errorf("Error happend for search %v", res)
	}

	resMap := map[string]interface{}{}
	err = json.NewDecoder(res.Body).Decode(&resMap)
	if err != nil {
		return nil, logging.Errorf(err.Error())
	}

	return processSearchResult(resMap)
}

func (e *elasticClientV7) Delete(index string, ids []string) {
	searchQuery, err := buildTermQuery("terms", map[string]interface{}{"id": ids}, nil)
	logging.Debugw("Delete es docs", "index", index, "ids", ids)
	_, err = e.client.DeleteByQuery([]string{index}, strings.NewReader(searchQuery))
	if err != nil {
		logging.Errorf("failded to deletes: %s", err.Error())
		return
	}
}

func (e *elasticClientV7) DeleteIndex(indexes []string) {
	req := esapi.IndicesDeleteRequest{
		Index: indexes,
	}

	_, err := req.Do(context.Background(), e.client)
	if err != nil {
		logging.Errorf(err.Error())
		return
	}

	logging.Debugf("index deleted %d", len(indexes))
}
