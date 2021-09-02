package elasticsearch

import "github.com/rock-go/rock/lua"

func (s *Search) LStart(L *lua.LState) int {
	go s.Start()
	return 0
}

func (s *Search) LClose(L *lua.LState) int {
	s.Close()
	return 0
}

func (s *Search) LFile(L *lua.LState) int {
	path := L.CheckString(1)
	go s.File(path)
	return 0
}

func (s *Search) Index(L *lua.LState, key string) lua.LValue {
	if key == "start" {
		return L.NewFunction(s.LStart)
	}
	if key == "close" {
		return L.NewFunction(s.LClose)
	}
	if key == "file" {
		return L.NewFunction(s.LFile)
	}

	return lua.LNil
}
