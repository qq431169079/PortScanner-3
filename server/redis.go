package server

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

func newRedis(redisUrl string) *redis.Client {
	option, err := redis.ParseURL(redisUrl)
	if err != nil {
		panic(err)
	}
	option.PoolSize = 100
	return redis.NewClient(option)
}

func heartBeat() {
	for {
		agentList, err := logic.RedisClient.SMembers("agent:ip").Result()
		if err != nil && err != redis.Nil {
			logic.Log.Error("Redis client get agent member error, ", err)
			continue
		}
		currentTime := time.Now().Unix()
		for _, agentIp := range agentList {
			agentTime, err := logic.RedisClient.HGet("agent:ip:time", agentIp).Result()
			if err != nil && err != redis.Nil {
				logic.Log.Errorf("Agent %s get time err,", agentIp, err)
				continue
			}
			resultTime, err := strconv.ParseInt(agentTime, 10, 64)
			if err != nil {
				logic.Log.Error("String to int64 error, ", err)
				continue
			}
			if currentTime-resultTime > 100 {
				message := fmt.Sprintf("节点 %s 停止心跳，请检查", agentIp)
				logic.Log.Error(message)
				if _, ok := logic.messageNum[agentIp]; ok {
					if logic.messageNum[agentIp] < 3 {
						if logic.Conf.EMail.Enable {
							go func() {
								err = logic.MailClient.Send("节点停止心跳", message)
								if err != nil {
									logic.Log.Error("Send mail error,", err)
								}
							}()
							logic.messageNum[agentIp]++
						}
					}
				}
			} else {
				if logic.messageNum[agentIp] > 0 {
					message := fmt.Sprintf("节点 %s 已经重启", agentIp)
					logic.Log.Error(message)
					if logic.Conf.EMail.Enable {
						go func() {
							err = logic.MailClient.Send("节点已经重启", message)
							if err != nil {
								logic.Log.Error("Send mail Error, ", err)
							}
						}()
					}
				}
				logic.messageNum[agentIp] = 0
			}
		}
		time.Sleep(time.Minute * 1)
	}
}
