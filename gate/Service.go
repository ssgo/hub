package gate

import (
	"github.com/ssgo/hub/dock"
	"github.com/ssgo/log"
	"github.com/ssgo/redis"
	"github.com/ssgo/s"
	"github.com/ssgo/u"
	"strings"
)

var logger = log.New(u.ShortUniqueId())

func Registers() {
	s.SetAuthChecker(dock.Auth)
	s.Restful(1, "GET", "/gateway", getGateway)
	s.Restful(2, "POST", "/gateway", setGateway)
}

type GatewayInfo struct {
	Key   string
	Field string
	Value string
}

func getGateway() (out struct{ Configs []GatewayInfo }) {
	redisPool := redis.GetRedis(dock.GetDiscover(), logger)
	proxyKeys := redisPool.KEYS("_*proxies")
	proxyKeys = append(proxyKeys, redisPool.KEYS("_*rewrites")...)

	out.Configs = make([]GatewayInfo, 0)
	for _, key := range proxyKeys {
		list := redisPool.Do("HGETALL", key).StringMap()
		for field, value := range list {
			out.Configs = append(out.Configs, GatewayInfo{
				Key:   key,
				Field: field,
				Value: value,
			})
		}
	}
	return
}

func setGateway(in struct{ Configs []GatewayInfo }) bool {
	redisPool := redis.GetRedis(dock.GetDiscover(), logger)

	datas := make(map[string]map[string]string)
	dataChangeds := make(map[string]bool)
	for _, data := range in.Configs {
		if !strings.HasPrefix(data.Key, "_") || (!strings.HasSuffix(data.Key, "proxies") && !strings.HasSuffix(data.Key, "rewrites")) {
			continue
		}
		if datas[data.Key] == nil {
			datas[data.Key] = redisPool.Do("HGETALL", data.Key).StringMap()
		}
		if datas[data.Key][data.Field] != data.Value {
			redisPool.HSET(data.Key, data.Field, data.Value)
			dataChangeds[data.Key] = true
		}
		delete(datas[data.Key], data.Field)
	}

	for key, data := range datas {
		for field, _ := range data {
			redisPool.HDEL(key, field)
			dataChangeds[key] = true
		}
	}

	for k := range dataChangeds {
		redisPool.Do("PUBLISH", "_CH"+k, 1)
	}
	return true
}
