package dock

import (
	"log"
	"strings"
	"fmt"
	"regexp"
)

var nodeFailedTimes = map[string]int{}
var dockerNameFilter *regexp.Regexp

func getRunningApps(nodeName string) []*AppRunInfo {
	runs := make([]*AppRunInfo, 0)
	out := shellFunc(nodeName, "ps", "--format", "'{{.ID}}, {{.Names}}, {{.Image}}, {{.Status}}'")
	for _, line := range strings.Split(out, "\n") {
		a1 := strings.Split(line, ", ")
		if len(a1) >= 4 {
			run := &AppRunInfo{Node: nodeName, Id: a1[0], Name: a1[1], Image: a1[2], UpTime: a1[3]}
			// 通过 --name 中隐藏的 tag 信息补全 Image
			a2 := strings.Split(run.Name, "-")
			if len(a2) >= 3 {
				run.Image += "#" + a2[1]
			}
			runs = append(runs, run)
		}
	}
	return runs
}

func startApp(appName, nodeName string, app *AppInfo) string {
	if dockerNameFilter == nil {
		dockerNameFilter, _ = regexp.Compile("[^a-zA-Z0-9]")
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
	dockerRunIndex := dcCache.HINCR("_appIndexes", dockerRunName)
	dockerRunName = fmt.Sprintf("%s-%d", dockerRunName, dockerRunIndex)

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

	// 解析启动参数
	runCmd := ""
	if strings.HasSuffix(app.Args, ">") {
		pos := strings.LastIndex(app.Args, " <")
		if pos != -1 {
			runCmd = app.Args[pos+2:len(app.Args)-1]
			app.Args = app.Args[0:pos]
		}
	}

	args = append(args, strings.Split(app.Args, " ")...)
	args = append(args, appName)
	if runCmd != "" {
		args = append(args, strings.Split(runCmd, " ")...)
	}
	log.Print("Dock	exec	run	", "docker ", strings.Join(args, " "))
	id := getLastLine(shellFunc(nodeName, args...))
	if len(id) > 12 {
		id = id[0:12]
	}
	log.Print("Dock	exec	run	result	", id)
	return id
}

func stopApp(run *AppRunInfo, app *AppInfo) bool {
	if run.Id == "" {
		return true
	}

	ok := true
	log.Println("Dock	exec	stop	", "docker", run.Node, "stop", run.Id)
	if out := getLastLine(shellFunc(run.Node, "stop", run.Id)); out != run.Id {
		log.Printf("Dock	exec	stop	error	%s	!=	%s", out, run.Id)
		ok = false
	}

	log.Println("Dock	exec	rm	", "docker", run.Node, "rm", run.Id)
	if out := getLastLine(shellFunc(run.Node, "rm", run.Id)); out != run.Id {
		log.Printf("Dock	exec	rm	error	%s	!=	%s", out, run.Id)
		ok = false
	}
	return ok
}

func getLastLine(out string) string {
	a := strings.Split(strings.TrimSpace(out), "\n")
	return a[len(a)-1]
}
