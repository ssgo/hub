package dock

import (
	"fmt"
	"github.com/ssgo/u"
	golog "log"
	"strconv"
	"strings"
)

func makeAppRunningInfos(isAll bool) bool {
	// 重置运行信息
	var makingNodes map[string]*NodeInfo
	var makingNodeStatus map[string]*NodeStatus
	if isAll {
		makingNodes = map[string]*NodeInfo{}
		for nodeName, node := range stoppingNodes {
			makingNodes[nodeName] = node
		}
		for nodeName, node := range nodes {
			makingNodes[nodeName] = node
		}
		makingNodeStatus = nodeStatus
		ctxRuns = map[string]map[string][]*AppStatus{}
	} else {
		makingNodes = stoppingNodes
		makingNodeStatus = stoppingNodeStatus
	}

	// 更新 Node.UsedCpu、Node.UsedMemory、runs 信息
	okNum := 0
	for nodeName := range makingNodes {
		nodeStat := NodeStatus{UsedCpu: 0, UsedMemory: 0}
		runningApps, err := getRunningApps(nodeName)
		if err != nil {
			continue
		}
		okNum++
		for _, running := range runningApps {
			ctx := ctxs[running.Ctx]
			if ctx == nil {
				continue
			}

			if !checkRun(running.Ctx, running) {
				continue
			}

			// 初始化ctxRuns
			runsByApp := ctxRuns[running.Ctx]
			if runsByApp == nil {
				runsByApp = map[string][]*AppStatus{}
				ctxRuns[running.Ctx] = runsByApp
			}

			// 删除的应用，进入停止队列
			app := ctx.Apps[running.Image]
			if app == nil {
				if stoppingCtxApps[running.Ctx] == nil {
					stoppingCtxApps[running.Ctx] = make(map[string]*AppInfo)
				}
				stoppingApps := stoppingCtxApps[running.Ctx]
				if stoppingApps[running.Image] != nil {
					// 正在重启，运行信息加入到重命名的旧应用
					app = stoppingApps[running.Image]
				}
				if app == nil {
					continue
				}
			}

			// 存入 runs
			runs := runsByApp[running.Image]
			if runs == nil {
				runs = make([]*AppStatus, 0)
			}
			// stoppingNodes 里面 未绑定部分不在本次处理范畴
			isStoppingInBinds := false
			if nodes[running.Node] == nil && stoppingNodes[running.Node] != nil {
				isStoppingInBinds = ctx.Binds[running.Image] != nil && !findIn(ctx.Binds[running.Image], running.Node)
			}

			if (isAll && isStoppingInBinds) || (!isAll && !isStoppingInBinds) {
				continue
			}
			// 更新信息
			nodeStat.TotalRuns++
			nodeStat.UsedCpu += app.Cpu
			nodeStat.UsedMemory += app.Memory
			runs = append(runs, running)
			runsByApp[running.Image] = runs
		}
		makingNodeStatus[nodeName] = &nodeStat
	}
	return len(makingNodes) == 0 || okNum > 0
}

func checkContext(ctxName string) (bool, bool, error) {
	changed := false
	ctx := ctxs[ctxName]
	runsByApp := ctxRuns[ctxName]
	if ctx == nil {
		return false, true, nil
	}
	if runsByApp == nil {
		runsByApp = map[string][]*AppStatus{}
		ctxRuns[ctxName] = runsByApp
	}

	// 启动需要的App
	for appName := range ctx.Apps {
		startChanged, startSucceed, err := checkAppForStart(ctxName, appName)
		if startChanged {
			changed = true
		}
		if startSucceed == false {
			return changed, false, err
		}
	}

	var ok bool
	var err error
	for appName := range runsByApp {
		if ok, err = checkAppForStop(ctxName, appName); ok {
			changed = true
		}
	}

	//// 清除多余的绑定信息
	//for appName := range ctx.Binds {
	//	if ctx.Apps[appName] != nil {
	//		delete(ctx.Binds, appName)
	//		save(ctxName, ctx)
	//	}
	//}

	// TODO 根据实际负债情况进行弹性伸缩

	return changed, true, err
}

func checkAppForStart(ctxName, appName string) (bool, bool, error) {

	changed := false
	ctx := ctxs[ctxName]
	runsByApp := ctxRuns[ctxName]
	if ctx == nil || runsByApp == nil {
		return false, true, nil
	}

	app := ctx.Apps[appName]
	runs := runsByApp[appName]
	if app == nil || app.Active == false {
		return false, true, nil
	}

	if runs == nil {
		runs = make([]*AppStatus, 0)
		runsByApp[appName] = runs
	}

	avaliableBinds := map[string]int{}
	appendBinds := make([]string, 0)

	if ctx.Binds[appName] != nil {
		for _, b := range ctx.Binds[appName] {
			avaliableBinds[b]++
		}
	}

	if len(avaliableBinds) > 0 && len(runs) > 0 {
		for _, run := range runs {
			if avaliableBinds[run.Node] > 0 {
				run.IsBind = true
				// 抵消已经分配的
				avaliableBinds[run.Node]--
				if avaliableBinds[run.Node] <= 0 {
					delete(avaliableBinds, run.Node)
				}
			} else {
				// 将未绑定的信息添加到 _ctx.Binds
				appendBinds = append(appendBinds, run.Node)
			}
		}
	}

	var err error
	for i := len(runs); i < app.Min; i++ {
		// 如果有绑定节点优先使用
		nodeName := ""
		isBind := false
		if len(avaliableBinds) > 0 {
			for tmpNodeName := range avaliableBinds {
				if avaliableBinds[tmpNodeName] > 0 {
					avaliableBinds[tmpNodeName]--
					nodeName = tmpNodeName
					if avaliableBinds[tmpNodeName] <= 0 {
						delete(avaliableBinds, tmpNodeName)
					}
					if nodeName != "" {
						isBind = true
						break
					}
				}
			}
		}

		// 没有绑定，使用得分最低的一个节点
		if nodeName == "" {
			nodeName = nextMinScoreNode(ctxName, appName)
			// 无节点可用
			if nodeName == "" {
				break
			}

			// 挂载了磁盘的应用，将其绑定在该节点上
			if strings.Index(app.Args, " -v ") != -1 || strings.Index(app.Args, " --volume ") != -1 {
				appendBinds = append(appendBinds, nodeName)
			}
		}

		var id, runName string
		if nodes[nodeName] != nil {
			// 如果有冲突的绑定应用，先停止旧的
			if ctx.Binds[appName] != nil && len(ctx.Binds[appName]) > 0 {
				pos := strings.IndexByte(appName, ':')
				if pos == -1 {
					pos = strings.IndexByte(appName, '#')
				}
				oldAppNamePrefix := appName
				if pos > 0 {
					oldAppNamePrefix = appName[0:pos]
				}
				for oldAppName := range runsByApp {
					if oldAppName != appName && strings.HasPrefix(oldAppName, oldAppNamePrefix) && ctx.Binds[oldAppName] != nil && len(ctx.Binds[oldAppName]) > 0{
						if strings.Join(ctx.Binds[oldAppName], ",") == strings.Join(ctx.Binds[appName], ",") {
							_, _ = checkAppForStop(ctxName, oldAppName)
						}
					}
				}
			}

			id, runName, err = startApp(ctxName, appName, nodeName, app)
			if id == "" || runName == "" {
				// 启动失败将不执行后面的过程
				return changed, false, err
			}
			changed = true
		}

		// 启动失败的 App，暂时占着坑，下次会重新尝试启动
		run := AppStatus{Name: runName, Ctx: ctxName, Node: nodeName, Id: id, Image: appName, UpTime: "Up 0 minutes", Cpu: app.Cpu, Memory: app.Memory, IsBind: isBind}
		nodeStat := nodeStatus[nodeName]
		if nodeStat != nil {
			nodeStat.TotalRuns++
			nodeStat.UsedCpu += app.Cpu
			nodeStat.UsedMemory += app.Memory
		}
		runs = append(runs, &run)
		runsByApp[appName] = runs
	}

	// 保存增加的绑定信息
	if len(appendBinds) > 0 {
		if ctx.Binds[appName] == nil {
			ctx.Binds[appName] = make([]string, 0)
		}
		ctx.Binds[appName] = append(ctx.Binds[appName], appendBinds...)
	}
	return changed, true, err
}

func checkAppForStop(ctxName, appName string) (bool, error) {
	ctx := ctxs[ctxName]
	runsByApp := ctxRuns[ctxName]
	if ctx == nil || runsByApp == nil {
		return false, nil
	}
	stoppingApps := stoppingCtxApps[ctxName]

	app := ctx.Apps[appName]
	runs := runsByApp[appName]
	changed := false

	var ok bool
	var err error
	if app == nil || app.Active == false {
		changed = true
		// 停掉已经不需要的App
		allDone := true
		//i := 0
		if runs != nil {
			leftRuns := make([]*AppStatus, 0)
			for _, run := range runs {
				//i++
				if ok, err = stopApp(ctxName, run); ok == false {
					// 停止失败，后续会再次尝试
					allDone = false
					leftRuns = append(leftRuns, run)
				} else {
					// 停止成功
					nodeStat := nodeStatus[run.Node]
					if nodeStat != nil {
						nodeStat.TotalRuns--
						nodeStat.UsedCpu -= run.Cpu
						nodeStat.UsedMemory -= run.Memory
					}
				}
			}
			runsByApp[appName] = leftRuns
		}

		if allDone {
			if app == nil {
				delete(runsByApp, appName)
			}
			if stoppingApps != nil && stoppingApps[appName] != nil {
				delete(stoppingApps, appName)
			}
		}
	} else {
		// 停掉多余的App
		if len(runs) > app.Max {
			for i := len(runs) - 1; i >= app.Max; i-- {
				if runs[i].Id != "" {
					changed = true

					if ok, err = stopApp(ctxName, runs[i]); ok {
						runs[i].Id = ""
					}
				}
			}
		}
	}
	return changed, err
}

func checkAppForStoppingNodes(ctxName, appName string) (bool, error) {
	ctx := ctxs[ctxName]
	runsByApp := ctxRuns[ctxName]
	if ctx == nil || runsByApp == nil {
		return false, nil
	}

	changed := false
	runs := runsByApp[appName]
	var ok bool
	var err error
	if runs != nil {
		leftRuns := make([]*AppStatus, 0)
		for _, run := range runs {
			if nodes[run.Node] == nil && stoppingNodes[run.Node] != nil && !strings.Contains(strings.Join(ctx.Binds[appName], " ")+" ", run.Node+" ") {
				if ok, err = stopApp(ctxName, run); ok == false {
					// 停止失败，后续会再次尝试
					leftRuns = append(leftRuns, run)
				} else {
					// 停止成功
					nodeStat := stoppingNodeStatus[run.Node]
					if nodeStat != nil {
						nodeStat.TotalRuns--
						nodeStat.UsedCpu -= run.Cpu
						nodeStat.UsedMemory -= run.Memory
					}
				}
			} else {
				// 正常的实例
				leftRuns = append(leftRuns, run)
			}
		}
		runsByApp[appName] = leftRuns
	}

	return changed, err
}

func nextMinScoreNode(ctxName, appName string) string {
	ctx := ctxs[ctxName]
	runsByApp := ctxRuns[ctxName]
	if ctx == nil || runsByApp == nil {
		return ""
	}

	app := ctx.Apps[appName]
	runs := runsByApp[appName]
	if app == nil || runs == nil {
		return ""
	}

	var minScore float32 = -1
	minNodeName := ""
	for nodeName, node := range nodes {
		nodeStat := nodeStatus[nodeName]
		if nodeStat == nil {
			continue
		}
		score := nodeStat.UsedMemory/node.Memory + nodeStat.UsedCpu/node.Cpu
		for _, run := range runs {
			// 已经有过的节点得分 +1，优先考虑平均分布
			if run.Node == nodeName {
				if strings.Index(app.Args, " -v ") != -1 || strings.Index(app.Args, " --volume ") != -1 {
					// 挂载磁盘的，尽可能的分布到不同节点，增加 10000% 权重
					score += 100
				} else {
					if app.Min <= 2 {
						// 2个节点 强平均分配，增加 300% 权重
						score += 3
					} else if app.Min <= 4 {
						// 3~4个节点 较强平均分配，增加 150% 权重
						score += 1.5
					} else if app.Min <= 6 {
						// 5~6个节点 略强平均分配，增加 80% 权重
						score += 0.8
					} else {
						// 7个及以上节点 弱平均分配，增加 30% 权重
						score += 0.3
					}
				}
			}
		}
		if minScore < 0 || score < minScore {
			minScore = score
			minNodeName = nodeName
		}
	}
	return minNodeName
}

func showStats() {
	outs := make([]string, 0)
	maxNodeNameLen := -1
	for nodeName, _ := range nodes {
		if maxNodeNameLen == -1 || len(nodeName) > maxNodeNameLen {
			maxNodeNameLen = len(nodeName)
		}
	}
	//b, _ := json.MarshalIndent(nodeStatus, "", "  ")
	//log.Println("	==========2	", string(b))

	outs = append(outs, fmt.Sprintf(">>  \033[7m[%s]\033[0m", "Nodes"))
	for nodeName, node := range nodes {
		nodeStat := nodeStatus[nodeName]
		if nodeStat == nil {
			continue
		}
		outs = append(outs, fmt.Sprintf(">>    %"+strconv.Itoa(maxNodeNameLen)+"s  %d  %.2f / %.2f  %.2f / %.2f", nodeName, nodeStat.TotalRuns, nodeStat.UsedCpu, node.Cpu, nodeStat.UsedMemory, node.Memory))
	}

	for ctxName, ctx := range ctxs {
		runsByApp := ctxRuns[ctxName]
		if ctx == nil || runsByApp == nil {
			continue
		}
		maxNameLen := -1
		for appName := range ctx.Apps {
			if maxNameLen == -1 || len(appName) > maxNameLen {
				maxNameLen = len(appName)
			}
		}
		for varName := range ctx.Vars {
			if maxNameLen == -1 || len(varName)+1 > maxNameLen {
				maxNameLen = len(varName) + 1
			}
		}
		for bindName := range ctx.Binds {
			if maxNameLen == -1 || len(bindName)+1 > maxNameLen {
				maxNameLen = len(bindName) + 1
			}
		}

		outs = append(outs, fmt.Sprintf(">>  \n>>  "+u.White("[%s]"), ctxName))
		for varName, varValue := range ctx.Vars {
			outs = append(outs, fmt.Sprintf(">>    "+u.Yellow("$%-"+strconv.Itoa(maxNameLen-1)+"s")+"  %s", varName, *varValue))
		}
		for bindName, bindValues := range ctx.Binds {
			outs = append(outs, fmt.Sprintf(">>    *%-"+strconv.Itoa(maxNameLen-1)+"s  %s", bindName, strings.Join(bindValues, ",")))
		}
		for appName, app := range ctx.Apps {
			runs := runsByApp[appName]
			if runs == nil {
				continue
			}
			outs = append(outs, fmt.Sprintf(">>    "+u.Cyan("-"+strconv.Itoa(maxNameLen)+"s")+"  %d (%d ~ %d)  %.2f, %.2f  %s  %s	%s",
				appName, len(runs), app.Min, app.Max, app.Cpu, app.Memory, app.Args, app.Command, app.Memo))
			for _, run := range runs {
				outs = append(outs, fmt.Sprintf(">>      "+u.Cyan("%10s")+"  %12s  %"+strconv.Itoa(maxNodeNameLen)+"s%s  %s",
					run.Name, run.Id, run.Node, u.StringIf(run.IsBind, "*", ""), run.UpTime))
			}
		}
	}

	logInfo("status",
		"nodes", nodes,
		"nodeStatus", nodeStatus,
		"globalVars", globalVars,
		"globalArgs", globalArgs,
		"contexts", ctxsSafely.Load(),
		"contextRuns", ctxRuns,
	)
	golog.Print("Status\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n>>  \n", strings.Join(outs, "\n"), "\n>>  \n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")
}

func findIn(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}
