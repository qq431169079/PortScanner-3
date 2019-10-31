# PortScanner

一款基于 golang 的分布式高并发端口扫描器，

### 安装部署

---

```bash
go get github.com/darkMoon1973/PortScanner

go build scanServer.go

go build scanAgent.go
```

##### 配置



```toml
app_name = "PortScanner"
# Redis Server URI
redis_url = "redis://root:password@redis-server:6379/0"
# Redis 队列长度
worker_num = 500

# 从文件读取扫描IP和IP白名单
[ip_list_file]
    enable = true
    scan_list = "scan_ip.txt"
    white_list = "white_ip.txt"
    
# HTTP 服务所需 Token 和开启端口
[http_config]
    http_token = "PortScanner"
    http_port = 1973
    
# Mongodb 相关配置
[mongodb_config]
    mongodb_url = "mongodb://127.0.0.1:27017"
    mongodb_dbname = "PortScanner"
    
# 日志文件和日志级别配置
[log_config]
    log_level = "error"
    log_file = "server.log"

# masscan 相关配置
# 如果 auto_interface 为 true，masscan 会自动获取网卡名称和路由地址.
# 为 false 将使用系统命令 [ip route get x.x.x.x] 获取 masscan 发包网卡和路由IP地址
# 具体 issues 见 https://github.com/robertdavidgraham/masscan/issues/220
[masscan_config]
    auto_interface = true
    port_range = "1-65535"
    scan_rate = "3000"

# 邮件提醒相关配置，enable 为是否开启邮件功能
[email_config]
    enable = false
    username = "username@163.com"
    password = "password"
    smtp_host = "smtp.163.com"
    smtp_port = "25"
    to_list = ["username@163.com"]
```



##### Server 端启动


```sh
nohup scanServer 1>nohup.out 2>stderr.out &
```

##### Agent 端启动

利用 curl 从 server 开启的 http 端口下载 agent 客户端，默认日志级别为 debug，worker 数量为10，日志文件名称为 `agent.log`

```shell
nohup curl http://server-host:1973/api/download --user "PortScanner:PortScanner" -o scanAgent && chmod +x scanAgent && ./scanAgent -server server-host:1973 -token PortScanner  -loglevel info -logfile agent.log -worker 10 2> stderr.out 1> nohup.out &
```

默认扫描调度只实现了读取扫描名单，过滤白名单的功能，需要更为复杂扫描逻辑的请自行修改代码添加。


### 参考资料

---

[goSkylar](https://github.com/LakeVilladom/goSkylar) 

