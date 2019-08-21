package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/darkMoon1973/PortScanner/common/lib/goworker"
	"github.com/darkMoon1973/PortScanner/common/util"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

var logic *Logic

type Logic struct {
	localIp     string
	Log         *logrus.Entry
	Conf        *config
	RedisClient *redis.Client
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logic = newLogic()
	logic.Log, logic.Conf = flagParse()
	agentRegister()
	logic.newRedis()
	logic.newGoworker()
	logic.getInterface()
}

func newLogic() *Logic {
	localIp, err := util.GetLocalIp()
	fmt.Println("local ip is ", localIp)
	if err != nil && localIp == "" {
		fmt.Println("Get local ip fail, ", err)
		os.Exit(1)
	}
	return &Logic{localIp: localIp}
}

func (l *Logic) newRedis() {
	l.RedisClient = newRedis(l.Conf.RedisUrl)
	l.Log.Debug("init redis success")
}

func (l *Logic) newGoworker() {
	l.Conf.workerSetting.URI = l.Conf.RedisUrl
	goworker.SetSettings(l.Conf.workerSetting)
	l.Log.Debug("init goworker success")
}

// 手动获取 Masscan 发包网卡
func (l *Logic) getInterface() {
	if l.Conf.Masscan.autoInterface == false {
		var err error
		// 获取外网网卡和路由地址
		_, l.Conf.Masscan.publicInterface, l.Conf.Masscan.publicRouteAddr, err = util.GetNetInfo("114.114.114.114")
		// 获取内网网卡和路由地址
		_, l.Conf.Masscan.privateInterface, l.Conf.Masscan.privateRouteAddr, err = util.GetNetInfo("10.1.1.1")
		if err != nil {
			l.Log.Fatal("Get interface name and route addr fail, check your network\n", err)
			os.Exit(1)
		}
	}
}

// goworker workfunc
func masscanTask(queue string, args ...interface{}) error {
	logic.Log.Debug("Start masscan task from queue, ", queue)
	ip := args[0].(string)
	rate := args[1].(string)
	port := args[2].(string)
	isPublic := args[3].(bool)
	results, err := masScanner(ip, port, rate, isPublic)
	if err != nil {
		logic.Log.Error("Masscan scan fail, ", err)
	}
	// masscan 结果 push 到 nmap 队列
	err = resultToNmapQueue(results, "Nmap", "Nmap")
	return nil
}

// goworker workfunc
func nmapTask(queue string, args ...interface{}) error {
	logic.Log.Debug("Start nmap task from queue, ", queue)
	var nmapResults []interface{}
	logic.Log.Debug("start Nmap Scanner, ", args[0], args[1])
	results, err := nmapScanner(args[0].(string), args[1].(string), args[2].(bool))
	logic.Log.Debug("Nmap Scanner result is: ", results)
	if err != nil {
		logic.Log.Error("Start nmap scanner fail, ", err)
		return err
	}
	for _, v := range results {
		r, _ := json.Marshal(v)
		nmapResults = append(nmapResults, r)
	}
	logic.RedisClient.RPush("nmapResults", nmapResults...)
	return nil
}

func Run() {
	goworker.Register("Masscan", masscanTask)
	goworker.Register("Nmap", nmapTask)

	go func() {
		heartBeat()
	}()
	err := goworker.Work()

	if err != nil {
		logic.Log.Error("Start goworker work fail, ", err)
	}
}
