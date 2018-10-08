package tests

import (
	"testing"
	"../dock"
	"time"
	"github.com/ssgo/s"
	"os"
	//"strings"
	"encoding/json"
	"crypto/sha1"
	"encoding/hex"
	"strings"
)

var as *s.AsyncServer
var nodes map[string]*dock.NodeInfo
var nodeStatus map[string]*dock.NodeStatus
var ctx dock.ContextInfo
var ctxRuns map[string][]dock.AppStatus

func TestStart(tt *testing.T) {
	os.Setenv("dock_dataPath", "/tmp/dock")
	os.Setenv("service_httpVersion", "2")
	os.RemoveAll("/tmp/dock")
	os.MkdirAll("/tmp/dock", 0700)
	fp, _ := os.OpenFile("/tmp/dock/global", os.O_CREATE|os.O_WRONLY, 0600)
	fp.Write([]byte("{\"Nodes\":{\"node1\":{\"cpu\":4,\"memory\":8}}}"))
	fp.Close()
	fp, _ = os.OpenFile("/tmp/dock/c1", os.O_CREATE|os.O_WRONLY, 0600)
	fp.Write([]byte("{\"name\":\"c1\", \"desc\":\"test ctx\", \"apps\": {\"app1\":{\"cpu\":1,\"memory\":1,\"min\":1,\"max\":1,\"active\":true,\"args\":\"... ${dc} ...\"}}, \"vars\":{\"dc\":\"-e 'discover_host=127.0.0.1' -e 'discover_port=6000' -e 'discover_password=hjfdasy7fdusihfyuasfs'\"}, \"binds\":{\"app1\":\"node1\"}}"))
	fp.Close()

	dock.SetShell(TestShell)
	dock.SetSleepUnit(time.Millisecond * 5)
	dock.Registers()
	dock.AsyncStart()
	as = s.AsyncStart()

	sha1Maker := sha1.New()
	sha1Maker.Write([]byte("SSGO-"))
	sha1Maker.Write([]byte("91dock"))
	sha1Maker.Write([]byte("-Dock"))
	as.SetGlobalHeader("Access-Token", hex.EncodeToString(sha1Maker.Sum([]byte{})))
}

func getStatus(ctxName string){
	nodes = map[string]*dock.NodeInfo{}
	nodeStatus = map[string]*dock.NodeStatus{}
	ctx = dock.ContextInfo{}
	ctxRuns = map[string][]dock.AppStatus{}

	nr := dock.GlobalInfo{}
	as.Get("/global").To(&nr)
	nodes = nr.Nodes
	nsr := struct{
		Nodes map[string]*dock.NodeStatus
	}{}
	as.Get("/global/status").To(&nsr)
	nodeStatus = nsr.Nodes
	as.Get("/"+ctxName).To(&ctx)
	rr := as.Get("/"+ctxName+"/status")
	rr.To(&ctxRuns)
}

func getOut() string {
	b, _ := json.MarshalIndent(s.Arr{nodes, nodeStatus, ctx, ctxRuns}, "", "  ")
	return string(b)
}

func TestLoad(tt *testing.T) {
	t := s.T(tt)

	getStatus("c1")
	t.Test(len(nodes) == 1 && nodes["node1"] != nil && nodes["node1"].Cpu == 4 && nodeStatus["node1"].TotalRuns == 1, "Load nodes", getOut())
	t.Test(len(ctxRuns) == 1 && ctx.Apps["app1"] != nil && len(ctxRuns["app1"]) == 1 && ctxRuns["app1"][0].Node == "node1", "Load app runs", getOut())
}

func TestBase(tt *testing.T) {
	t := s.T(tt)

	nodes["node2"] = &dock.NodeInfo{Cpu:8, Memory:16}
	// 添加节点
	as.Post("/global", s.Map{"nodes": nodes})
	getStatus("c1")
	t.Test(len(nodes) == 2 &&
		nodes["node1"] != nil && nodes["node1"].Cpu == 4 &&
		nodes["node2"] != nil && nodes["node2"].Memory == 16,
		"Add Nodes", getOut())

	// 添加应用
	ctx.Apps["app2"] = &dock.AppInfo{Cpu:2,Memory:4,Min:2,Max:4,Active:true,Args:"..."}
	as.Post("/c1", ctx)
	getStatus("c1")
	t.Test(len(ctx.Apps) == 2 && ctx.Apps["app2"] != nil && ctx.Apps["app2"].Min == 2 &&
		len(ctxRuns["app2"]) == 2 && strings.Index(ctxRuns["app2"][0].Id+ctxRuns["app2"][1].Id, "<01>") != -1,
		"Add App 1&2", getOut())
}

func TestBaseApi(tt *testing.T) {
	t := s.T(tt)

	// post node9
	nodes["node9"] = &dock.NodeInfo{Cpu:8, Memory:16}
	as.Post("/global", s.Map{"nodes": nodes})
	getStatus("c1")
	t.Test(len(nodes) == 3, "Add Nodes By API", getOut())

	// post app9
	vv9 := "999999"
	ctx.Vars["vv9"] = &vv9
	ctx.Binds["app9"] = append(ctx.Binds["app9"], "node9", "node9")
	ctx.Apps["app9"] = &dock.AppInfo{Cpu:2,Memory:2,Min:2,Max:2,Active:true,Args:"${vv9}"}
	as.Post("/c1", ctx)
	getStatus("c1")
	t.Test(len(ctx.Apps) == 3 && ctx.Apps["app9"] != nil && ctx.Apps["app9"].Min == 2 &&
		len(ctxRuns["app9"]) == 2 && ctxRuns["app9"][0].Node == "node9" && ctxRuns["app9"][1].Node == "node9" &&
		strings.Index(ctxRuns["app9"][0].Id+ctxRuns["app9"][1].Id, "<01>") != -1,
		"Add App9 By API", getOut())

	ctx.Binds["app9:2#2"] = append(ctx.Binds["app9"], "node9", "node9")
	ctx.Apps["app9:2#2"] = &dock.AppInfo{Cpu:2,Memory:2,Min:2,Max:2,Active:true,Args:"${vv9}"}
	as.Post("/c1", ctx)
	getStatus("c1")
	t.Test(len(ctx.Apps) == 4 && ctx.Apps["app9:2#2"] != nil && ctx.Apps["app9:2#2"].Min == 2 &&
		len(ctxRuns["app9:2#2"]) == 2 && ctxRuns["app9:2#2"][0].Node == "node9" && ctxRuns["app9:2#2"][1].Node == "node9" &&
		strings.Index(ctxRuns["app9:2#2"][0].Id+ctxRuns["app9:2#2"][1].Id, "<03>") != -1,
		"Add App9 By API Replace", getOut())

	// post vv9
	vv9 = "999999-999"
	ctx.Vars["vv9"] = &vv9
	ctx.Binds["app9:2#3"] = append(ctx.Binds["app9"], "node9", "node9")
	ctx.Apps["app9:2#3"] = &dock.AppInfo{Cpu:2,Memory:2,Min:2,Max:2,Active:true,Args:"${vv9}"}
	ctx.Apps["app9:2#2"].Active = false
	as.Post("/c1", ctx)
	getStatus("c1")
	t.Test(len(ctx.Apps) == 5 && ctx.Apps["app9:2#3"] != nil && ctx.Apps["app9:2#3"].Min == 2 &&
		len(ctxRuns["app9"]) == 2 &&
		len(ctxRuns["app9:2#2"]) == 0 &&
		len(ctxRuns["app9:2#3"]) == 2 && ctxRuns["app9:2#3"][0].Node == "node9" && ctxRuns["app9:2#3"][1].Node == "node9" &&
		strings.Index(ctxRuns["app9:2#3"][0].Id+ctxRuns["app9:2#3"][1].Id, "<05>") != -1,
		"Update Var vv9 By API", getOut())

	// 删除 app9:2#2
	delete(ctx.Apps, "app9:2#2")
	delete(ctx.Apps, "app9:2#3")
	delete(ctx.Binds, "app9:2#3")
	as.Post("/c1", ctx)
	getStatus("c1")
	t.Test(len(ctx.Apps) == 3 && ctx.Apps["app9:2#2"] == nil && ctx.Binds["app9:2#2"] != nil && ctx.Apps["app9:2#3"] == nil && ctx.Binds["app9:2#3"] == nil, "Remove App 9#2&#3 By API", getOut())

	ctx.Apps["app8"] = &dock.AppInfo{Cpu:2,Memory:2,Min:5,Max:5,Active:true,Args:""}
	as.Post("/c1", ctx)
	getStatus("c1")
	t.Test(len(ctx.Apps) == 4 && ctx.Apps["app8"] != nil && len(ctxRuns["app8"]) == 5,
		"Add App8", getOut())

	// 删除 node9
	delete(nodes, "node9")
	as.Post("/global", s.Map{"nodes": nodes})
	time.Sleep(time.Millisecond * 100)
	getStatus("c1")
	t.Test(len(nodes) == 2 && len(ctx.Apps) == 4 && len(ctxRuns["app9"]) == 2 && len(ctxRuns["app8"]) == 5, "Remove Node 9 By API")
	has9 := false
	for _, st := range ctxRuns["app8"] {
		if st.Node == "node9" {
			has9 = true
			break
		}
	}
	t.Test(!has9, "Remove Node 9 By API, App8 Moved")
	// TODO 删除节点时，清除该节点上的应用
	//t.Test(len(nodes) == 2 && len(ctx.Apps) == 2 && ctx.Apps["app9"] == nil, "Remove Node 9 By API")

	// 删除 app9
	delete(ctx.Apps, "app9")
	as.Post("/c1", ctx)
	getStatus("c1")
	t.Test(len(ctx.Apps) == 3 && ctx.Apps["app9"] == nil, "Remove App 9 By API", getOut())

	// 删除 app8
	delete(ctx.Apps, "app8")
	as.Post("/c1", ctx)
	getStatus("c1")
	t.Test(len(ctx.Apps) == 2 && ctx.Apps["app8"] == nil, "Remove App 8 By API", getOut())
}

func TestBind(tt *testing.T) {
	t := s.T(tt)

	// 添加带 -v 的应用
	ctx.Apps["app3"] = &dock.AppInfo{Cpu:1,Memory:1,Min:1,Max:1,Active:true,Args:"... -v ..."}
	ctx.Apps["app4"] = &dock.AppInfo{Cpu:2,Memory:4,Min:2,Max:4,Active:true,Args:"... --volume ..."}
	as.Post("/c1", ctx)
	getStatus("c1")
	t.Test(len(ctx.Apps) == 4 && ctx.Apps["app4"] != nil && ctx.Apps["app4"].Min == 2 &&
		len(ctxRuns["app3"]) == 1 && len(ctxRuns["app4"]) == 2,
		"Add App 3&4 With -v")
	app3Node := ctxRuns["app3"][0].Node
	app4Node1 := ctxRuns["app4"][0].Node
	app4Node2 := ctxRuns["app4"][1].Node

	// 删除 app 2 & 3 & 4
	delete(ctx.Apps, "app2")
	delete(ctx.Apps, "app3")
	delete(ctx.Apps, "app4")
	as.Post("/c1", ctx)
	getStatus("c1")
	t.Test(len(ctx.Apps) == 1 && ctx.Apps["app4"] == nil && ctx.Apps["app1"] != nil,
		"Remove App 2&3&4")

	// 添加带 -v 的应用
	ctx.Apps["app3"] = &dock.AppInfo{Cpu:1,Memory:1,Min:1,Max:1,Active:true,Args:"... -v ..."}
	ctx.Apps["app4"] = &dock.AppInfo{Cpu:2,Memory:4,Min:2,Max:4,Active:true,Args:"... --volume ..."}
	as.Post("/c1", ctx)
	getStatus("c1")
	t.Test(len(ctx.Apps) == 3 && ctx.Apps["app4"] != nil &&
		len(ctxRuns["app3"]) == 1 && app3Node == ctxRuns["app3"][0].Node &&
		len(ctxRuns["app4"]) == 2 &&
		(app4Node1 == ctxRuns["app4"][0].Node || app4Node1 == ctxRuns["app4"][1].Node ) &&
		(app4Node2 == ctxRuns["app4"][0].Node || app4Node2 == ctxRuns["app4"][1].Node ),
		"Add App 3&4 For Bind")

	// 删除 app 2 & 3 & 4
	delete(ctx.Apps, "app3")
	delete(ctx.Apps, "app4")
	as.Post("/c1", ctx)
	time.Sleep(time.Millisecond * 100)
	getStatus("c1")
	t.Test(len(ctx.Apps) == 1 && ctx.Apps["app4"] == nil && ctx.Apps["app1"] != nil,
		"Remove App 2&3&4")
}

func TestEnd(tt *testing.T) {
	as.Stop()
	dock.AsyncStop()
}
