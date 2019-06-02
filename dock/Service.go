package dock

import (
	"fmt"
	"github.com/ssgo/s"
	"github.com/ssgo/u"
	"net/http"
	"strings"
	"time"
)

func Registers() {
	s.SetAuthChecker(auth)
	s.Static("/", "www/")
	s.Restful(0, "POST", "/login", login)

	s.Restful(1, "GET", "/global/status", getGlobalStatus)
	s.Restful(1, "GET", "/global", getGlobalInfo)
	s.Restful(1, "GET", "/contexts", getContextList)
	s.Restful(1, "GET", "/{name}", getContext)
	s.Restful(1, "GET", "/{name}/status", getContextRuns)

	s.Restful(2, "POST", "/global", setGlobalInfo)
	s.Restful(2, "POST", "/{name}", setContext)
	s.Restful(2, "DELETE", "/{name}", removeContext)
}

func auth(authLevel int, url *string, in *map[string]interface{}, request *http.Request) bool {
	switch authLevel {
	//case 1:
	//	return request.Header.Get("Access-Token") == dockConfig.AccessToken || request.Header.Get("Access-Token") == dockConfig.ManageToken
	case 1, 2:
		return request.Header.Get("Access-Token") == dockConfig.ManageToken
	}
	return false
}

func login(request *http.Request) int {
	//if request.Header.Get("Access-Token") == dockConfig.AccessToken {
	//	return 1
	//}
	if request.Header.Get("Access-Token") == dockConfig.ManageToken {
		return 2
	}
	return 0
}

type GlobalInfo struct {
	Nodes map[string]*NodeInfo
	Vars  map[string]*string
	Args  string
}

func getGlobalInfo() (out struct {
	GlobalInfo
	PublicKey string
}) {
	out.Nodes = nodesSafely.Load().(map[string]*NodeInfo)
	out.Vars = globalVars
	out.Args = globalArgs
	out.PublicKey, _ = u.ReadFile(dataPath(".ssh", "id_dsa.pub"), 2048)
	return
}

type globalStatusResult struct {
	Nodes map[string]*NodeStatus
}

func getGlobalStatus() globalStatusResult {
	return globalStatusResult{
		Nodes: nodeStatusSafely.Load().(map[string]*NodeStatus),
	}
}

func getContextList() map[string]string {
	return ctxListSafely.Load().(map[string]string)
}

func getContext(in struct{ Name string }) *ContextInfo {
	ctxsTemp := ctxsSafely.Load().(map[string]*ContextInfo)
	return ctxsTemp[in.Name]
}

func getContextRuns(in struct{ Name string }) map[string][]*AppStatus {
	ctxRunsTemp := ctxRunsSafely.Load().(map[string]map[string][]*AppStatus)
	return ctxRunsTemp[in.Name]
}

type SetResult struct {
	Ok    bool
	Error string
}

func setGlobalInfo(in GlobalInfo) SetResult {
	makingLocker.Lock()

	if makeAppRunningInfos(true) {
		// 如果有在要去掉节点上的应用，存储到 stoppingNodes
		//changedCtxs := make([]string, 0)
		for nodeName, node := range nodes {
			if in.Nodes[nodeName] == nil {
				stoppingNodes[nodeName] = node
			}
		}

		//for _, ctxName := range changedCtxs {
		//	checkContext(ctxName)
		//}

		nodes = in.Nodes
		for nodeName := range nodes {
			if nodeStatus[nodeName] == nil {
				nodeStatus[nodeName] = &NodeStatus{UsedCpu: 0, UsedMemory: 0}
			}
		}
		for nodeName := range nodeStatus {
			if nodes[nodeName] == nil {
				delete(nodeStatus, nodeName)
			}
		}
		nodesSafely.Store(nodes)

		globalVars = in.Vars
		globalArgs = in.Args
	}
	makingLocker.Unlock()

	//save("nodes", nodes)
	save("global", in)
	save(fmt.Sprintf("bak/global/%s", time.Now().Format("2006-01/02/15:04:05")), in)
	return SetResult{Ok: true}
}

func setContext(in ContextInfo) SetResult {
	// not support - / , because docker id need -
	if in.Name == "" || in.Name == "global" || in.Name == "nodes" || strings.IndexByte(in.Name, '-') != -1 || strings.IndexByte(in.Name, '/') != -1 {
		return SetResult{Error: "bad name"}
	}

	makingLocker.Lock()

	var checkSucceed bool
	var checkChanged bool
	var err error
	if makeAppRunningInfos(true) {

		ctxList[in.Name] = in.Desc

		if ctxRuns[in.Name] == nil {
			ctxRuns[in.Name] = make(map[string][]*AppStatus)
		}
		if stoppingCtxApps[in.Name] == nil {
			stoppingCtxApps[in.Name] = make(map[string]*AppInfo)
		}

		prevCtx := ctxs[in.Name]
		ctxs[in.Name] = &in

		// 立刻更新
		checkChanged, checkSucceed, err = checkContext(in.Name)
		if checkSucceed {
			if checkChanged {
				nodesSafely.Store(nodes)
				nodeStatusSafely.Store(nodeStatus)
				ctxListSafely.Store(ctxList)
				ctxsSafely.Store(ctxs)
				ctxRunsSafely.Store(ctxRuns)
				showStats()
			}
		} else {
			ctxs[in.Name] = prevCtx
		}
	}
	makingLocker.Unlock()

	if !checkSucceed {
		if err == nil {
			return SetResult{Error: ""}
		} else {
			return SetResult{Error: err.Error()}
		}
	}

	save(in.Name, ctxs[in.Name])
	save(fmt.Sprintf("bak/%s/%s", in.Name, time.Now().Format("2006-01/02/15:04:05")), ctxs[in.Name])
	return SetResult{Ok: true}
}

func removeContext(in struct{ Name string }) SetResult {
	if in.Name == "" || in.Name == "global" || in.Name == "nodes" || ctxs[in.Name] == nil {
		return SetResult{Error: "bad name"}
	}

	// 立刻更新，停掉所有节点
	var checkChanged bool
	var err error
	makingLocker.Lock()
	if makeAppRunningInfos(true) {
		ctxs[in.Name].Apps = make(map[string]*AppInfo)
		checkChanged, _, err = checkContext(in.Name)
		if checkChanged {
			nodesSafely.Store(nodes)
			nodeStatusSafely.Store(nodeStatus)
			ctxListSafely.Store(ctxList)
			ctxsSafely.Store(ctxs)
			ctxRunsSafely.Store(ctxRuns)
			showStats()
		}
	}
	makingLocker.Unlock()

	delete(ctxList, in.Name)
	delete(ctxs, in.Name)
	delete(ctxRuns, in.Name)
	remove(in.Name)

	if err != nil {
		return SetResult{Error: err.Error()}
	}
	return SetResult{Ok: true}
}
