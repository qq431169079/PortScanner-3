package server

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/darkMoon1973/PortScanner/common/lib/go-nmap"
	"github.com/darkMoon1973/PortScanner/common/lib/goworker"
	"github.com/darkMoon1973/PortScanner/common/util"
	"github.com/darkMoon1973/email"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

// 唯一全局逻辑变量
var logic *Logic

type Logic struct {
	messageNum  map[string]int
	Log         *logrus.Entry
	Conf        *config
	RedisClient *redis.Client
	MongoDriver MongoDriver
	MailClient  *email.AuthMail
}

// 初始化
func init() {
	rand.Seed(time.Now().Unix())
	logic = newLogic()
	logic.Log, logic.Conf = configParse()
	logic.newMongoDB()
	logic.newRedis()
	logic.newGoworker()
	if logic.Conf.EMail.Enable {
		logic.newEmail()
	}
	logic.messageNum = map[string]int{}
}

func newLogic() *Logic {
	return &Logic{}
}

func (l *Logic) newRedis() {
	l.RedisClient = newRedis(l.Conf.RedisUrl)
	l.Log.Debug("init redis success")
}

func (l *Logic) newMongoDB() {
	l.MongoDriver = MongoDriver{
		Url:    l.Conf.Mongo.Url,
		DbName: l.Conf.Mongo.DBName,
	}
	l.MongoDriver.Init()
	l.Log.Debug("init mongodb success")
}

func (l *Logic) newGoworker() {
	goworker.SetSettings(l.Conf.workerSetting)
	l.Log.Debug("init goworker success")
}

func (l *Logic) newEmail() {
	var err error
	l.MailClient, err = email.New(
		l.Conf.APPName,
		l.Conf.EMail.SMTPHost,
		l.Conf.EMail.SMTPPort,
		l.Conf.EMail.Username,
		l.Conf.EMail.Password,
		l.Conf.EMail.ToList,
	)
	if err != nil {
		panic(err)
	}
	l.Log.Debug("init email success")
}

func Run() {
	logic.Log.Debug("Run Port Scan Success!")
	go func() {
		newHttpServer(logic.Conf.Http.Port)
	}()

	go func() {
		heartBeat()
	}()
	// 将本地文件导入 Mongodb
	if logic.Conf.IpList.Enable {
		insertIp(logic.Conf.IpList.ScanList, logic.Conf.IpList.WhiteList)
	}

	// 从 Mongodb 获取需要扫描的IP
	tasks := make(chan ipInfo)
	scanList := getScanList()
	// 将需扫描 IP 放入 channel
	go func() {
		for {
			count := 0
			start := time.Now().Unix()
			for _, v := range scanList {
				tasks <- v
				count++
				logic.Log.Infof("Add count is %d, ip is %s", count, v.Ip)
			}
			end := time.Now().Unix()
			message := fmt.Sprintf("扫描任务添加成功，约耗时：%d开始时间: %s, 结束时间:%s",
				end-start, util.TimeToStr(start), util.TimeToStr(end))
			logic.Log.Debug(message)
			if logic.Conf.EMail.Enable {
				go func() {
					err := logic.MailClient.Send("此轮扫描任务添加完成", message)
					if err != nil {
						logic.Log.Error("Send email fail, ", err)
					}
				}()
			}
		}
	}()

	go func() {
		for {
			currentTaskNum, err := logic.RedisClient.LLen("PortScan:queue:Masscan").Result()
			if err != nil {
				logic.Log.Error("Get masscan queue error, ", err)
			}
			n := int(currentTaskNum)
			if n < logic.Conf.WorkerNum {
				taskNum := logic.Conf.WorkerNum - n
				for i := 0; i < taskNum; i++ {
					v := <-tasks
					err = taskToMasscan("Masscan", "Masscan", v.Ip, logic.Conf.Masscan.Rate, logic.Conf.Masscan.PortRange, v.IsPublic)
					if err != nil {
						logic.Log.Error("goworker push task fail, ", err)
					} else {
						logic.Log.Debug("push task success", v.Ip)
					}
				}
			} else {
				time.Sleep(time.Second)
			}
		}
	}()

	nmapTicker := time.NewTicker(time.Second * 10)
	go func() {
		for {
			select {
			case <-nmapTicker.C:
				if s2, err := logic.RedisClient.LLen("nmapResults").Result(); err != nil {
					logic.Log.Error("Get nmap results value fail, ", err)
				} else {
					conn := int(s2)
					if conn > 500 {
						conn = 500
					}
					nmapResultsToMongo(conn)
				}
			}
		}
	}()

	select {}
}

func nmapResultsToMongo(redisConn int) {
	for i := 0; i < redisConn; i++ {
		go func() {
			var n nmap.Result
			reply, err := logic.RedisClient.LPop("nmapResults").Bytes()
			if err != nil {
				time.Sleep(time.Second)
				logic.Log.Error("Lpop from redis nmapResults queue fail, ", err)
				return
			}
			if reply != nil {
				err = json.Unmarshal(reply, &n)
				err = resultToMongo(n)
				if err != nil {
					logic.Log.Error("Save to nmap fail", err)
				}
			}
		}()
	}
}
