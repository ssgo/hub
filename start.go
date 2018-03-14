package main

import (
	"./dock"
	"github.com/ssgo/s"
)

// TODO 使用 ssgo/s 启动，支持 check 检查服务运行状态，支持 http 接口查看信息
func main() {
	dock.Registers()
	dock.AsyncStart()
	s.Start()
	dock.AsyncStop()
}





