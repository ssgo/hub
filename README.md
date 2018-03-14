# 基于 ssgo/s 的一个docker自动管理工具

docker run -d ssgo/dock

# 存储依赖

应用会依赖 dock.json 中 registry 指定的 redis 配置访问注册信息，默认值为 "dock:14"

根据 redis db 中的 _nodes, _apps, _binds 的内容，进行自动部署

## 基本配置

可在项目根目录放置一个 proxy.json

```json
{
  "CheckInterval": 5,
  "logFile": "",
  "registry": "dock:14",
  "accessToken": "",
  "managerToken": "",
  "nodes": {},
  "apps": {},
  "binds": {},
  "privateKey": "-----BEGIN RSA PRIVATE KEY-----,....,....,....,-----END RSA PRIVATE KEY-----"
}```

checkInterval 检查 redis db 中数据变化的间隔时间，单位 秒
如果修改了配置也可以使用 publish _refresh 1 进行立刻刷新

accessToken 用来调用 /status 是需要在 Header 中传递 Access-Token 以获得访问权限

nodes、apps、binds 作为初始配置，可部署出基本服务，例如：redis

privateKey 是通过 ssh 访问 nodes 的私钥

配置内容也可以同时使用 env.json 或环境变量设置（优先级高于配置文件）


## 应用配置

### nodes

```shell
hset _nodes 10.1.1.3 20,120
hset _nodes 10.1.1.4:22 8,32
```

key 为节点的IP地址+ssh端口号

value 为 CPUs,Memorys CPU和内存支持浮点数，内存的单位为 G


### apps

```shell
hset _apps mysql/mysql-server:5.7 "4,32,1,1,-p 3306:3306 -v /opt/db:/var/lib/mysql"
hset _apps redis#2 "1,4,1,1,-p 6000:6379 -v /opt/redis2:/data <--requirepass xx2xx>"
```

key 镜像名称，可在镜像后面用 # 增加一个自定义 tag，没有实际意义仅用于区分不同的应用配置

value 为 CPUs,Memorys,Min,Max,Args

CPU和内存支持浮点数，内存的单位为 G

Min和Max 分别为最小实例数和最大实例数

Args 是 docker run 后面的启动参数，系统为自动配置 -d --restart=always，并且自动维护 --name，不要重复设置

Args 最后如果跟着 <....> 表示使用启动参数，例如配置redis服务的密码

