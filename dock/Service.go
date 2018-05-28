package dock

import (
	"github.com/ssgo/s"
	"net/http"
	"strings"
	"fmt"
	"time"
)

func Registers() {
	s.SetAuthChecker(auth)
	s.Static("/", "www")
	s.Restful(0, "POST", "/login", login)

	s.Restful(1, "GET", "/nodes/status", getNodeStatus)
	s.Restful(1, "GET", "/nodes", getNodeList)
	s.Restful(1, "GET", "/contexts", getContextList)
	s.Restful(1, "GET", "/{name}/status", getContextRuns)
	s.Restful(1, "GET", "/{name}", getContext)

	s.Restful(2, "POST", "/nodes", setNodes)
	s.Restful(2, "POST", "/{name}", setContext)
	s.Restful(2, "DELETE", "/{name}", removeContext)
}

func auth(authLevel uint, url *string, in *map[string]interface{}, request *http.Request) bool {
	switch authLevel {
	case 1:
		return request.Header.Get("Access-Token") == config.AccessToken || request.Header.Get("Access-Token") == config.ManageToken
	case 2:
		return request.Header.Get("Access-Token") == config.ManageToken
	}
	return false
}

func login(request *http.Request) int {
	if request.Header.Get("Access-Token") == config.AccessToken {
		return 1
	}
	if request.Header.Get("Access-Token") == config.ManageToken {
		return 2
	}
	return 0
}

func getNodeList() map[string]*NodeInfo {
	return nodesSafely.Load().(map[string]*NodeInfo)
}

func getNodeStatus() map[string]*NodeStatus {
	return nodeStatusSafely.Load().(map[string]*NodeStatus)
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

func setNodes(in struct{ Nodes map[string]*NodeInfo }) bool {
	makingLocker.Lock()

	makeAppRunningInfos(true)
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
	makingLocker.Unlock()

	save("nodes", nodes)
	save(fmt.Sprintf("bak/nodes/%s", time.Now().Format("2006-01/02/15:04:05")), nodes)
	return true
}

func setContext(in ContextInfo) bool {
	// not support - / , because docker id need -
	if in.Name == "" || in.Name == "nodes" || strings.IndexByte(in.Name, '-') != -1 || strings.IndexByte(in.Name, '/') != -1 {
		return false
	}

	makingLocker.Lock()

	makeAppRunningInfos(true)

	ctxList[in.Name] = in.Desc

	if ctxRuns[in.Name] == nil {
		ctxRuns[in.Name] = make(map[string][]*AppStatus)
	}
	if stoppingCtxApps[in.Name] == nil {
		stoppingCtxApps[in.Name] = make(map[string]*AppInfo)
	}

	ctxs[in.Name] = &in

	// 立刻更新
	if checkContext(in.Name) {
		nodesSafely.Store(nodes)
		nodeStatusSafely.Store(nodeStatus)
		ctxListSafely.Store(ctxList)
		ctxsSafely.Store(ctxs)
		ctxRunsSafely.Store(ctxRuns)
		showStats()
	}
	makingLocker.Unlock()

	save(in.Name, ctxs[in.Name])
	save(fmt.Sprintf("bak/%s/%s", in.Name, time.Now().Format("2006-01/02/15:04:05")), ctxs[in.Name])

	return true
}

func removeContext(in struct{ Name string }) bool {
	if in.Name == "" || in.Name == "nodes" || ctxs[in.Name] == nil {
		return false
	}

	// 立刻更新，停掉所有节点
	makingLocker.Lock()
	makeAppRunningInfos(true)
	ctxs[in.Name].Apps = make(map[string]*AppInfo)
	if checkContext(in.Name) {
		nodesSafely.Store(nodes)
		nodeStatusSafely.Store(nodeStatus)
		ctxListSafely.Store(ctxList)
		ctxsSafely.Store(ctxs)
		ctxRunsSafely.Store(ctxRuns)
		showStats()
	}
	makingLocker.Unlock()

	delete(ctxList, in.Name)
	delete(ctxs, in.Name)
	delete(ctxRuns, in.Name)
	remove(in.Name)
	return false
}
