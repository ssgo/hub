package dock

import (
	redigo "github.com/garyburd/redigo/redis"
	"github.com/ssgo/redis"
	"log"
	"time"
	"os"
	"github.com/ssgo/base"
	"os/signal"
	"syscall"
	"strings"
)

var dcCache *redis.Redis

var config = struct {
	CheckInterval  int
	ReviewInterval int
	LogFile        string
	Nodes          map[string]*string
	Apps           map[string]*string
	Binds          map[string]*string
	Registry       string
	PrivateKey     string
}{}

var sleepUnit = time.Second
var isRunning = false
var isRefresh = true
var syncConn *redigo.PubSubConn

var syncerStartChan = make(chan bool)
var syncerStopChan = make(chan bool)
var pingStopChan = make(chan bool)
var startChan chan bool
var stopChan chan bool

func SetSleepUnit(unit time.Duration) {
	sleepUnit = unit
}

func initConfig() {
	base.LoadConfig("dock", &config)

	log.SetFlags(log.Ldate | log.Lmicroseconds)
	if config.LogFile != "" {
		f, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			log.SetOutput(f)
		} else {
			log.SetOutput(os.Stdout)
			log.Print("ERROR	", err)
		}
	} else {
		log.SetOutput(os.Stdout)
	}

	if config.CheckInterval < 3 {
		config.CheckInterval = 5
	}

	if config.ReviewInterval < 10 {
		config.ReviewInterval = 30
	}
	if config.Registry == "" {
		config.Registry = "dock:14"
	}
	if config.PrivateKey != "" {
		f, err := os.OpenFile("/opt/privateKey", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err == nil {
			f.Write([]byte(strings.Replace(config.PrivateKey, ",", "\n", 100)))
			f.Close()
		}
	}

	dcCache = redis.GetRedis(config.Registry)
	if shellFunc == nil {
		shellFunc = defaultShell
	}
}

func Start() {
	initConfig()
	closeChan := make(chan os.Signal, 2)
	signal.Notify(closeChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-closeChan
		log.Print("Dock	stopping ...")
		isRunning = false
		if syncConn != nil {
			syncConn.Unsubscribe("_refresh")
			syncConn.Close()
			syncConn = nil
		}
	}()

	isRunning = true
	go pingRedis()
	go syncNotice()
	<-syncerStartChan

	log.Print("Dock	started")
	if startChan != nil {
		startChan <- true
	}

	reviewInterval := 0
	for {
		// 更新节点
		nodeChanged := updateNodesInfo()
		// 更新应用，产生 startingApps stoppingApps
		appChanged := updateAppsInfo()

		// 变化了或者到了 config.ReviewInterval 执行一次 review
		if nodeChanged || appChanged || reviewInterval >= config.ReviewInterval {
			// 获取实时运行状态
			makeAppRunningInfo()

			// 启动startingApps，停止stoppingApps
			checkChanged := checkApps()

			if nodeChanged || appChanged || checkChanged {
				// 打印当前状态
				showStats()
			}
		}
		if reviewInterval >= config.ReviewInterval {
			reviewInterval = 0
		}

		if !isRunning {
			break
		}

		for i := 0; i < config.CheckInterval*2; i++ {
			time.Sleep(sleepUnit / 2)
			if !isRunning {
				break
			}
			if isRefresh {
				isRefresh = false
				break
			}
		}
		if !isRunning {
			break
		}

		reviewInterval += config.CheckInterval
	}
	log.Print("Dock	waitting for noitce syncer ...")
	<-syncerStopChan
	log.Print("Dock	waitting for noitce pinger ...")
	<-pingStopChan
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
	if syncConn != nil {
		syncConn.Unsubscribe("_refresh")
		syncConn.Close()
		syncConn = nil
	}
	<-stopChan
}

func syncNotice() {
	for {
		syncConn = &redigo.PubSubConn{Conn: dcCache.GetConnection()}
		err := syncConn.Subscribe("_refresh")
		syncerStartChan <- true
		if err != nil {
			log.Print("REDIS SUBSCRIBE	", err)
			syncConn.Close()
			syncConn = nil

			time.Sleep(sleepUnit)
			if !isRunning {
				break
			}
			continue
		}

		// 开始接收订阅数据
		for {
			isErr := false
			switch v := syncConn.Receive().(type) {
			case redigo.Message:
				isRefresh = true
			case error:
				if !strings.Contains(v.Error(), "connection closed") {
					log.Printf("REDIS RECEIVE ERROR	%s", v)
				}
				isErr = true
				break
			}
			if isErr {
				break
			}
		}
		if !isRunning {
			break
		}
		time.Sleep(sleepUnit)
		if !isRunning {
			break
		}
	}

	if syncConn != nil {
		syncConn.Unsubscribe("_refresh")
		syncConn.Close()
		syncConn = nil
	}
	syncerStopChan <- true
}

// 保持 redis 链接，否则会因为超时而发生错误
func pingRedis() {
	n := 15
	if dcCache.ReadTimeout > 2000 {
		n = dcCache.ReadTimeout / 1000 / 2
	} else if dcCache.ReadTimeout > 0 {
		n = 1
	}
	for {
		for i := 0; i < n; i++ {
			time.Sleep(sleepUnit)
			if !isRunning {
				break
			}
		}
		if !isRunning {
			break
		}
		if syncConn != nil {
			syncConn.Ping("1")
		}
		if !isRunning {
			break
		}
	}
	pingStopChan <- true
}
