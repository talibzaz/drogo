package elastic

import (
	"testing"
	"github.com/spf13/viper"
	"context"
	"fmt"
)

func TestElasticSearch_DeactivateEventByID(t *testing.T) {
	viper.Set("ES_HOST", "139.59.85.55")
	viper.Set("ES_PORT", "9200")

	viper.Set("ES_INDEX", "events")
	viper.Set("ES_INDEX_TYPE", "event")

	es, _ := NewElasticSearch(context.Background())

	err := es.DeactivateEventByID("bf4o5fp4r7jg00bhof70")

	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	fmt.Println("No error")

	return
}

func TestElasticSearch_CheckIndex(t *testing.T) {
	viper.Set("ES_HOST", "139.59.85.55")
	viper.Set("ES_PORT", "9200")

	viper.Set("ES_INDEX", "events")
	viper.Set("ES_INDEX_TYPE", "event")

	es, _ := NewElasticSearch(context.Background())

	es.CheckIndex()

}