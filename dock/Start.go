package dock

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/ssgo/config"
	"github.com/ssgo/log"
	"github.com/ssgo/u"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var hubConfig = struct {
	CheckInterval int
	DataPath      string
	//LogFile       string
	//AccessToken string
	ManageToken   string
	DockerUser    string
	DockerCommand string
	//PrivateKey  string
}{}

var sleepUnit = time.Second
var isRunning = false

//var isMaking = false
var makingLocker sync.Mutex

var startChan chan bool
var stopChan chan bool

var logger = log.New(u.ShortUniqueId())

//var installToken = u.ShortUniqueId()
var installToken = "eqWTGOckcbi"

func logInfo(info string, extra ...interface{}) {
	logger.Info("Dock: "+info, extra...)
}

func logError(error string, extra ...interface{}) {
	logger.Error("Dock: "+error, extra...)
}

func SetSleepUnit(unit time.Duration) {
	sleepUnit = unit
}

func initConfig() {
	config.LoadConfig("hub", &hubConfig)

	//log.SetFlags(log.Ldate | log.Lmicroseconds)
	//if hubConfig.LogFile != "" {
	//	f, err := os.OpenFile(hubConfig.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	//	if err == nil {
	//		log.SetOutput(f)
	//	} else {
	//		log.SetOutput(os.Stdout)
	//		log.Print("ERROR	", err)
	//	}
	//} else {
	//	log.SetOutput(os.Stdout)
	//}

	if hubConfig.CheckInterval == 0 {
		hubConfig.CheckInterval = 5
	}
	if hubConfig.DataPath == "" {
		hubConfig.DataPath = "/opt/data"
	}
	if hubConfig.DockerUser == "" {
		hubConfig.DockerUser = "docker"
	}
	if hubConfig.DockerCommand == "" {
		hubConfig.DockerCommand = "docker"
	}
	//if hubConfig.PrivateKey != "" {
	//	f, err := os.OpenFile("/opt/privateKey", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	//	if err == nil {
	//		f.Write([]byte(strings.Replace(hubConfig.PrivateKey, ",", "\n", 100)))
	//		f.Close()
	//	}
	//}

	pubKeyFile := dataPath(".ssh", "id_ecdsa.pub")
	if !u.FileExists(pubKeyFile) {
		priKeyFile := dataPath(".ssh", "id_ecdsa")
		u.CheckPath(priKeyFile)
		_, err := u.RunCommand("ssh-keygen", "-f", priKeyFile, "-t", "ecdsa", "-N", "", "-C", "ssgo/dock")
		if err != nil {
			logError(err.Error())
		}
	}

	//if hubConfig.AccessToken == "" {
	//	hubConfig.AccessToken = "51hub"
	//}
	if hubConfig.ManageToken == "" {
		hubConfig.ManageToken = "91hub"
	}

	hubConfig.ManageToken = EncodeToken(hubConfig.ManageToken)

	if shellFunc == nil {
		shellFunc = defaultShell
	}
}

func EncodeToken(token string) string {
	sha1Maker := sha1.New()
	sha1Maker.Write([]byte("SSGO-"))
	sha1Maker.Write([]byte(token))
	sha1Maker.Write([]byte("-Dock"))
	return hex.EncodeToString(sha1Maker.Sum([]byte{}))
}

func Start() {
	closeChan := make(chan os.Signal, 2)
	signal.Notify(closeChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-closeChan
		logInfo("stopping")
		isRunning = false
	}()

	initConfig()

	global := GlobalInfo{}
	load("global", &global)
	if global.Registry.User == "" {
		global.Registry.Image = "registry"
		global.Registry.DataPath = "/opt/registry"
		global.Registry.HubDataPath = "/opt/hub"
		global.Registry.User = "hub"
		global.Registry.Password = u.ShortUniqueId()
		globalRegistry.Image = global.Registry.Image
		globalRegistry.DataPath = global.Registry.DataPath
		globalRegistry.HubDataPath = global.Registry.HubDataPath
		globalRegistry.User = global.Registry.User
		globalRegistry.Password = global.Registry.Password
		save("global", global)

		// 生成 htpasswd
		passwordBytes, _ := bcrypt.GenerateFromPassword([]byte(global.Registry.Password), bcrypt.DefaultCost)
		u.WriteFile(dataPath("registryAuth"), global.Registry.User+":"+string(passwordBytes)+"\n")
		logger.Info("create auth file for simple registry", "file", dataPath("registryAuth"))
	}
	if len(global.Nodes) == 0 && len(global.Vars) == 0 && global.Args == "" {
		// 兼容之前的 nodes 存储
		load("nodes", &global.Nodes)
	}
	nodes = global.Nodes
	globalVars = global.Vars
	globalArgs = global.Args
	globalRegistry = global.Registry
	nodeStatus = make(map[string]*NodeStatus)

	files, err := ioutil.ReadDir(hubConfig.DataPath)
	if err == nil {
		for _, file := range files {
			fileName := file.Name()
			if fileName[0] == '.' || fileName == "global" || fileName == "nodes" || file.IsDir() {
				continue
			}
			ctx := newContext()
			load(fileName, &ctx)
			logInfo("load context",
				"file", fileName,
				"context", ctx,
			)
			//fmt.Println("Dock	loading	context	", fileName)
			if ctx.Name == fileName {
				ctxList[ctx.Name] = ctx.Desc
				ctxs[ctx.Name] = ctx
				ctxRuns[ctx.Name] = make(map[string][]*AppStatus)
				stoppingCtxApps[ctx.Name] = make(map[string]*AppInfo)
			}
		}
	}
	if makeAppRunningInfos(true) {
		for ctxName := range ctxs {
			_, _, err = checkContext(ctxName)
			if err != nil {
				logError(err.Error())
			}
		}
	}

	showStats()

	isRunning = true

	logInfo("started")
	//log.Print("Dock	started")
	if startChan != nil {
		startChan <- true
	}

	nodesSafely.Store(nodes)
	nodeStatusSafely.Store(nodeStatus)
	ctxListSafely.Store(ctxList)
	ctxsSafely.Store(ctxs)
	ctxRunsSafely.Store(ctxRuns)

	// 开始轮询处理
	for {
		makingLocker.Lock()

		// 获取实时运行状态
		if makeAppRunningInfos(true) {
			changed := false
			for ctxName := range ctxs {
				checkChanged, _, _ := checkContext(ctxName)
				if checkChanged {
					changed = true
				}
				if !isRunning {
					break
				}
			}

			// 停掉不存在的节点上的实例
			if len(stoppingNodes) > 0 {
				makeAppRunningInfos(false)
				for ctxName := range ctxs {
					runsByApp := ctxRuns[ctxName]
					if runsByApp != nil {
						for appName := range runsByApp {
							if ok, _ := checkAppForStoppingNodes(ctxName, appName); ok {
								changed = true
							}
						}
					}
					if !isRunning {
						break
					}
				}

				for nodeName := range stoppingNodes {
					//log.Println("	aaaaaaaaa	", nodeName)
					if stoppingNodeStatus[nodeName] == nil || stoppingNodeStatus[nodeName].TotalRuns == 0 {
						//log.Println("	aaaaaaaaa	clear ", nodeName)
						delete(stoppingNodes, nodeName)
						delete(stoppingNodeStatus, nodeName)
					}
				}
			}

			nodeStatusSafely.Store(nodeStatus)
			ctxRunsSafely.Store(ctxRuns)
			if changed {
				nodesSafely.Store(nodes)
				ctxListSafely.Store(ctxList)
				ctxsSafely.Store(ctxs)
				showStats()
			}
		}

		makingLocker.Unlock()
		if !isRunning {
			break
		}

		// 检查 registry
		//CheckSimpleRegistry()

		if !isRunning {
			break
		}
		for i := 0; i < hubConfig.CheckInterval; i++ {
			time.Sleep(sleepUnit)
			if !isRunning {
				break
			}
		}
	}
	if stopChan != nil {
		stopChan <- true
	}
	logInfo("stopped")
	//log.Print("Dock	stopped")
}

func AsyncStart() {
	startChan = make(chan bool)
	go Start()
	<-startChan
}

func AsyncStop() {
	stopChan = make(chan bool)
	isRunning = false
	<-stopChan
}
