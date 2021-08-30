package elasticsearch

import (
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/xcall"
)

func LuaInjectApi(env xcall.Env) {
	env.Set("es", lua.NewFunction(newLuaEs))
}
