package service

import (
	"code.drogo.gw.com/database"
	"context"
)

type Server struct {
	*database.ArangoDB
}

// NewServer creates ArangoDB clients
func NewServer() (Server, error) {

	server := Server{}
	var err error

	ctx := context.Background()

	// init ArangoDB
	server.ArangoDB, err = database.NewArangoDB(ctx)
	return server, err
}

// NewArangoDB creates only NewArangoDB client
func NewArangoDB(ctx context.Context) (*database.ArangoDB, error) {
	return database.NewArangoDB(ctx)
}