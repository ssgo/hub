# 基于 ssgo/s 的一个docker自动管理工具

docker run --name=dock -d --restart=always ssgo/dock:0.0.1

# 存储依赖

应用会依赖 service.json 中 registryCalls 指定的 redis 配置访问注册信息，默认值为 "discover:15"

同时会访问该 redis db 中的 proxies 的内容，进行动态配置，如果修改 proxies 的内容需要进行 INCR proxiesVersion 操作提升版本

## 配置

可在项目根目录放置一个 proxy.json

```json
{
  "checkInterval": 5,
  "proxies": {
    "localhost:8080": "mainapp",
    "127.0.0.1/status": "status",
    "/hello": "welcome",
    "forDev": ".*?/(.*?)(/.*)"
  }
}
```

checkInterval 同步 redis db 中 proxies 动态配置的间隔时间，单位 秒

proxies 中 支持
 - Host => App
 - Path => App
 - Host&Path => App
 - Host&Path 进行正则匹配后 $1 是 App $2 是 requestPath，key 没有实际意义

配置内容也可以同时使用 env.json 或环境变量设置（优先级高于配置文件）
