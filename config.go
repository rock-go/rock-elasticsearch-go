package elasticsearch

import (
	"context"
	es "github.com/olivere/elastic/v7"
	"github.com/rock-go/rock/lua"
	"reflect"
)

var ELASTIC = reflect.TypeOf((*ES)(nil)).String()

type config struct {
	version  int
	name     string
	addr     []string
	user     string
	password string
	index    string

	buffer int
}

type ES struct {
	lua.Super
	cfg *config

	client *es.Client

	ctx    context.Context
	cancel context.CancelFunc
}

type Handler interface {
	Handle([]byte) error
}
