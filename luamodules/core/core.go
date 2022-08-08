package core

import (
	libs "github.com/vadv/gopher-lua-libs"
	lua "github.com/yuin/gopher-lua"
)

type Core struct{}

func New() *Core {
	return &Core{}
}

func (_ Core) Name() string {
	return "core"
}

func Preload(L *lua.LState) {
	libs.Preload(L)
}

func (_ Core) Loader() lua.LGFunction {
	return func(L *lua.LState) int {
		return 0
	}
}
