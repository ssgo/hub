# 基于 ssgo/s 的一个docker自动管理工具

docker run -d --restart=always --name dock --network=host -v /opt/dock:/opt/data -e 'dock_privateKey=-----BEGIN RSA PRIVATE KEY-----,......,-----END RSA PRIVATE KEY-----' ssgo/dock:0.2

# 存储依赖

数据会存储在 /opt/data 下，可以使用 -v /opt/dock:/opt/data 来挂在外部磁盘

# SSH Key

dock服务使用 ssh 对节点进行管理，使用 ssh-keygen 创建一堆密钥，私钥用逗号替换换行配置在环境变量中

-e 'dock_privateKey=-----BEGIN RSA PRIVATE KEY-----,......,-----END RSA PRIVATE KEY-----'

公钥配置在初始化好docker环境的节点的docker账号中

# 管理界面

http://xx.xx.xx.xx:8888/

默认使用 8888 端口，可以使用 -p xxxx:8888 来改变端口

使用 -e dock_accessToken=51dock 和 -e dock_managerToken=91dock 配置查看和管理两个口令进行登录授权

使用 -e service_xxxx 来配置 http 相关参数，例如可以配置为基于 https 访问，具体配置请参考 https://github.com/ssgo/s

## 基本配置

可在项目根目录放置一个 proxy.json

```json
{
  "CheckInterval": 5,
  "dataPath": "",
  "accessToken": "51dock",
  "manageToken": "91dock",
  "privateKey": "-----BEGIN RSA PRIVATE KEY-----,....,....,....,-----END RSA PRIVATE KEY-----"
}
```

checkInterval 检查应用状态的间隔时间，单位 秒

dataPath 存储数据的路径，默认为 /opt/data

accessToken 以只读方式查看节点和应用的运行状态

manageToken 以只读方式查看节点和应用的运行状态

privateKey 是通过 ssh 访问节点的私钥

可以使用 -e 'dock_xxxxxx=xxxx' 进行配置

