package gate

import (
	"github.com/ssgo/hub/dock"
	"github.com/ssgo/log"
	"github.com/ssgo/redis"
	"github.com/ssgo/s"
	"github.com/ssgo/u"
)

type gateConfig struct {
	Proxies  map[string]string
	Rewrites map[string]string
	Prefix   string
}

var logger = log.New(u.ShortUniqueId())

func Registers() {
	s.SetAuthChecker(dock.Auth)
	s.Restful(1, "GET", "/gateway", getGatewayInfo)
	s.Restful(2, "POST", "/gateway", setGatewayInfo)
}

func getPrefix() string {
	redisPool := redis.GetRedis(dock.GetDiscover(), logger)
	proxyKeys := redisPool.KEYS("_*proxies")
	if len(proxyKeys) < 1 {
		return "_"
	}
	proxiesKey := proxyKeys[0]
	lenProxiesKey := len(proxiesKey)
	if lenProxiesKey <= 7 {
		return "_"
	}
	prefix := proxiesKey[0 : lenProxiesKey-7]
	return prefix
}

func getGatewayInfo() (gatewayConfig gateConfig) {
	redisPool := redis.GetRedis(dock.GetDiscover(), logger)
	prefix := getPrefix()
	gatewayConfig.Proxies = redisPool.Do("HGETALL", prefix+"proxies").StringMap()
	gatewayConfig.Rewrites = redisPool.Do("HGETALL", prefix+"rewrites").StringMap()
	gatewayConfig.Prefix = prefix
	return
}

func setGatewayInfo(gatewayConfig gateConfig) bool {
	prefix := getPrefix()
	newProxies := gatewayConfig.Proxies
	newRewrites := gatewayConfig.Rewrites
	redisPool := redis.GetRedis(dock.GetDiscover(), logger)
	return saveMulti(prefix, "proxies", newProxies, redisPool) && saveMulti(prefix, "rewrites", newRewrites, redisPool)
}

func saveMulti(prefix string, key string, newList map[string]string, redisPool *redis.Redis) bool {
	oldList := redisPool.Do("HGETALL", prefix+key).StringMap()
	currentKey := prefix + key
	for index, single := range newList {
		if !redisPool.HSET(currentKey, index, single) {
			return false
		}
	}
	for index, _ := range oldList {
		_, ok := newList[index]
		if !ok {
			redisPool.HDEL(currentKey, index)
		}
	}
	redisPool.Do("PUBLISH", "_CH"+currentKey, 1)
	return true
}
