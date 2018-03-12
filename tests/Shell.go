package tests

import (
	"fmt"
	"../dock"
	"strings"
)

var sequences = map[string]int{}
var runs = map[string]map[string]string{}

func testShell(nodeName string, args ...string) string {
	if runs[nodeName] == nil {
		runs[nodeName] = map[string]string{}
	}

	if args[0] == "run" {
		if dock.GetStats().Nodes[nodeName] == nil {
			return ""
		}
		sequences[nodeName]++
		id := fmt.Sprintf("%s<%.2d>", nodeName, sequences[nodeName])
		runs[nodeName][id] = args[len(args)-1]
		return id
	}
	if args[0] == "ps" {
		lines := make([]string, 0)
		for id, image := range runs[nodeName] {
			lines = append(lines, id+", "+image+", Up 1 minutes")
		}
		return strings.Join(lines, "\n")
	}
	if args[0] == "stop" {
		return args[1]
	}
	if args[0] == "rm" {
		delete(runs[nodeName], args[1])
		return args[1]
	}
	return ""
}
