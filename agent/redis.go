package agent

import (
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

// redis heart beat
func heartBeat() {
	for {
		time.Sleep(time.Second * 3)
		logic.Log.Trace("Send heart beat start, ", logic.localIp)
		err := logic.RedisClient.SAdd("agent:ip", logic.localIp).Err()
		if err != nil && err != redis.Nil {
			logic.Log.Error("Agent SAdd error, ", err)
		}
		err = logic.RedisClient.HSet("agent:ip:time", logic.localIp, time.Now().Unix()).Err()
		if err != nil && err != redis.Nil {
			logic.Log.Error("Agent HSet error, ", err)
		}
		logic.Log.Trace("Send heart beat success, ", logic.localIp)
	}
}
