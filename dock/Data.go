package dock

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/ssgo/u"
	"io"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
)

type NodeInfo struct {
	Cpu    float32
	Memory float32
}
type NodeStatus struct {
	UsedCpu    float32
	UsedMemory float32
	TotalRuns  int
}

type ContextInfo struct {
	Name  string
	Desc  string
	Token string
	Vars  map[string]*string
	Binds map[string][]string
	Apps  map[string]*AppInfo
}

type AppInfo struct {
	Cpu     float32
	Memory  float32
	Min     int
	Max     int
	Args    string
	Command string
	Memo    string
	Active  bool
}
type AppStatus struct {
	Name   string
	Image  string
	Id     string
	Ctx    string
	Node   string
	UpTime string
	Cpu    float32
	Memory float32
	IsBind bool
}

var globalVars map[string]*string
var globalArgs string

var nodes = map[string]*NodeInfo{}
var nodeStatus = map[string]*NodeStatus{}
var ctxList = map[string]string{}
var ctxs = map[string]*ContextInfo{}
var ctxRuns = map[string]map[string][]*AppStatus{}
var stoppingCtxApps = map[string]map[string]*AppInfo{}
var stoppingNodes = map[string]*NodeInfo{}
var stoppingNodeStatus = map[string]*NodeStatus{}

var nodesSafely atomic.Value
var nodeStatusSafely atomic.Value
var ctxListSafely atomic.Value
var ctxsSafely atomic.Value
var ctxRunsSafely atomic.Value

func newContext() *ContextInfo {
	ctx := new(ContextInfo)
	ctx.Apps = make(map[string]*AppInfo)
	ctx.Binds = make(map[string][]string)
	ctx.Vars = make(map[string]*string)
	return ctx
}

func dataPath(names ...string) string {
	return fmt.Sprintf("%s%c%s", hubConfig.DataPath, os.PathSeparator, strings.Join(names, string(os.PathSeparator)))
}

func checkPath(file string) {
	pos := strings.LastIndexByte(file, '/')
	if pos < 0 {
		return
	}
	path := file[0:pos]
	if _, err := os.Stat(path); err != nil {
		_ = os.MkdirAll(path, 0700)
	}
}

func load(file string, to interface{}) {
	file = fmt.Sprintf("%s/%s", hubConfig.DataPath, file)
	//file = dataPath(file)
	checkPath(file)

	fp, err := os.Open(file)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(fp)
	data := map[string]interface{}{}
	err = decoder.Decode(&data)
	if err != nil {
		logError("load file failed: "+err.Error(), "file", file)
	}
	_ = fp.Close()
	err = mapstructure.WeakDecode(&data, to)
	if err != nil {
		logError("load file decode failed: "+err.Error(), "file", file)
	}
}

func save(file string, data interface{}) {
	file = fmt.Sprintf("%s/%s", hubConfig.DataPath, file)
	//file = dataPath(file)
	checkPath(file)

	fp, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		logError(err.Error())
	} else {
		n, err := fp.Write(b)
		if err != nil {
			logError(err.Error(), "file", file, "wrote", n, data, string(b))
		}
	}
	_ = fp.Close()
}

func remove(file string) {
	file = fmt.Sprintf("%s/%s", hubConfig.DataPath, file)
	//file = dataPath(file)
	err := os.Remove(file)
	if err != nil {
		logError(err.Error(), "file", file)
	}
}

func incr(file string) int {
	file = fmt.Sprintf("%s/.incr/%s", hubConfig.DataPath, file)
	//file = dataPath(".incr", file)
	checkPath(file)

	fp, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return u.GlobalRand1.Intn(999999)
	}
	buf := make([]byte, 20)
	n, err := fp.Read(buf)
	i := 0
	//if err != nil {
	//fp.Close()
	//return u.Rander.Intn(999999)
	//}

	if err == nil {
		i, err = strconv.Atoi(string(buf[0:n]))
		if err != nil {
			i = 0
		}
	}

	if i >= 999999 {
		i = 0
	}

	i++
	s := strconv.Itoa(i)
	_, _ = fp.Seek(0, io.SeekStart)
	n, err = fp.Write([]byte(s))
	if err != nil {
		logError(err.Error(), "file", file, "wrote", n)
	}
	_ = fp.Close()
	return i
}
