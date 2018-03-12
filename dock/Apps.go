package dock

import (
	"fmt"
	"strings"
	"strconv"
	"log"
)

const RUNNING = 1
const STOPPING = 2

type AppRunInfo struct {
	Node   string
	Id     string
	Image  string
	UpTime string
}
type AppInfo struct {
	Cpu    float32
	Memory float32
	Min    int
	Max    int
	Args   string
	Status int
	Runs   []*AppRunInfo
	Binds  []string
}

func (app *AppInfo) String() string {
	return fmt.Sprintf("%.2f,%.2f,%d,%d,%s", app.Cpu, app.Memory, app.Min, app.Max, app.Args)
}

var apps = map[string]*AppInfo{}

func updateAppsInfo() bool {
	changed := false

	// 强制使用配置中的应用
	remoteApps := map[string]string{}
	for appName, appString := range config.Apps {
		remoteApps[appName] = *appString
	}

	// 检查是否有需要启动的应用
	dcCache.Do("HGETALL", "_apps").To(&remoteApps)

	for appName, appString := range remoteApps {
		if apps[appName] != nil {
			if apps[appName].Status == STOPPING {
				apps[appName].Status = RUNNING
			}
			continue
		}

		a := strings.SplitN(appString, ",", 5)
		if len(a) < 3 {
			continue
		}
		cpu, err1 := strconv.ParseFloat(a[0], 10)
		memory, err2 := strconv.ParseFloat(a[1], 10)
		min, err3 := strconv.Atoi(a[2])
		max, err4 := strconv.Atoi(a[3])
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			continue
		}
		args := a[4]

		log.Printf("Dock	apps	add	%s	%s", appName, appString)
		changed = true
		apps[appName] = &AppInfo{Status: RUNNING, Cpu: float32(cpu), Memory: float32(memory), Min: min, Max: max, Args: args, Runs: make([]*AppRunInfo, 0)}
	}

	// 检查是否有需要删除的应用
	for appName, app := range apps {
		if remoteApps[appName] == "" {
			log.Printf("Dock	apps	remove	%s	%s", appName, app.String())
			changed = true
			app.Status = STOPPING
		}

		// 重置绑定信息
		app.Binds = make([]string, 0)
	}

	// 强制使用配置中的绑定
	remoteBinds := map[string]string{}
	for appName, nodesString := range config.Binds {
		remoteBinds[appName] = *nodesString
	}

	// 检查是否有需要启动的应用
	dcCache.Do("HGETALL", "_binds").To(&remoteBinds)

	for appName, nodesString := range remoteBinds {
		app := apps[appName]
		if app == nil {
			continue
		}
		app.Binds = strings.Split(nodesString, ",")
	}

	return changed
}
