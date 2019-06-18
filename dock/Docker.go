package dock

import (
	"fmt"
	"regexp"
	"strings"
)

var nodeFailedTimes = map[string]int{}
var dockerNameFilter *regexp.Regexp
var dockerVarReplacer *regexp.Regexp

func getRunningApps(nodeName string) ([]*AppStatus, error) {
	runs := make([]*AppStatus, 0)
	out, _, err := shellFunc(15000, nodeName, "ps", "--format", "{{.ID}},{{.Names}},{{.Image}},{{.Status}}")
	if err != nil {
		return nil, err
	}

	//log.Println("	******	", out)
	for _, line := range strings.Split(out, "\n") {
		a1 := strings.Split(line, ",")
		if len(a1) >= 4 {
			a2 := strings.Split(a1[1], "-")
			run := &AppStatus{Ctx: a2[0], Node: nodeName, Id: a1[0], Name: a1[1], Image: a1[2], UpTime: a1[3]}
			// 通过 --name 中隐藏的 tag 信息补全 Image
			if len(a2) >= 4 {
				run.Image += "#" + a2[2]
			}
			runs = append(runs, run)
		}
	}
	return runs, nil
}

func checkRun(ctxName string, run *AppStatus) bool {
	// TODO if check failed, how to kill for root's
	//if run.Id == "" {
	//	return true
	//}
	//
	//if out := getLastLine(shellFunc(10000, run.Node, "exec", run.Id, "echo", run.Id)); out != run.Id {
	//	log.Printf("Dock	exec	echo	[%s]	error	%s	!=	%s", ctxName, out, run.Id)
	//	return false
	//}
	return true
}

func startApp(ctxName, appName, nodeName string, app *AppInfo) (string, string, error) {
	ctx := ctxs[ctxName]
	if ctx == nil {
		return "", "", nil
	}

	if dockerNameFilter == nil {
		dockerNameFilter = regexp.MustCompile("[^a-zA-Z0-9]")
	}
	if dockerVarReplacer == nil {
		dockerVarReplacer = regexp.MustCompile("\\${[a-zA-Z0-9._-]+}")
	}

	// 解析后缀
	var postfix string
	a := strings.SplitN(appName, "#", 2)
	if len(a) > 1 {
		appName = a[0]
		postfix = a[1]
	}

	// 生成 docker --name
	var dockerRunName string
	a = strings.SplitN(appName, "/", 2)
	if strings.IndexByte(a[0], '.') != -1 && len(a) > 1 {
		dockerRunName = a[1]
	} else {
		dockerRunName = a[0]
	}
	dockerRunName = dockerNameFilter.ReplaceAllString(dockerRunName, "")
	if postfix != "" {
		dockerRunName += "-" + postfix
	}
	dockerRunIndex := incr(dockerRunName)
	dockerRunName = fmt.Sprintf("%s-%s-%d", ctxName, dockerRunName, dockerRunIndex)

	args := make([]string, 0)
	args = append(args, "run", "--name", dockerRunName, "-d", "--restart=always")
	if globalArgs != "" {
		args = append(args, PraseCommandArgs(globalArgs)...)
	}
	if app.Cpu > 0.01 {
		args = append(args, "--cpus", fmt.Sprintf("%.2f", app.Cpu))
	}
	if app.Memory > 4 {
		args = append(args, "-m", fmt.Sprintf("%.0fg", app.Memory))
	} else if app.Memory > 0.01 {
		args = append(args, "-m", fmt.Sprintf("%.0fm", app.Memory*1024))
	}

	// 替换变量
	tmpArgs := app.Args
	tmpArgs = dockerVarReplacer.ReplaceAllStringFunc(tmpArgs, func(varName string) string {
		s := ctx.Vars[varName[2:len(varName)-1]]
		if s == nil {
			s = globalVars[varName[2:len(varName)-1]]
			if s == nil {
				return ""
			}
		}
		return *s
	})

	//// 解析启动参数
	//runCmd := ""
	//if strings.HasSuffix(tmpArgs, ">") {
	//	pos := strings.LastIndex(tmpArgs, " <")
	//	if pos != -1 {
	//		runCmd = tmpArgs[pos+2:len(tmpArgs)-1]
	//		tmpArgs = tmpArgs[0:pos]
	//	}
	//}

	parsedArgs := PraseCommandArgs(tmpArgs)
	for i := 0; i < len(parsedArgs); i++ {
		arg := parsedArgs[i]

		if arg == "-d" || strings.HasPrefix(arg, "--restart=") || strings.HasPrefix(arg, "--name=") || strings.HasPrefix(arg, "--cpus=") || strings.HasPrefix(arg, "--memory=") {
			continue
		}

		if arg == "--restart" || arg == "--name" || arg == "--cpus" || arg == "--memory" || arg == "-m" {
			i++
			continue
		}

		args = append(args, arg)
	}

	//args = append(args, PraseCommandArgs(tmpArgs)...)
	args = append(args, appName)
	if app.Command != "" {
		args = append(args, PraseCommandArgs(app.Command)...)
	}
	//log.Print("Dock	exec	run	[", ctxName, "]	\033[32mdocker ", strings.Join(args, " "), "\033[0m")
	shellOut, usedTime, err := shellFunc(60000, nodeName, args...)
	if err != nil {
		logError("docker run failed: "+err.Error(),
			"context", ctxName,
			"app", appName,
			"node", nodeName,
			"shell", "docker "+strings.Join(args, " "),
			"usedTime", usedTime,
			"limitTime", 60000,
			"result", shellOut,
		)
	} else {
		logInfo("docker run",
			"context", ctxName,
			"app", appName,
			"node", nodeName,
			"shell", "docker "+strings.Join(args, " "),
			"usedTime", usedTime,
			"limitTime", 60000,
			"result", shellOut,
		)
	}

	id := getLastLine(shellOut)
	if len(id) > 12 {
		id = id[0:12]
	}
	//log.Print("Dock	exec	run	[", ctxName, "]	result	", id)
	return id, dockerRunName, err
}

func stopApp(ctxName string, run *AppStatus) (bool, error) {
	if run.Id == "" {
		return true, nil
	}

	stopIsOk := true
	//log.Printf("Dock	exec	[%s]	%s	%s	\033[31mdocker stop %s\033[0m", ctxName, run.Image, run.Node, run.Id)
	shellOut, usedTime, err := shellFunc(60000, run.Node, "stop", run.Id)
	if out := getLastLine(shellOut); out != run.Id {
		//log.Printf("Dock	exec	stop	[%s]	error	%s	!=	%s", ctxName, out, run.Id)
		stopIsOk = false
	}
	if err != nil {
		logError("docker stop failed: "+err.Error(),
			"context", ctxName,
			"app", run.Image,
			"id", run.Id,
			"node", run.Node,
			"name", run.Name,
			"upTime", run.UpTime,
			"shell", "docker stop "+run.Id,
			"usedTime", usedTime,
			"limitTime", 60000,
			"result", shellOut,
			"stopIsOk", stopIsOk,
		)
	} else {
		logInfo("docker stop",
			"context", ctxName,
			"app", run.Image,
			"id", run.Id,
			"node", run.Node,
			"name", run.Name,
			"upTime", run.UpTime,
			"shell", "docker stop "+run.Id,
			"usedTime", usedTime,
			"limitTime", 60000,
			"result", shellOut,
			"stopIsOk", stopIsOk,
		)
	}

	//log.Printf("Dock	exec	[%s]	%s	%s	\033[31mdocker rm %s\033[0m", ctxName, run.Image, run.Node, run.Id)
	shellOut, usedTime, _ = shellFunc(30000, run.Node, "rm", run.Id)
	rmIsOk := true
	if out := getLastLine(shellOut); out != run.Id {
		//log.Printf("Dock	exec	[%s]	rm	error	%s	!=	%s", ctxName, out, run.Id)
		rmIsOk = false
	}
	if err != nil {
		logError("docker rm failed: "+err.Error(),
			"context", ctxName,
			"app", run.Image,
			"id", run.Id,
			"node", run.Node,
			"name", run.Name,
			"upTime", run.UpTime,
			"shell", "docker rm "+run.Id,
			"usedTime", usedTime,
			"limitTime", 60000,
			"result", shellOut,
			"stopIsOk", rmIsOk,
		)
	} else {
		logInfo("docker rm",
			"context", ctxName,
			"app", run.Image,
			"id", run.Id,
			"node", run.Node,
			"name", run.Name,
			"upTime", run.UpTime,
			"shell", "docker rm "+run.Id,
			"usedTime", usedTime,
			"limitTime", 60000,
			"result", shellOut,
			"stopIsOk", rmIsOk,
		)
	}
	return stopIsOk, err
}

func getLastLine(out string) string {
	a := strings.Split(strings.TrimSpace(out), "\n")
	return a[len(a)-1]
}

func PraseCommandArgs(cmd string) []string {
	cmd = strings.TrimSpace(cmd) + " "
	args := make([]string, 0)
	start := -1
	var quota int32 = 0
	for i, c := range cmd {
		if start == -1 {
			start = i
			if c == '"' || c == '\'' {
				quota = c
			}
		} else if c == ' ' {
			if quota == 0 {
				if cmd[start] == cmd[i-1] && (cmd[start] == '"' || cmd[start] == '\'') {
					args = append(args, strings.ReplaceAll(cmd[start+1:i-1], fmt.Sprintf("\\%c", cmd[start]), fmt.Sprintf("%c", cmd[start])))
				} else {
					args = append(args, cmd[start:i])
				}
				start = -1
			}
		} else if c == quota {
			if i > 0 && cmd[i-1] != '\\' {
				quota = 0
			}
		}
	}
	return args
}
