package redisc

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	log "github.com/sirupsen/logrus"
	"main/config"
	"main/util/logger"
)

func Pop(count int, queue string) []*config.Message {
	var lst []*config.Message
	rc := RedisConnPool.Get()
	defer rc.Close()
	for i := 0; i < count; i++ {
		reply, err := redis.String(rc.Do("RPOP", queue))
		if err != nil {
			if err != redis.ErrNil {
				log.WithFields(logger.Weblog).Errorf("rpop queue:%s failed, err: %v", queue, err)
			}
			break
		}

		if reply == "" || reply == "nil" {
			continue
		}

		var message config.Message
		//reply = util.ConvertToString(reply, "gbk", "utf-8")
		err = json.Unmarshal([]byte(reply), &message)
		if err != nil {
			log.WithFields(logger.Weblog).Errorf("unmarshal message failed, err: %v, redis reply: %v", err, reply)
			continue
		}

		lst = append(lst, &message)
	}

	return lst
}
