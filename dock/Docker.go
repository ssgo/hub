package dock

import (
	"log"
	"strings"
	"fmt"
	"regexp"
)

var nodeFailedTimes = map[string]int{}
var dockerNameFilter *regexp.Regexp
var dockerVarReplacer *regexp.Regexp

func getRunningApps(nodeName string) []*AppStatus {
	runs := make([]*AppStatus, 0)
	out := shellFunc(nodeName, "ps", "--format", "'{{.ID}}, {{.Names}}, {{.Image}}, {{.Status}}'")
	//log.Println("	******	", out)
	for _, line := range strings.Split(out, "\n") {
		a1 := strings.Split(line, ", ")
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
	return runs
}

func startApp(ctxName, appName, nodeName string, app *AppInfo) (string, string) {
	ctx := ctxs[ctxName]
	if ctx == nil {
		return "", ""
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
		if s != nil {
			return *s
		}
		return ""
	})

	// 解析启动参数
	runCmd := ""
	if strings.HasSuffix(tmpArgs, ">") {
		pos := strings.LastIndex(tmpArgs, " <")
		if pos != -1 {
			runCmd = tmpArgs[pos+2:len(tmpArgs)-1]
			tmpArgs = tmpArgs[0:pos]
		}
	}

	args = append(args, strings.Split(tmpArgs, " ")...)
	args = append(args, appName)
	if runCmd != "" {
		args = append(args, strings.Split(runCmd, " ")...)
	}
	log.Print("Dock	exec	run	[", ctxName, "]	\033[32mdocker ", strings.Join(args, " "), "\033[0m")
	id := getLastLine(shellFunc(nodeName, args...))
	if len(id) > 12 {
		id = id[0:12]
	}
	log.Print("Dock	exec	run	[", ctxName, "]	result	", id)
	return id, dockerRunName
}

func stopApp(ctxName string, run *AppStatus) bool {
	if run.Id == "" {
		return true
	}

	ok := true
	log.Printf("Dock	exec	[%s]	%s	%s	\033[31mdocker stop %s\033[0m", ctxName, run.Image, run.Node, run.Id)
	if out := getLastLine(shellFunc(run.Node, "stop", run.Id)); out != run.Id {
		log.Printf("Dock	exec	stop	[%s]	error	%s	!=	%s", ctxName, out, run.Id)
		ok = false
	}

	log.Printf("Dock	exec	[%s]	%s	%s	\033[31mdocker rm %s\033[0m", ctxName, run.Image, run.Node, run.Id)
	if out := getLastLine(shellFunc(run.Node, "rm", run.Id)); out != run.Id {
		log.Printf("Dock	exec	[%s]	rm	error	%s	!=	%s", ctxName, out, run.Id)
	}
	return ok
}

func getLastLine(out string) string {
	a := strings.Split(strings.TrimSpace(out), "\n")
	return a[len(a)-1]
}
