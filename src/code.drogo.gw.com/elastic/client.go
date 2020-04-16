package elastic

import (
	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
	"context"
	"github.com/spf13/viper"
	"fmt"
	"encoding/json"
)

type ElasticSearch struct {
	addr   string
	client *elastic.Client
	ctx    context.Context
}

func NewElasticSearch(context context.Context) (*ElasticSearch, error) {

	client, err := getClient()

	if err != nil {
		return nil, err
	}

	es := &ElasticSearch{
		addr:   "http://" + viper.GetString("ES_HOST") + ":" + viper.GetString("ES_PORT"),
		client: client,
		ctx:    context,
	}

	return es, nil
}

func getClient() (*elastic.Client, error) {
	return elastic.NewClient(
		elastic.SetURL("http://"+viper.GetString("ES_HOST")+":"+viper.GetString("ES_PORT")),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
	)
}

func (es *ElasticSearch) DeactivateEventByID(eventId string) error {

	res, err :=  es.client.Update().
		Index(viper.GetString("ES_INDEX")).
		Type(viper.GetString( "ES_INDEX_TYPE")).
		Id(eventId).
		Script("ctx._source.deactivated = true").
		Upsert(map[string]interface{}{"deactivated": true}).
		Do()

	log.Info("elastic update event with id: ", res.Id)

	return err
}

func (es *ElasticSearch) ActivateEventByID(eventId string) error {

	res, err :=  es.client.Update().
		Index(viper.GetString("ES_INDEX")).
		Type(viper.GetString( "ES_INDEX_TYPE")).
		Id(eventId).
		Script("ctx._source.deactivated = false").
		Upsert(map[string]interface{}{"deactivated": false}).
		Do()

	log.Info("elastic update event with id: ", res.Id)

	return err
}

func (es *ElasticSearch) CheckIndex() {
	res, err := es.client.Get().
		Index("events").
		Type("event").
		Id("bfo5k5741br000d5dcu0").
		//Script("ctx._source.eventDetail.name = 'External Testing'").
		Do()
	if err != nil {
		fmt.Println("error: ",err)
		return
	}
	j, _ := json.Marshal(res.Source)
	fmt.Println("Response: ", string(j))
}