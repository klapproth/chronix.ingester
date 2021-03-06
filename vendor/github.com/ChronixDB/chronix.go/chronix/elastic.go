package chronix

import (
	"encoding/json"
	"fmt"
	"time"
	"github.com/olivere/elastic"
	"context"
	"github.com/prometheus/common/log"
)

type elasticClient struct {
	elastic *elastic.Client
}

// Only for test purposes
func NewElasticTestStorage(url *string) StorageClient {
	client, err := elastic.NewClient(elastic.SetURL(*url), elastic.SetHealthcheck(false), elastic.SetSniff(false))
	if err != nil {
		log.Fatal(fmt.Errorf("error creating elasticserach client: %v", err))
		return nil
	}

	return &elasticClient{
		elastic: client,
	}
}

// NewSolrStorage creates a new Solr client.
func NewElasticStorage(url *string, withIndex *bool, deleteIfExists *bool) StorageClient {

	client, err := elastic.NewClient(elastic.SetURL(*url))
	if err != nil {
		log.Fatal(fmt.Errorf("error creating elasticserach client: %v", err))
		return nil
	}

	if *withIndex {
		configureIndex(client, deleteIfExists)
	}

	return &elasticClient{
		elastic: client,
	}
}

func configureIndex(client *elastic.Client, deleteIfExists *bool) {
	//Delete if exists
	exists, err := client.IndexExists("chronix").Do(context.Background())
	if err != nil {
		log.Fatal(fmt.Errorf("error checking if index 'chronix' exists: %v", err))
	}

	//if the index exists and we should delete the index
	if exists && *deleteIfExists {
		log.Info("Delete index and create a new one")
		client.DeleteIndex("chronix").Do(context.Background())

		mapping :=
			`{"settings":{"number_of_shards":1,"number_of_replicas":0}, "mappings":{	"doc":{
			"properties":{
				"data":{"type":"binary", "doc_values": false},
				"start":{"type":"date", "format": "epoch_millis"},
				"end":{"type":"date", "format": "epoch_millis"},
				"name":{"type":"text"},
				"type":{"type":"text"}
			}}}}`

		createIndex, err := client.CreateIndex("chronix").Body(mapping).Do(context.Background())
		if err != nil {
			// Handle error
			log.Fatal(err)
		}

		if !createIndex.Acknowledged {
			// Not acknowledged
		}
	}
}

// Update implements StorageClient.
func (c *elasticClient) Update(data []map[string]interface{}, commit bool, commitWithin time.Duration) error {

	var bulk = c.elastic.Bulk()

	//loop over the documents
	for k := range data {
		buf, err := json.Marshal(data[k])
		if err != nil {
			return fmt.Errorf("error marshalling JSON: %v", err)
		}

		req := elastic.NewBulkIndexRequest().
			Index("chronix").
			Type("doc").
			Doc(string(buf))

		bulk.Add(req)

	}
	bulk.Do(context.Background())
	return nil
}

func (c *elasticClient) Query(q, cj, fl string) ([]byte, error) {
	return nil, fmt.Errorf("not yet implmented")
}

func (c *elasticClient) NeedPostfixOnDynamicField() bool {
	return false
}
