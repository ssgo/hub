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
	s.SetAuthChecker(Auth)
	s.Static("/", "www/")
	s.Restful(0, "POST", "/login", login)

	s.Restful(1, "GET", "/global/status", getGlobalStatus)
	s.Restful(1, "GET", "/global", getGlobalInfo)
	s.Restful(1, "GET", "/contexts", getContextList)
	s.Restful(1, "GET", "/{name}", getContext)
	s.Restful(1, "GET", "/{name}/status", getContextRuns)

	s.Restful(3, "POST", "/global", setGlobalInfo)
	s.Restful(2, "POST", "/{name}", setContext)
	s.Restful(3, "DELETE", "/{name}", removeContext)

	s.Restful(9, "GET", "/install/{token}", getNodeInstaller)
}

func Auth(authLevel int, url *string, in map[string]interface{}, request *http.Request) bool {
	token := request.Header.Get("Access-Token")
	switch authLevel {
	case 1:
		return authManage(token) || authAnyContext(token)
	case 2:
		return authManage(token) || authContext(token, u.String(in["name"]))
	case 3:
		return authManage(token)
	case 9:
		return in["token"] == installToken
	}
	return false
}

func authManage(token string) bool {
	return token != "" && token == hubConfig.ManageToken
}

func authContext(token, contextName string) bool {
	if token == "" {
		return false
	}
	ctxs := ctxsSafely.Load().(map[string]*ContextInfo)
	for ctxName := range ctxListSafely.Load().(map[string]string) {
		if ctxName == contextName {
			ctx := ctxs[ctxName]
			return EncodeToken(ctx.Token) == token
		}
	}
	return false
}

func authAnyContext(token string) bool {
	if token == "" {
		return false
	}
	ctxs := ctxsSafely.Load().(map[string]*ContextInfo)
	for ctxName := range ctxListSafely.Load().(map[string]string) {
		ctx := ctxs[ctxName]
		if EncodeToken(ctx.Token) == token {
			return true
		}
	}
	return false
}

func login(request *http.Request) int {
	token := request.Header.Get("Access-Token")
	//if request.Header.Get("Access-Token") == hubConfig.AccessToken {
	//	return 1
	//}
	if authManage(token) {
		return 3
	}
	if authAnyContext(token) {
		return 2
	}
	return 0
}

type GlobalInfo struct {
	Nodes    map[string]*NodeInfo
	Vars     map[string]*string
	Args     string
	Registry SimpleRegistryInfo
}

func getGlobalInfo() (out struct {
	GlobalInfo
	PublicKey          string
	InstallToken       string
	RegistryRunCommand string
}) {
	out.Nodes = nodesSafely.Load().(map[string]*NodeInfo)
	out.Vars = globalVars
	out.Args = globalArgs
	out.Registry = globalRegistry
	out.PublicKey, _ = u.ReadFile(dataPath(".ssh", "id_ecdsa.pub"), 2048)
	out.InstallToken = installToken
	if globalRegistry.Domain != "" {
		portConfig := "80:5000"
		if strings.Contains(globalRegistry.Domain, ":") {
			portConfig = strings.Split(globalRegistry.Domain, ":")[1] + ":5000"
		}
		out.RegistryRunCommand = fmt.Sprintln("docker run --name registry -d --restart=always -p", portConfig, "-e REGISTRY_STORAGE_DELETE_ENABLED=true -v", globalRegistry.HubDataPath+"/registryAuth:/root/registryAuth -e REGISTRY_AUTH=htpasswd -e REGISTRY_AUTH_HTPASSWD_PATH=/root/registryAuth -e REGISTRY_AUTH_HTPASSWD_REALM=Registry -v", globalRegistry.DataPath + ":/var/lib/registry", globalRegistry.Image)
	}
	return
}

type globalStatusResult struct {
	Nodes          map[string]*NodeStatus
	RegistryStatus string
}

func getNodeInstaller() string {
	publicKey, _ := u.ReadFile(dataPath(".ssh", "id_ecdsa.pub"), 2048)
	publicKey = strings.TrimSpace(publicKey)
	registrySetting := ""
	if globalRegistry.Domain != "" {
		registrySetting = `echo "# setting docker ..."
echo '{"insecure-registries":["` + globalRegistry.Domain + `"]}' > /etc/docker/daemon.json
systemctl restart docker
`
	}

	return `
echo "# creating doker user ..."
useradd docker -g docker

echo "# installing limit-docker ..."
cat > /home/docker/limit-docker << EOF
cmdarr=(\$SSH_ORIGINAL_COMMAND)
cmd=\${cmdarr[0]}
if [ \$cmd != "docker" ];then
    echo "\$cmd is not allow"
    exit
fi
\$SSH_ORIGINAL_COMMAND
EOF

` + registrySetting + `
echo "# installing ssh key ..."
mkdir /home/docker/.ssh
echo 'command="/home/docker/limit-docker",no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty ` + publicKey + `' > /home/docker/.ssh/authorized_keys
chmod 500 /home/docker/.ssh
chmod 400 /home/docker/.ssh/authorized_keys
chown -R docker:docker /home/docker/.ssh
chmod +x /home/docker/limit-docker

echo "# done"
`
}

func getGlobalStatus() globalStatusResult {
	return globalStatusResult{
		Nodes:          nodeStatusSafely.Load().(map[string]*NodeStatus),
		//RegistryStatus: GetSimpleRegistryStatus(),
	}
}

func getContextList(request *http.Request) map[string]string {
	list := ctxListSafely.Load().(map[string]string)
	token := request.Header.Get("Access-Token")
	if authManage(token) {
		return list
	} else {
		ctxs := ctxsSafely.Load().(map[string]*ContextInfo)
		out := map[string]string{}
		for k, v := range list {
			ctx := ctxs[k]
			if EncodeToken(ctx.Token) == token {
				out[k] = v
			}
		}
		return out
	}
}

func GetDiscover() string {
	discover, ok := globalVars["discover"]
	if !ok {
		return ""
	}
	discoverVar := *discover

	discoverPos := strings.Index(discoverVar, "=")
	if discoverPos < 1 {
		return ""
	}
	lenDiscover := len(discoverVar)
	discoverVar = discoverVar[discoverPos+1 : lenDiscover]
	discoverVar = strings.Trim(discoverVar, " ")
	discoverVar = strings.Trim(discoverVar, "'")
	return strings.Trim(discoverVar, "\"")
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
		globalRegistry.Image = in.Registry.Image
		globalRegistry.Domain = in.Registry.Domain
		globalRegistry.DataPath = in.Registry.DataPath
		globalRegistry.HubDataPath = in.Registry.HubDataPath
		//globalRegistry.Start = in.Registry.Start
		in.Registry.User = globalRegistry.User
		in.Registry.Password = globalRegistry.Password
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

	if in.Token == "" {
		in.Token = u.ShortUniqueId()
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
