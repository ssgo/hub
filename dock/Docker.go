package dock

import (
	"log"
	"strings"
	"fmt"
)

var nodeFailedTimes = map[string]int{}

func getRunningApps(nodeName string) []*AppRunInfo {
	runs := make([]*AppRunInfo, 0)
	out := shellFunc(nodeName, "ps", "--format", "'{{.ID}}, {{.Image}}, {{.Status}}'")
	for _, line := range strings.Split(out, "\n") {
		a := strings.Split(line, ", ")
		if len(a) >= 3 {
			runs = append(runs, &AppRunInfo{Node: nodeName, Id: a[0], Image: a[1], UpTime: a[2]})
		}
	}
	return runs
}

func startApp(appName, nodeName string, app *AppInfo) string {
	args := make([]string, 0)
	args = append(args, "run", "-d", "--restart=always")
	if app.Cpu > 0.01 {
		args = append(args, "--cpus", fmt.Sprintf("%.2f", app.Cpu))
	}
	if app.Memory > 4 {
		args = append(args, "-m", fmt.Sprintf("%.0fg", app.Memory))
	} else if app.Memory > 0.01 {
		args = append(args, "-m", fmt.Sprintf("%.0fm", app.Memory*1024))
	}
	args = append(args, strings.Split(app.Args, " ")...)
	args = append(args, appName)
	log.Print("Dock	exec	run	", "docker ", strings.Join(args, " "))
	id := getLastLine(shellFunc(nodeName, args...))
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
