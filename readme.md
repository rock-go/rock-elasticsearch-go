# 说明

​		rock-elasticsearch-go模块基于rock-go框架开发，用于连接elasticsearch，进行搜索和索引等操作，并能够将聚合查询结果绘制成html格式的直方图，折线图，或饼图。

# 使用

搜索：模块简单地根据elasticsearch的搜索请求body格式，取出数据（[]byte类型，满足json格式），缓存于此搜索对应的channel中，其他模块通过该通道获取数据并进行处理。

结果处理：一个搜索只设置了一个channel，故一个搜索的结果只能被一个方式消费，比如存入文件，或画图。

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

local body = ssl_vpn_es.search_body {
    source = "resource/access_log/host_count_percentage.txt",
    date_field = { t = "string", v = "@timestamp" },
    gte = { t = "time", v = "-1h" },
    lte = { t = "time", v = "now" },
}

-- 下列步骤可定时运行
local search = ssl_vpn_es.new_search("agg", body)
--local now = "20210828"
--res = "resource/res-" .. now .. ".json"
-- 结果存入文件
--search.file(res)

local search1 = ssl_vpn_es.new_search("agg", body)
--local search2 = ssl_vpn_es.new_search("agg", body)

local vis = ssl_vpn_es.new_vis()
--vis.pie(search, "count", "各域名请求数占比")
vis.line(search1, 3, "count", "各域名请求数24h趋势图", "时间", "请求数")
vis.bar(search, 3, "count", "各域名请求数24h趋势图", "时间", "请求数")
vis.page("resource/waf_access/20210901-股吧.html")
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

返回结果示例

```text
{
  "buckets": [
    {
      "3": {
        "doc_count_error_upper_bound": 0,
        "sum_other_doc_count": 1,
        "buckets": [
          {
            "1": {
              "value": 136281050
            },
            "key": "x1.domain.com",
            "doc_count": 31983
          },
          {
            "1": {
              "value": 31778785
            },
            "key": "x3.domain.com",
            "doc_count": 408
          },
          {
            "1": {
              "value": 20060620
            },
            "key": "x2.domain.com",
            "doc_count": 4281
          },
          {
            "1": {
              "value": 133724
            },
            "key": "x2topic.domain.com",
            "doc_count": 29
          },
          {
            "1": {
              "value": 6674
            },
            "key": "gb.domain.com",
            "doc_count": 47
          }
        ]
      },
      "key_as_string": "2021-08-31T16:47:00.000+08:00",
      "key": 1630399620000,
      "doc_count": 36749
    },
    {
      "3": {
        "doc_count_error_upper_bound": 0,
        "sum_other_doc_count": 11,
        "buckets": [
          {
            "1": {
              "value": 1071190890
            },
            "key": "x1.domain.com",
            "doc_count": 233212
          },
          {
            "1": {
              "value": 270854743
            },
            "key": "x3.domain.com",
            "doc_count": 3489
          },
          {
            "1": {
              "value": 150602518
            },
            "key": "x2.domain.com",
            "doc_count": 32609
          },
          {
            "1": {
              "value": 901956
            },
            "key": "x2topic.domain.com",
            "doc_count": 157
          },
          {
            "1": {
              "value": 66030
            },
            "key": "gb.domain.com",
            "doc_count": 465
          }
        ]
      },
      "key_as_string": "2021-08-31T16:48:00.000+08:00",
      "key": 1630399680000,
      "doc_count": 269943
    },
    {
      "3": {
        "doc_count_error_upper_bound": 0,
        "sum_other_doc_count": 18,
        "buckets": [
          {
            "1": {
              "value": 1045137008
            },
            "key": "x1.domain.com",
            "doc_count": 237108
          },
          {
            "1": {
              "value": 248166531
            },
            "key": "x3.domain.com",
            "doc_count": 3240
          },
          {
            "1": {
              "value": 158578408
            },
            "key": "x2.domain.com",
            "doc_count": 28323
          },
          {
            "1": {
              "value": 1270670
            },
            "key": "x2topic.domain.com",
            "doc_count": 239
          },
          {
            "1": {
              "value": 69154
            },
            "key": "gb.domain.com",
            "doc_count": 487
          }
        ]
      },
      "key_as_string": "2021-08-31T16:49:00.000+08:00",
      "key": 1630399740000,
      "doc_count": 269415
    }
  ]
}
```

#### 参数说明

##### 连接ES

- name: 模块名称，用于日志标识和其他服务调用
- addr: es地址，格式须为 http://ip:port/ 。如果有多个地址，用逗号分隔
- user: 连接需要认证的elasticsearch时，所需的用户名
- password: 连接需要认证的elasticsearch时，所需的密码
- index: 需要搜索数据的索引名称
- buffer: 缓存搜索结果的通道大小。其他模块调用时，从该通道获取结果。默认大小 4096 

##### 创建search body

- source：固定字段，代表搜索的请求body字符串，或存储的文件路径，请参考上述示例。
- 其它：其它字符串的名称用来替换source中的占位符，占位符为%xxx%。其值为lua的表，包含两个固定字段，t：类型，为"time"时会转化为对应的值，支持格式如： now, -1m, -2s, -24h, 2021.08.28 15:15:00，通常用于时间范围字段，如gte和lte；其它类型则直接替换；v：替换后的值。

##### 搜索

- 函数名：.new_search("agg", body)
- 调用对象：rock.es
- 参数1：agg，表示聚合查询，通常用于统计数据；bool，布尔查询，通常用于直接获取原始结果
- 参数2：search body，对应的userdata
- 返回结果：返回search对应的userdata。结果缓存在golang的channel中

##### 结果处理-存储到文件

- 函数名：.file(path)
- 调用对象：search
- 参数1：结果存储路径
- 返回结果：数据以json格式存储至指定路径

##### 结果处理-绘图

.new_vis()  es对象调用，新建一个图表处理对象vis

###### 饼图

- 函数名：.pie(search, "count", "title")
- 调用对象：图表处理对象vis
- 参数1：搜索对象
- 参数2：取值类型，count 统计的值为文档个数，value 统计的值为文档某字段的值（求和，平均等） 
- 参数3：图表标题
- 返回结果：图表对象，存储到golang channel中

###### 直方图

- 函数名：.bar(search, 2, "count", "tile", "x name", "y name")
- 调用对象：图表处理对象vis
- 参数1：搜索对象
- 参数2：图表类型。2 基础类型，一个x轴坐标点对应一个类型的数据；3 聚合类型，一个x轴坐标点对应多个类型的数据
- 参数3：取值类型，count 统计的值为文档个数，value 统计的值为文档某字段的值（求和，平均等） 
- 参数4：图表标题
- 参数5：x轴显示名称
- 参数6：y轴显示名称
- 返回结果：图表对象，存储到golang channel中

###### 折线图

- 函数名：.line(search, 2, "count", "tile", "x name", "y name")
- 调用对象：图表处理对象vis
- 参数1：搜索对象
- 参数2：图表类型。2 基础类型，一个x轴坐标点对应一个类型的数据；3 聚合类型，一个x轴坐标点对应多个类型的数据
- 参数3：取值类型，count 统计的值为文档个数，value 统计的值为文档某字段的值（求和，平均等） 
- 参数4：图表标题
- 参数5：x轴显示名称
- 参数6：y轴显示名称
- 返回结果：图表对象，存储到golang channel中

### 其它模块调用

rock-elasticsearch-go模块实现了Input接口，该接口返回了模块缓存搜索结果的通道，其它模块调用时，从该通道读取数据并处理搜索结果。

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