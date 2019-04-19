package main

import (
	"github.com/ssgo/dock/dock"
	"github.com/ssgo/s"
)

func main() {
	dock.Registers()
	dock.AsyncStart()
	s.Start()
	dock.AsyncStop()
}
