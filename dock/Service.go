package dock

import (
	"github.com/ssgo/s"
	"net/http"
)

func Registers() {
	s.SetAuthChecker(auth)
	s.Register(1, "/status", GetStats)
}

func auth(authLevel uint, url *string, in *map[string]interface{}, request *http.Request) bool {
	switch authLevel {
	case 1:
		return request.Header.Get("Access-Token") == config.AccessToken
	case 2:
		return request.Header.Get("Manager-Token") == config.AccessToken
	}
	return false
}
