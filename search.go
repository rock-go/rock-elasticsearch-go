package elasticsearch

import (
	"context"
	es "github.com/olivere/elastic/v6"
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"io"
	"io/ioutil"
	"os"
	"time"
)

// Search 搜索结果，缓存在buffer内
type Search struct {
	lua.NoReflect
	lua.Super

	client *es.Client
	index  string
	body   interface{}
	typ    string
	buffer chan []byte

	tk     *time.Ticker // 当buffer已满时，定期轮询，防止阻塞
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *Search) Start() error {
	scroll := s.client.Scroll().
		Index(s.index).Body(s.body).
		//Query(s.query).
		Size(10000).
		Pretty(true)

	for {
		select {
		case <-s.ctx.Done():
			// 查询未完成，关闭此次查询
			close(s.buffer)
			logger.Errorf("search exit")
			return nil
		default:
			searchResult, err := scroll.Do(s.ctx)
			if err == io.EOF {
				// 查询结束
				close(s.buffer)
				logger.Infof("search complete")
				_ = s.Close()
				return nil
			}

			if err != nil {
				close(s.buffer)
				logger.Errorf("search body [%v] error: %v", s.body, err)
				return nil
			}

			if s.typ == "agg" {
				if searchResult.Aggregations == nil {
					close(s.buffer)
					return nil
				}
			}

			s.handleRes(searchResult)
		}
	}
}

// 处理每次查询的结果
func (s *Search) handleRes(res *es.SearchResult) {
	switch s.typ {
	case "bool":
		for _, hit := range res.Hits.Hits {
			sourceData, err := hit.Source.MarshalJSON()
			if err != nil {
				logger.Errorf("marshall source to json error: %v", err)
				continue
			}
			// 此处可能会存在阻塞
			s.buffer <- sourceData
		}
	case "agg":
		for _, hit := range res.Aggregations {
			aggData, err := hit.MarshalJSON()
			if err != nil {
				logger.Errorf("marshal source to json error: %v", err)
				return
			}
			//todo
			aggData, _ = ioutil.ReadFile("resource/waf_access/result_for_bar_or_line_multi.json")
			s.buffer <- aggData
		}
	default:
		return
	}
}

// Close 关闭本次搜索，将关闭结果缓存通道
func (s *Search) Close() error {
	s.cancel()
	s.handleBuffer()
	s.tk.Stop()
	s.S = lua.CLOSE
	logger.Errorf("search closed")
	return nil
}

// 关闭时，解决阻塞情况
func (s *Search) handleBuffer() {
	i := 0
	for {
		if len(s.buffer) == 0 {
			return
		}
		time.Sleep(1 * time.Second)
		i++
		if i >= 10 {
			// 超时，清空缓存数据，防止阻塞
			for {
				_ = <-s.buffer
				if len(s.buffer) == 0 {
					logger.Errorf("search buffer consume timeout, exit")
					return
				}
			}
		}
		logger.Errorf("search buffer wait consuming for %ds", i)
	}
}

func (s *Search) File(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	for {
		select {
		case data, ok := <-s.buffer:
			if !ok {
				return nil
			}
			f.Write(data)
		}
	}
}

// GetBuffer Input interface
func (s *Search) GetBuffer() *chan []byte {
	return &s.buffer
}

func (s Search) GetName() string {
	return "es_search"
}
