package elasticsearch

import (
	"context"
	"fmt"
	"testing"
	"time"
)

var doc = `{
  "version": true,
  "size": 500,
  "sort": [
    {
      "@timestamp": {
        "order": "desc",
        "unmapped_type": "boolean"
      }
    }
  ],
  "_source": {
    "excludes": []
  },
  "aggs": {
    "2": {
      "date_histogram": {
        "field": "@timestamp",
        "interval": "30s",
        "time_zone": "Asia/Shanghai",
        "min_doc_count": 1
      }
    }
  },
  "stored_fields": [
    "*"
  ],
  "script_fields": {},
  "docvalue_fields": [
    "@timestamp",
    "time_created",
    "timestamp"
  ],
  "query": {
    "bool": {
      "must": [
        {
          "match_all": {}
        },
        {
          "range": {
            "@timestamp": {
              "gte": 1629995167105,
              "lte": 1629996067105,
              "format": "epoch_millis"
            }
          }
        }
      ],
      "filter": [],
      "should": [],
      "must_not": []
    }
  },
  "highlight": {
    "pre_tags": [
      "@kibana-highlighted-field@"
    ],
    "post_tags": [
      "@/kibana-highlighted-field@"
    ],
    "fields": {
      "*": {}
    },
    "fragment_size": 2147483647
  }
}`

var agg = `{
  "size": 0,
  "_source": {
    "excludes": []
  },
  "aggs": {
    "2": {
      "date_histogram": {
        "field": "%date_field%",
        "interval": "10m",
        "time_zone": "Asia/Shanghai",
        "min_doc_count": 1
      }
    }
  },
  "stored_fields": [
    "*"
  ],
  "script_fields": {},
  "docvalue_fields": [
    "@timestamp",
    "time_created",
    "timestamp"
  ],
  "query": {
    "bool": {
      "must": [
        {
          "match_all": {}
        },
        {
          "range": {
            "@timestamp": {
              "gte": %gte%,
              "lte": %lte%,
              "format": "epoch_millis"
            }
          }
        }
      ],
      "filter": [],
      "should": [],
      "must_not": []
    }
  }
}`

func TestSearch_Start(t *testing.T) {
	es := ES{cfg: &config{
		name:     "test",
		addr:     []string{"http://172.16.88.58:9200/"},
		user:     "",
		password: "",
		index:    "ssl*",
		buffer:   4096,
	}}

	es.Start()

	//query := elastic.RawStringQuery(doc)
	search := &Search{
		client: es.client,
		index:  es.cfg.index,
		body:   agg,
		typ:    "agg",
		buffer: make(chan []byte, es.cfg.buffer),
		tk:     time.NewTicker(10 * time.Second),
	}
	search.ctx, search.cancel = context.WithCancel(context.Background())

	go search.Start()

	//f, _ := os.OpenFile("res.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	//defer f.Close()
	//for {
	//	select {
	//	case data := <-search.buffer:
	//		f.Write(data)
	//	default:
	//
	//	}
	//}
	search.File("test.json")
}

func TestChan(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 1

	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(1 * time.Second)
			fmt.Println(i)
		}
		close(ch)
	}()

	for {
		select {
		case d, ok := <-ch:
			if ok {
				fmt.Println(d)
			} else {
				fmt.Println(22)
				return
			}
		default:
		}
	}
}

func TestBody_String(t *testing.T) {
	body := &Body{
		source: agg,
		fields: map[string]Filed{
			"date_field": {
				t: "string",
				v: "@timestamp",
			},
			"gte": {
				t: "time",
				v: "-24h",
			},
			"lte": {
				t: "time",
				v: "now",
			},
		},
	}

	s := body.String()
	fmt.Println(s)
}
