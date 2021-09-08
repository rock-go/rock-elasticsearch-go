package elasticsearch

import "github.com/rock-go/rock/lua"

func (c *Chart) LPie(L *lua.LState) int {
	n := L.GetTop()
	if n != 3 {
		L.RaiseError("pie chart need 3 args, got %d", n)
		return 0
	}

	search := L.CheckLightUserData(1)
	if search == nil {
		L.RaiseError("search light userdata is nil")
		return 0
	}

	s, ok := search.Value.(*Search)
	if !ok {
		L.RaiseError("userdata must be *Body")
		return 0
	}

	dimension := L.CheckString(2)
	title := L.CheckString(3)

	c.Pie(*s, dimension, title)
	return 0
}

func (c *Chart) LBar(L *lua.LState) int {
	n := L.GetTop()
	if n != 6 {
		L.RaiseError("bar chart need 6 args, got %d", n)
		return 0
	}

	search := L.CheckLightUserData(1)
	if search == nil {
		L.RaiseError("search light userdata is nil")
		return 0
	}

	s, ok := search.Value.(*Search)
	if !ok {
		L.RaiseError("userdata must be *Body")
		return 0
	}

	typ := L.CheckInt(2)
	dimension := L.CheckString(3)
	title := L.CheckString(4)
	x := L.CheckString(5)
	y := L.CheckString(6)

	c.Bar(*s, typ, dimension, title, x, y)
	return 0
}

func (c *Chart) LLine(L *lua.LState) int {
	n := L.GetTop()
	if n != 6 {
		L.RaiseError("line chart need 6 args, got %d", n)
		return 0
	}

	search := L.CheckLightUserData(1)
	if search == nil {
		L.RaiseError("search light userdata is nil")
		return 0
	}

	s, ok := search.Value.(*Search)
	if !ok {
		L.RaiseError("userdata must be *Body")
		return 0
	}

	typ := L.CheckInt(2)
	dimension := L.CheckString(3)
	title := L.CheckString(4)
	x := L.CheckString(5)
	y := L.CheckString(6)

	c.Line(*s, typ, dimension, title, x, y)
	return 0
}

func (c *Chart) LGeo(L *lua.LState) int {
	n := L.GetTop()
	if n != 5 {
		L.RaiseError("geo chart need 5 args, got %d", n)
		return 0
	}

	search := L.CheckLightUserData(1)
	if search == nil {
		L.RaiseError("search light userdata is nil")
		return 0
	}

	s, ok := search.Value.(*Search)
	if !ok {
		L.RaiseError("userdata must be *Body")
		return 0
	}

	mapT := L.CheckString(2)
	dimension := L.CheckString(3)
	title := L.CheckString(4)
	dbPath := L.CheckString(5)

	c.Geo(*s, mapT, dimension, title, dbPath)
	return 0
}

func (c *Chart) LPage(L *lua.LState) int {
	path := L.CheckString(1)
	err := c.Page(path)
	if err != nil {
		L.RaiseError("generate visualization page error: %v", err)
	}

	return 0
}

func (c *Chart) Index(L *lua.LState, key string) lua.LValue {
	if key == "pie" {
		return lua.NewFunction(c.LPie)
	}
	if key == "bar" {
		return lua.NewFunction(c.LBar)
	}
	if key == "line" {
		return lua.NewFunction(c.LLine)
	}
	if key == "geo" {
		return lua.NewFunction(c.LGeo)
	}
	if key == "page" {
		return lua.NewFunction(c.LPage)
	}

	return lua.LNil
}
