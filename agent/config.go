package agent

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/darkMoon1973/PortScanner/common/lib/goworker"
	"github.com/darkMoon1973/PortScanner/common/lib/logs"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type (
	config struct {
		RedisUrl      string
		Masscan       masscanConf
		Http          httpConf
		workerSetting goworker.WorkerSettings
	}
	httpConf struct {
		ServerHost string
		Token      string
	}
	masscanConf struct {
		autoInterface    bool
		privateInterface string
		privateRouteAddr string
		publicInterface  string
		publicRouteAddr  string
	}
)

func flagParse() (*logrus.Entry, *config) {
	conf := new(config)
	var logLevel, logFile string
	conf.workerSetting = workerSetting()
	//flag.StringVar(&config.agentType, "type", "", "agent type, Public or Private")
	flag.StringVar(&conf.Http.Token, "token", "", "auth token")
	flag.StringVar(&conf.Http.ServerHost, "server", "127.0.0.1:1967", "server host, ip:port")
	flag.IntVar(&conf.workerSetting.Concurrency, "worker", 10, "worker num")
	flag.StringVar(&logLevel, "loglevel", "debug", "log level")
	flag.StringVar(&logFile, "logfile", "agent.log", "log file")
	flag.Parse()
	log := logs.GetLogger(logLevel, logFile)
	return log, conf
}

// goworker 配置
func workerSetting() goworker.WorkerSettings {
	return goworker.WorkerSettings{
		Connections:    10,
		UseNumber:      true,
		ExitOnComplete: false, // 结束时关闭连接
		IsStrict:       true,
		Interval:       time.Second * 5, // 轮询时间
		Queues: []goworker.Queue{
			goworker.Queue{
				Name:   "Masscan",
				PerNum: 1,
			},
			goworker.Queue{
				Name:   "Nmap",
				PerNum: 1,
			},
		},
		Namespace: "PortScan:",
	}
}

// 获取redis url 和 是否需要自动获取网卡
func agentRegister() {
	client := &http.Client{}
	url := fmt.Sprintf("http://%s/api/register", logic.Conf.Http.ServerHost)
	req, err := http.NewRequest("GET", url, nil)
	token := logic.Conf.Http.Token
	req.SetBasicAuth(token, token)
	resp, err := client.Do(req)
	if err != nil {
		logic.Log.Fatal("Request verify fail, check your server host and http token\n", err)
		os.Exit(1)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	redisUrl := gjson.GetBytes(bodyText, "data.redisUrl")
	autoInterface := gjson.GetBytes(bodyText, "data.autoInterface")
	logic.Conf.RedisUrl = redisUrl.String()
	logic.Conf.Masscan.autoInterface = autoInterface.Bool()
}
