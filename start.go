package main

import (
	"./dock"
	"github.com/ssgo/s"
)

func main() {
	dock.Registers()
	dock.AsyncStart()
	s.Start()
	dock.AsyncStop()
}
