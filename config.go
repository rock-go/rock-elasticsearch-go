package elasticsearch

import (
	"context"
	es "github.com/olivere/elastic/v6"
	"github.com/rock-go/rock/lua"
	"reflect"
)

var ELASTIC = reflect.TypeOf((*ES)(nil)).String()

type config struct {
	name     string
	addr     []string
	user     string
	password string
	index    string
	//fields    []string
	//query     string
	//gte       string
	//lte       string
	//rangeName string
	//interval  int // 每次查询的时间间隔，如果小于等于0，则只查询一次

	buffer int
}

type ES struct {
	lua.Super
	cfg *config

	client *es.Client

	//status lua.LightUserDataStatus
	//uptime string

	ctx    context.Context
	cancel context.CancelFunc
}

type Handler interface {
	Handle([]byte) error
}
