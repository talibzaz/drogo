package client

import (
	"testing"
	"encoding/json"
	"os"
	log "github.com/sirupsen/logrus"
)

func TestFetchEventById(t *testing.T) {
	eventId := "bf84nqluvaqg00ddt910"
	event, err := FetchEventById(eventId)
	if err != nil {
		log.Errorf("eror: %v", err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	enc.Encode(event)
}

func TestUpdateFeaturedEvent(t *testing.T) {
	eventId := "bde1rk0dlj0000fct0v0"
	response, err := UpdateFeaturedEvent(eventId, false)
	if err != nil {
		log.Errorf("eror: %v", err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	enc.Encode(response)
}