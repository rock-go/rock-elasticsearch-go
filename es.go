package elasticsearch

import (
	"context"
	"errors"
	"fmt"
	es7 "github.com/olivere/elastic/v7"
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"time"
)

func newEs(cfg *config) *ES {
	e := &ES{cfg: cfg}
	e.S = lua.INIT
	e.T = ELASTIC
	return e
}

func (e *ES) Conn() error {
	var err error
	e.client, err = es7.NewClient(
		es7.SetBasicAuth(e.cfg.user, e.cfg.password),
		es7.SetURL(e.cfg.addr...),
	)

	return err
}

func (e *ES) Start() error {
	if err := e.Conn(); err != nil {
		logger.Errorf("connect to es error: %v", err)
		return err
	}

	e.S = lua.RUNNING
	e.U = time.Now()
	e.ctx, e.cancel = context.WithCancel(context.Background())

	logger.Errorf("%s elasticsearch start successfully", e.cfg.name)
	return nil
}

func (e *ES) Close() error {
	e.S = lua.CLOSE
	e.cancel()
	logger.Errorf("%s elasticsearch close successfully", e.cfg.name)
	return nil
}

// NewSearch 新建查询
func (e *ES) NewSearch(typ string, b *Body) (*Search, error) {
	if e.S == lua.CLOSE {
		logger.Errorf("es7 module's status is close, search exit")
		return nil, errors.New("es7 status is close")
	}

	body := b.String()
	search := &Search{
		client: e.client,
		index:  e.cfg.index,
		typ:    typ,
		body:   body,
		buffer: make(chan []byte, e.cfg.buffer),
		tk:     time.NewTicker(10 * time.Second),
	}
	search.ctx, search.cancel = context.WithCancel(context.Background())
	return search, nil
}

func (e *ES) Type() string {
	return "search for elasticsearch"
}

func (e *ES) Name() string {
	return e.cfg.name
}

func (e *ES) Status() string {
	return fmt.Sprintf("elasticsearch name: %s, uptime: %s, index: %s",
		e.cfg.name, e.U.Format("2006.01.02 15:04:05"), e.cfg.index)
}

func (e *ES) State() lua.LightUserDataStatus {
	return e.S
}

func (e *ES) GetName() string {
	return e.cfg.name
}
