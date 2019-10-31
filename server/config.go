package server

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/darkMoon1973/PortScanner/common/lib/goworker"
	"github.com/darkMoon1973/PortScanner/common/lib/logs"
	"github.com/sirupsen/logrus"
)

const configFile = "config.toml"

type (
	config struct {
		APPName       string      `toml:"app_name"`
		RedisUrl      string      `toml:"redis_url"`
		WorkerNum     int         `toml:"worker_num"`
		IpList        ipListFile  `toml:"ip_list_file"`
		Http          httpConf    `toml:"http_config"`
		Mongo         mongodbConf `toml:"mongodb_config"`
		Log           logConf     `toml:"log_config"`
		EMail         mailConf    `toml:"email_config"`
		Masscan       masscanConf `toml:"masscan_config"`
		workerSetting goworker.WorkerSettings
	}
	ipListFile struct {
		Enable    bool   `toml:"enable"`
		WhiteList string `toml:"white_list"`
		ScanList  string `toml:"scan_list"`
	}
	masscanConf struct {
		Enable    bool   `toml:"auto_interface"`
		PortRange string `toml:"port_range"`
		Rate      string `toml:"scan_rate"`
	}
	mailConf struct {
		Enable   bool     `toml:"enable"`
		Username string   `toml:"username"`
		Password string   `toml:"password"`
		SMTPHost string   `toml:"smtp_host"`
		SMTPPort string   `toml:"smtp_port"`
		ToList   []string `toml:"to_list"`
	}
	httpConf struct {
		Token string `toml:"http_token"`
		Port  int    `toml:"http_port"`
	}
	mongodbConf struct {
		Url    string `toml:"mongodb_url"`
		DBName string `toml:"mongodb_dbname"`
	}
	logConf struct {
		Level string `toml:"log_level"`
		File  string `toml:"log_file"`
	}
)

// 解析 toml 相关配置
func configParse() (*logrus.Entry, *config) {
	conf := new(config)
	_, err := toml.DecodeFile(configFile, conf)
	if err != nil {
		fmt.Println(err)
		panic("Please check your config file")
	}

	conf.workerSetting = workerSetting()
	conf.workerSetting.URI = conf.RedisUrl
	conf.workerSetting.Concurrency = conf.WorkerNum
	log := logs.GetLogger(conf.Log.Level, conf.Log.File)
	return log, conf
}

// 默认 worker 配置
func workerSetting() goworker.WorkerSettings {
	return goworker.WorkerSettings{
		Queues: []goworker.Queue{
			goworker.Queue{
				Name:   "Masscan",
				PerNum: 2000,
			},
		},
		Connections:    100,
		UseNumber:      true,
		ExitOnComplete: false, // 结束时关闭
		IsStrict:       true,
		Interval:       time.Second, // poller 线程轮询时间
		Namespace:      "PortScan:",
	}
}
