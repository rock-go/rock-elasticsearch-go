# 说明

rock-elasticsearch-go模块基于rock-go框架开发，用于连接elasticsearch，进行搜索和索引等操作。

# 使用

搜索：模块从elasticsearch取出数据（[]byte类型，满足json格式），缓存于通道中，其他模块通过该通道获取数据并进行处理<br>
索引：导入数据，TODO

## 导入

```go
import elasticsearch "github.com/rock-go/rock-elasticsearch-go"
```

## 组件注册

```go
rock.Inject(xcall.Rock, elasticsearch.LuaInjectApi)
```

## 搜索模块

### lua 脚本配置

```lua
-- 配置连接es服务器
ssl_vpn_es = rock.es {
    name = "config_es_search",
    addr = "http://172.16.88.58:9200/",
    user = "",
    password = "",
    index = "ssl*",
    buffer = 4096
}
proc.start(ssl_vpn_es)

-- 通过search body来搜索
local body = ssl_vpn_es.search_body {
    source = "resource/count_time_body.txt",
    date_field = { t = "string", v = "@timestamp" },
    gte = { t = "time", v = "-24h" },
    lte = { t = "time", v = "now" },
}
-- 新建搜索，参数1：搜索类型，bool和agg，对应es的bool查询和agg聚合查询。参数2：请求body，如上。
local search = ssl_vpn_es.new_search("agg", body)
-- 结果存储至文件。search实现了input的接口，调用方法参考后续说明。
search.file("resource/res.json")
```

resource/count_time_body.txt中的内容，其中%xxx%为占位符，搜索的时候，会根据search_body的配置，替换掉其中的值。如上述配置中，源文本的%date_field%会被替换为@timestamp，%gte%和%lte%会被替换为对应的时间，从而生成一个可供ES查询的正确请求body。正确的请求body可参考elasticsearch官方文档，也可通过kibana搜索，查看http请求的包

```text
{
  "size": 0,
  "_source": {
    "excludes": []
  },
  "aggs": {
    "2": {
      "date_histogram": {
        "field": "%date_field%",
        "interval": "1m",
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
}
```

生成的body

```text
{
  "size": 0,
  "_source": {
    "excludes": []
  },
  "aggs": {
    "2": {
      "date_histogram": {
        "field": "@timestamp",
        "interval": "1m",
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
              "gte": 1630046447000,
              "lte": 1630132847000,
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
}
```

#### 参数说明

连接

- name: 模块名称，用于日志标识和其他服务调用

- addr: es地址，格式须为 http://ip:port/。如果有多个地址，用逗号分隔

- user: 连接需要认证的elasticsearch时，所需的用户名

- password: 连接需要认证的elasticsearch时，所需的密码

- index: 需要查询的索引名称

- buffer: 缓存搜索结果的通道大小。其他模块调用时，从该通道获取结果。默认大小 4096 <br>

  搜索
  <br>创建search body
- source：固定字段，代表搜索的请求body字符串，或存储的文件路径。
- 其它：其它字符串的名称用来替换source中的占位符，占位符为%xxx%。其值为lua的表，包含两个固定字段，t：类型，为"time"时会转化为对应的值，支持格式如： now,-1m,-2s,-24h,2021.08.28 15:15:
  00，通常用于时间范围字段，如gte和lte；其它类型则直接替换；v：值。

### 其它模块调用

rock-elasticsearch-go模块实现了Input接口，该接口返回了模块缓存搜索结果的通道，其它模块调用时，从该通道读取数据。

```go
// example

package main

import (
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"time"
)

type config struct {
	name  string
	input Input
}

// Temp 模块调用rock-elasticsearch-go模块对象，读取其结果
type Temp struct {
	lua.Super
	c      config
	buffer chan []byte
	status lua.LightUserDataStatus
}

// Input 接口
type Input interface {
	GetBuffer() *chan []byte // 返回rock-elasticsearch-go的结果缓存通道地址
	GetName() string         // 返回模块名称
}

// 从通道读取数据
func (t *Temp) Read() []byte {
	for {
		select {
		case data := <-t.buffer:
			logger.Errorf("%s", data)
			time.Sleep(1 * time.Second)
		}
	}
}

// 与lua虚拟机交互，获取到es结果缓存通道
func createTempUserData(L *lua.LState) int {
	opt := L.CheckTable(1)
	cfg := config{
		name:  opt.CheckString("name", "temp_test"),
		input: checkInput(opt.CheckLightUserData(L, "input")),
	}

	temp := &Temp{c: cfg}
	temp.buffer = *temp.c.input.GetBuffer()

	proc := L.NewProc(temp.c.name, TMP)
	proc.Value = temp
	L.Push(proc)
	return 1
}

func checkInput(data *lua.LightUserData) Input {
	if input, ok := data.Value.(interface{}).(Input); ok {
		return input
	}

	return nil
}
```

```lua
--lua 配置
local tmp = rock.temp {
    name = "temp_test",
    input = service.elasticsearch.config_es_search
}
-- 从通道读取数据
tmp.read()
```

## 索引模块

### todo