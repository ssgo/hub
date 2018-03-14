package tests

import (
	"testing"
	"../dock"
	"time"
	"github.com/ssgo/redis"
	"github.com/ssgo/s"
	"strings"
	"os"
)

var dcCache *redis.Redis

func T1estWithoutRedis(tt *testing.T) {
	t := s.T(tt)

	os.Setenv("dock_nodes_node1", "4,8")
	os.Setenv("dock_apps_app1", "1,1,1,1,...")
	os.Setenv("dock_binds_app1", "node1")

	dock.SetShell(testShell)
	dock.SetSleepUnit(time.Millisecond * 10)
	dock.AsyncStart()

	// 添加节点
	time.Sleep(time.Millisecond * 100)
	s := dock.GetStats()
	t.Test(len(s.Nodes) == 1 && len(s.Apps) == 1 &&
		s.Nodes["node1"] != nil && s.Nodes["node1"].TotalCpu == 4 &&
		s.Apps["app1"] != nil && s.Apps["app1"].Runs[0].Node == "node1",
		"TestWithoutRedis")

	dock.AsyncStop()
}

func TestStart(tt *testing.T) {
	os.Setenv("dock_nodes_node1", "4,8")
	os.Setenv("dock_apps_app1", "1,1,1,1,...")
	os.Setenv("dock_binds_app1", "node2")

	dcCache = redis.GetRedis("dock:14")
	dcCache.DEL("_nodes", "_apps", "_binds", "_appIndexes")

	dock.SetShell(testShell)
	dock.SetSleepUnit(time.Millisecond * 10)
	dock.AsyncStart()
}

func TestBase(tt *testing.T) {
	t := s.T(tt)

	// 添加节点
	dcCache.HSET("_nodes", "node2", "8,16")
	time.Sleep(time.Millisecond * 100)
	s := dock.GetStats()
	t.Test(len(s.Nodes) == 2 &&
		s.Nodes["node1"] != nil && s.Nodes["node1"].TotalCpu == 4 &&
		s.Nodes["node2"] != nil && s.Nodes["node2"].TotalMemory == 16,
		"Add Nodes")

	// 添加应用
	dcCache.HSET("_apps", "app2", "2,4,2,4,...")
	time.Sleep(time.Millisecond * 100)
	s = dock.GetStats()
	t.Test(len(s.Apps) == 2 && s.Apps["app2"] != nil && s.Apps["app2"].Min == 2 &&
		len(s.Apps["app2"].Runs) == 2 && strings.Index(s.Apps["app2"].Runs[0].Id+s.Apps["app2"].Runs[1].Id, "<01>") != -1,
		"Add App 1&2")
}

func TestUpdate(tt *testing.T) {
	t := s.T(tt)

	// 添加节点
	dcCache.HSET("_nodes", "node2", "8,24")
	time.Sleep(time.Millisecond * 100)
	s := dock.GetStats()
	t.Test(len(s.Nodes) == 2 &&
		s.Nodes["node1"] != nil && s.Nodes["node1"].TotalCpu == 4 &&
		s.Nodes["node2"] != nil && s.Nodes["node2"].TotalMemory == 24,
		"Update Nodes")
	app2Node1 := s.Apps["app2"].Runs[0].Node
	app2Node2 := s.Apps["app2"].Runs[1].Node

	// 添加应用
	dcCache.HSET("_apps", "app2", "2,8,2,4,...")
	time.Sleep(time.Millisecond * 100)
	s = dock.GetStats()
	t.Test(len(s.Apps) == 2 && s.Apps["app2"] != nil && s.Apps["app2"].Min == 2 &&
		len(s.Apps["app2"].Runs) == 2 && s.Apps["app2"].Memory == 8 &&
		s.Apps["app2"].Runs[0].Node != app2Node1 && s.Apps["app2"].Runs[1].Node != app2Node2,
		"Update App 1&2")
}

func TestBind(tt *testing.T) {
	t := s.T(tt)

	// 添加带 -v 的应用
	dcCache.HMSET("_apps", "app3", "1,1,1,1,... -v ...", "app4", "2,4,2,4,... --volume ...")
	time.Sleep(time.Millisecond * 100)
	s := dock.GetStats()
	t.Test(len(s.Apps) == 4 && s.Apps["app4"] != nil && s.Apps["app4"].Min == 2 &&
		len(s.Apps["app3"].Runs) == 1 && len(s.Apps["app4"].Runs) == 2,
		"Add App 3&4 With -v")
	app3Node := s.Apps["app3"].Runs[0].Node
	app4Node1 := s.Apps["app4"].Runs[0].Node
	app4Node2 := s.Apps["app4"].Runs[1].Node

	// 删除 app 2 & 3 & 4
	dcCache.HDEL("_apps", "app2", "app3", "app4")
	time.Sleep(time.Millisecond * 100)
	s = dock.GetStats()
	t.Test(len(s.Apps) == 1 && s.Apps["app4"] == nil && s.Apps["app1"] != nil,
		"Remove App 2&3&4")

	// 添加带 -v 的应用
	dcCache.HMSET("_apps", "app3", "1,1,1,1,... -v ...", "app4", "2,4,2,4,... --volume ...")
	time.Sleep(time.Millisecond * 100)
	s = dock.GetStats()
	t.Test(len(s.Apps) == 3 && s.Apps["app4"] != nil &&
		len(s.Apps["app3"].Runs) == 1 && app3Node == s.Apps["app3"].Runs[0].Node &&
		len(s.Apps["app4"].Runs) == 2 &&
		(app4Node1 == s.Apps["app4"].Runs[0].Node || app4Node1 == s.Apps["app4"].Runs[1].Node ) &&
		(app4Node2 == s.Apps["app4"].Runs[0].Node || app4Node2 == s.Apps["app4"].Runs[1].Node ),
		"Add App 3&4 For Bind")

	// 删除 app 3 & 4
	dcCache.HDEL("_apps", "app2", "app3", "app4")
	time.Sleep(time.Millisecond * 100)
	s = dock.GetStats()
	t.Test(len(s.Apps) == 1 && s.Apps["app4"] == nil && s.Apps["app1"] != nil,
		"Remove App 3&4")

	// 添加 app2 on node1
	dcCache.HSET("_binds", "app2", "node1,node1,node1")
	dcCache.HMSET("_apps", "app2", "2,4,3,3,...")
	time.Sleep(time.Millisecond * 100)
	s = dock.GetStats()
	t.Test(len(s.Apps) == 2 && len(s.Apps["app2"].Runs) == 3 &&
		s.Apps["app2"].Runs[0].Node == "node1" &&
		s.Apps["app2"].Runs[1].Node == "node1" &&
		s.Apps["app2"].Runs[2].Node == "node1",
		"Add App 2 With Binds", len(s.Apps), len(s.Apps["app2"].Runs))

	// 添加带 -v 的应用
	dcCache.HMSET("_apps", "app3", "1,1,1,1,... -v ...", "app4", "2,4,2,4,... --volume ...")
	time.Sleep(time.Millisecond * 100)
	s = dock.GetStats()
	t.Test(len(s.Apps) == 4 && s.Apps["app4"] != nil &&
		len(s.Apps["app3"].Runs) == 1 && app3Node == s.Apps["app3"].Runs[0].Node &&
		len(s.Apps["app4"].Runs) == 2 &&
		(app4Node1 == s.Apps["app4"].Runs[0].Node || app4Node1 == s.Apps["app4"].Runs[1].Node ) &&
		(app4Node2 == s.Apps["app4"].Runs[0].Node || app4Node2 == s.Apps["app4"].Runs[1].Node ),
		"Add App 3&4 For Bind Again")
}

func TestPostfix(tt *testing.T) {
	t := s.T(tt)

	// 测试后缀
	dcCache.HSET("_apps", "app5#a", "1,1,2,2,... <xxx>")
	time.Sleep(time.Millisecond * 100)
	s := dock.GetStats()
	t.Test(len(s.Apps) == 5 && len(s.Apps["app5#a"].Runs) == 2,
		"Add App 5#a")

	dock.SetSleepUnit(time.Second)
	time.Sleep(time.Millisecond * 100)

}

func TestPubSub(tt *testing.T) {
	t := s.T(tt)

	// 添加 app5，测试 pub sub
	dcCache.HSET("_apps", "app6", "1,4,2,4,...")
	dcCache.Do("PUBLISH", "_refresh", "1")
	time.Sleep(time.Second)
	s := dock.GetStats()
	t.Test(len(s.Apps) == 6 && len(s.Apps["app6"].Runs) == 2,
		"Add App 6")
}

func TestEnd(tt *testing.T) {
	dock.AsyncStop()
	dcCache.DEL("_nodes", "_apps", "_binds", "_appIndexes")
}
