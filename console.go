package elasticsearch

import "github.com/rock-go/rock/lua"

func (e *ES) Header(out lua.Printer) {
	out.Printf("type: %s", e.T)
	out.Printf("uptime: %s", e.U)
	out.Printf("version: v1.0.0")
	out.Println("")
}

func (e *ES) Show(out lua.Printer) {
	e.Header(out)

	out.Printf("name: %s", e.cfg.name)
	out.Printf("addr: %s", e.cfg.addr)
	out.Printf("user: %s", e.cfg.user)
	out.Printf("index: %s", e.cfg.index)
	out.Printf("buffer: %d", e.cfg.buffer)
	out.Println("")
}

func (e *ES) Help(out lua.Printer) {
	e.Header(out)

	out.Printf(".start() 启动")
	out.Printf(".close() 关闭")
	out.Println("")
}
