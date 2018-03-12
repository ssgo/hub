package dock

import (
	"log"
	"strings"
	"fmt"
	"strconv"
)

type Stats struct {
	Nodes map[string]*NodeInfo
	Apps  map[string]*AppInfo
}

// TODO redis 中的配置没有覆盖 config
// TODO 支持 镜像别名（区分不同实例使用不通配置）
// TODO 支持 docker 启动脚本（redis slave）
func makeAppRunningInfo() {
	// 重置运行信息
	for _, app := range apps {
		app.Runs = make([]*AppRunInfo, 0)
	}

	// 更新 Node.UsedCpu、Node.UsedMemory、app.Runs 信息
	for nodeName, node := range nodes {
		node.UsedCpu = 0
		node.UsedMemory = 0
		for _, run := range getRunningApps(nodeName) {
			app := apps[run.Image]
			if app != nil {
				node.UsedCpu += app.Cpu
				node.UsedMemory += app.Memory
				app.Runs = append(app.Runs, run)
			}
		}
	}
}

func checkApps() bool {
	changed := false

	// 启动需要的App
	for appName, app := range apps {
		avaliableBinds := map[string]int{}
		appendBinds := make([]string, 0)
		for _, b := range app.Binds {
			avaliableBinds[b] ++
		}

		if len(avaliableBinds) > 0 && len(app.Runs) > 0 {
			for _, run := range app.Runs {
				if avaliableBinds[run.Node] > 0 {
					// 抵消已经分配的
					avaliableBinds[run.Node] --
					if avaliableBinds[run.Node] <= 0 {
						delete(avaliableBinds, run.Node)
					}
				} else {
					// 将未绑定的信息添加到 _binds
					appendBinds = append(appendBinds, run.Node)
				}
			}
		}

		for i := len(app.Runs); i < app.Min; i++ {
			// 如果有绑定节点优先使用
			nodeName := ""
			if len(avaliableBinds) > 0 {
				for tmpNodeName := range avaliableBinds {
					if avaliableBinds[tmpNodeName] > 0 {
						avaliableBinds[tmpNodeName] --
						nodeName = tmpNodeName
					}
					if avaliableBinds[tmpNodeName] <= 0 {
						delete(avaliableBinds, tmpNodeName)
					}
					if nodeName != "" {
						break
					}
				}
			}

			// 没有绑定，使用得分最低的一个节点
			if nodeName == "" {
				nodeName = nextMinScoreNode(app)
				// 无节点可用
				if nodeName == "" {
					break
				}

				// 挂载了磁盘的应用，将其绑定在该节点上
				if strings.Index(app.Args, " -v ") != -1 || strings.Index(app.Args, " --volume ") != -1 {
					appendBinds = append(appendBinds, nodeName)
				}
			}

			id := startApp(appName, nodeName, app)
			changed = true

			// 启动失败的 App，暂时占着坑，下次会重新尝试启动
			run := AppRunInfo{Node: nodeName, Id: id, Image: appName, UpTime: "Up 0 hours"}
			node := nodes[nodeName]
			if node != nil {
				node.UsedCpu += app.Cpu
				node.UsedMemory += app.Memory
			}
			app.Runs = append(app.Runs, &run)
		}

		// 保存增加的绑定信息
		if len(appendBinds) > 0 {
			dcCache.HSET("_binds", appName, strings.Join(append(app.Binds, appendBinds...), ","))
		}
	}

	for appName, app := range apps {
		//if app.Status == RUNNING {
		//	// 停掉多余的App
		//	if len(app.Runs) > app.Max {
		//		for i := len(app.Runs)-1; i >= app.Max; i-- {
		//			changed = true
		//			if stopApp(app.Runs[i], app) {
		//				app.Runs[i].Id = ""
		//			}
		//		}
		//	}
		//} else
		if app.Status == STOPPING {
			changed = true
			// 停掉已经不需要的App
			allDone := true
			for _, run := range app.Runs {
				if stopApp(run, app) == false {
					// 停止失败，后续会再次尝试
					allDone = false
				} else {
					// 停止成功
					node := nodes[run.Node]
					if node != nil {
						node.UsedCpu -= app.Cpu
						node.UsedMemory -= app.Memory
					}
				}
			}
			if allDone {
				delete(apps, appName)
			}
		}
	}

	// TODO 根据实际负债情况进行弹性伸缩
	return changed
}

func showStats() {
	outs := make([]string, 0)
	outs = append(outs, "### Nodes")
	maxNodeNameLen := -1
	for nodeName, _ := range nodes {
		if maxNodeNameLen == -1 || len(nodeName) > maxNodeNameLen {
			maxNodeNameLen = len(nodeName)
		}
	}
	for nodeName, Node := range nodes {
		outs = append(outs, fmt.Sprintf("   %"+strconv.Itoa(maxNodeNameLen)+"s   %.2f / %.2f   %.2f / %.2f", nodeName, Node.UsedCpu, Node.TotalCpu, Node.UsedMemory, Node.TotalMemory))
	}
	log.Print(strings.Join(outs, "\n"))

	outs = make([]string, 0)
	outs = append(outs, "### Apps")
	maxAppNameLen := -1
	for appName, _ := range apps {
		if maxAppNameLen == -1 || len(appName) > maxAppNameLen {
			maxAppNameLen = len(appName)
		}
	}

	for appName, app := range apps {
		outs = append(outs, fmt.Sprintf("   %"+strconv.Itoa(maxAppNameLen)+"s   %d (%d ~ %d)	 %.2f, %.2f", appName, len(app.Runs), app.Min, app.Max, app.Cpu, app.Memory))
		for i, run := range app.Runs {
			outs = append(outs, fmt.Sprintf("      %2d   %12s   %"+strconv.Itoa(maxNodeNameLen)+"s   %s", i+1, run.Id, run.Node, run.UpTime))
		}
	}
	log.Print(strings.Join(outs, "\n"))
	//out, _ := json.MarshalIndent(GetStats(), "", "  ")
	//log.Print(string(out))
}

func GetStats() Stats {
	return Stats{Nodes: nodes, Apps: apps}
}
