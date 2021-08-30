package elasticsearch

import (
	"github.com/rock-go/rock/lua"
	"strings"
)

func (e *ES) start(L *lua.LState) int {
	err := e.Start()
	if err != nil {
		L.RaiseError("es search module start error: %v", err)
	}
	return 0
}

func (e *ES) close(L *lua.LState) int {
	err := e.Close()
	if err != nil {
		L.RaiseError("es search module close error: %v", err)
	}
	return 0
}

func (e *ES) LToJson(L *lua.LState) int {
	v, _ := e.ToJson()
	L.Push(lua.LString(v))
	return 1
}

func (e *ES) LNewSearch(L *lua.LState) int {
	n := L.GetTop()
	if n != 2 {
		L.RaiseError("new_search func must have 2 args, got %d", n)
		return 0
	}

	typ := L.CheckString(1)
	ud := L.CheckUserData(2)
	body, ok := ud.Value.(*Body)
	if !ok {
		L.RaiseError("args #2 must be *Body")
		return 0
	}

	search, err := e.NewSearch(typ, body)
	if err != nil {
		L.RaiseError("new search error: %v", err)
		return 0
	}

	L.Push(&lua.LightUserData{Value: search})
	go search.Start()
	return 1
}

func (e *ES) Index(L *lua.LState, key string) lua.LValue {
	if key == "start" {
		return lua.NewFunction(e.start)
	}
	if key == "close" {
		return lua.NewFunction(e.close)
	}
	if key == "json" {
		return lua.NewFunction(e.LToJson)
	}
	if key == "search_body" {
		return lua.NewFunction(newSearchBody)
	}
	if key == "new_search" {
		return lua.NewFunction(e.LNewSearch)
	}
	return lua.LNil
}

func newLuaEs(L *lua.LState) int {
	opt := L.CheckTable(1)
	cfg := config{
		name:     opt.CheckString("name", "es"),
		addr:     stringToSlice(opt.CheckString("addr", "http://127.0.0.1:9200/")),
		user:     opt.CheckString("user", ""),
		password: opt.CheckString("password", ""),
		index:    opt.CheckString("index", ""),
		buffer:   opt.CheckInt("buffer", 4096),
	}

	proc := L.NewProc(cfg.name, ELASTIC)
	if proc.IsNil() {
		proc.Set(newEs(&cfg))
	} else {
		proc.Value.(*ES).cfg = &cfg
	}

	L.Push(proc)
	return 1
}

// 逗号分隔string，返回slice
func stringToSlice(s string) []string {
	var data []string
	s = strings.Trim(s, " ")
	data = strings.Split(s, ",")
	return data
}
