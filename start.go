package main

import (
	"github.com/ssgo/hub/dock"
	"github.com/ssgo/hub/gate"
	"github.com/ssgo/s"
)

func main() {
	dock.Registers()
	dock.AsyncStart()
	gate.Registers()
	s.Start()
	dock.AsyncStop()
}
