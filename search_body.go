package elasticsearch

import (
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"github.com/spf13/cast"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type Body struct {
	source string           // body 字符串或文本路径
	fields map[string]Filed // body中，要替换的字段的类型和值，一般只有时间
}

type Filed struct {
	t string // 字段类型
	v string // 值
}

func newSearchBody(L *lua.LState) int {
	tb := L.CheckTable(1)
	var body = &Body{fields: make(map[string]Filed)}

	tb.ForEach(func(k lua.LValue, v lua.LValue) {
		if k.String() == "source" {
			body.source = v.String()
			return
		}
		field := Filed{}
		field.t = v.(*lua.LTable).CheckString("t", "")
		field.v = v.(*lua.LTable).CheckString("v", "")
		body.fields[k.String()] = field
	})

	L.Push(&lua.LUserData{Value: body})
	return 1
}

// String 格式化body，返回字符串
// 对于一些定时类任务，需要更新时间。此时，将原始数据需要替换的值用%field%占位，本函数用配置的值去替换
func (b *Body) String() string {
	var body string
	stat, err := os.Stat(b.source)
	if err != nil {
		body = b.source
	} else if stat.IsDir() {
		logger.Errorf("body source parse error: got directory, but need file")
		return body
	}

	bodyBytes, err := ioutil.ReadFile(b.source)
	if err != nil {
		logger.Errorf("body parse error: %v", err)
		return ""
	}
	body = string(bodyBytes)
	for k, v := range b.fields {
		k = "%" + k + "%"
		// 替换模版请求txt中的%field%
		body = replace(body, k, v)
	}

	return body
}

func replace(b string, k string, f Filed) string {
	switch f.t {
	case "time":
		b = strings.Replace(b, k, cast.ToString(formatTime(f.v)), 1)
	default:
		b = strings.Replace(b, k, f.v, 1)
	}

	return b
}

// 转化gte和lte为毫秒的时间戳
func formatTime(t string) int64 {
	// now
	if t == "now" {
		return time.Now().Unix() * 1000
	}

	// e.g. -10m
	unit := []string{"h", "m", "s"}
	for _, u := range unit {
		if strings.Contains(t, u) {
			now := time.Now()
			delta, err := time.ParseDuration(t)
			if err != nil {
				logger.Errorf("parse time duration [%s] error: %v", t, err)
				return now.Unix() * 1000
			}
			ti := now.Add(delta).Unix() * 1000
			return ti
		}
	}

	// 2006.01.02 15:04:05
	ti, err := time.Parse("2006.01.02 15:04:05", t)
	if err != nil {
		logger.Errorf("parse time format [%s] error: %v", t, err)
		return time.Now().Unix() * 1000
	}

	return ti.Unix() * 1000
}
