package dock

import (
	"log"
	"time"
	"os"
	"github.com/ssgo/base"
	"os/signal"
	"syscall"
	"strings"
	"io/ioutil"
	"sync"
	"crypto/sha1"
	"encoding/hex"
)

var config = struct {
	CheckInterval int
	DataPath      string
	//LogFile       string
	AccessToken   string
	ManageToken   string
	PrivateKey    string
}{}

var sleepUnit = time.Second
var isRunning = false
//var isMaking = false
var makingLocker sync.Mutex

var startChan chan bool
var stopChan chan bool

func SetSleepUnit(unit time.Duration) {
	sleepUnit = unit
}

func initConfig() {
	base.LoadConfig("dock", &config)

	//log.SetFlags(log.Ldate | log.Lmicroseconds)
	//if config.LogFile != "" {
	//	f, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	//	if err == nil {
	//		log.SetOutput(f)
	//	} else {
	//		log.SetOutput(os.Stdout)
	//		log.Print("ERROR	", err)
	//	}
	//} else {
	//	log.SetOutput(os.Stdout)
	//}

	if config.CheckInterval == 0 {
		config.CheckInterval = 5
	}
	if config.DataPath == "" {
		config.DataPath = "/opt/data"
	}
	if config.PrivateKey != "" {
		f, err := os.OpenFile("/opt/privateKey", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err == nil {
			f.Write([]byte(strings.Replace(config.PrivateKey, ",", "\n", 100)))
			f.Close()
		}
	}

	if config.AccessToken == "" {
		config.AccessToken = "51dock"
	}
	if config.ManageToken == "" {
		config.ManageToken = "91dock"
	}

	sha1Maker := sha1.New()
	sha1Maker.Write([]byte("SSGO-"))
	sha1Maker.Write([]byte(config.AccessToken))
	sha1Maker.Write([]byte("-Dock"))
	config.AccessToken = hex.EncodeToString(sha1Maker.Sum([]byte{}))
	sha1Maker.Reset()
	sha1Maker.Write([]byte("SSGO-"))
	sha1Maker.Write([]byte(config.ManageToken))
	sha1Maker.Write([]byte("-Dock"))
	config.ManageToken = hex.EncodeToString(sha1Maker.Sum([]byte{}))

	if shellFunc == nil {
		shellFunc = defaultShell
	}
}

func Start() {
	closeChan := make(chan os.Signal, 2)
	signal.Notify(closeChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-closeChan
		log.Print("Dock	stopping ...")
		isRunning = false
	}()

	initConfig()

	load("nodes", &nodes)
	nodeStatus = make(map[string]*NodeStatus)

	files, err := ioutil.ReadDir(config.DataPath)
	if err == nil {
		for _, file := range files {
			fileName := file.Name()
			if fileName[0] == '.' || fileName == "nodes" || file.IsDir() {
				continue
			}
			ctx := newContext()
			load(fileName, &ctx)
			log.Println("Dock	loding	context	", fileName)
			if ctx.Name == fileName {
				ctxList[ctx.Name] = ctx.Desc
				ctxs[ctx.Name] = ctx
				ctxRuns[ctx.Name] = make(map[string][]*AppStatus)
				stoppingCtxApps[ctx.Name] = make(map[string]*AppInfo)
			}
		}
	}
	makeAppRunningInfos(true)
	for ctxName := range ctxs {
		checkContext(ctxName)
	}

	showStats()

	isRunning = true

	log.Print("Dock	started")
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
		makeAppRunningInfos(true)
		if !isRunning {
			break
		}
		changed := false
		for ctxName := range ctxs {
			checkChanged, _ := checkContext(ctxName)
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
						if checkAppForStoppingNodes(ctxName, appName) {
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

		makingLocker.Unlock()

		if !isRunning {
			break
		}
		for i := 0; i < config.CheckInterval; i++ {
			time.Sleep(sleepUnit)
			if !isRunning {
				break
			}
		}
	}
	if stopChan != nil {
		stopChan <- true
	}
	log.Print("Dock	stopped")
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
