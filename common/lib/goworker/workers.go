package goworker

import (
	"encoding/json"
	"fmt"
)

var (
	workers map[string]workerFunc
)

func init() {
	workers = make(map[string]workerFunc)
}

// Register registers a goworker worker function. Class
// refers to the Ruby name of the class which enqueues the
// job. Worker is a function which accepts a queue and an
// arbitrary array of interfaces as arguments.
func Register(class string, worker workerFunc) {
	workers[class] = worker
}

func Enqueue(job *Job) error {
	err := Init()
	if err != nil {
		return err
	}

	buffer, err := json.Marshal(job.Payload)
	if err != nil {
		logger.Error("Cant marshal payload on enqueue")
		return err
	}

	err = redisClient.RPush(fmt.Sprintf("%squeue:%s", workerSettings.Namespace, job.Queue), buffer).Err()
	if err != nil {
		logger.Error("Cant push to queue")
		return err
	}

	return nil
}

func EnqueueMoreOne(queue string, payloads ...*Payload) error {
	if len(payloads) > 0 {
		err := Init()
		if err != nil {
			return err
		}

		bufferList := make([]interface{}, len(payloads))
		for i, p := range payloads {
			bufferList[i], err = json.Marshal(p)
			if err != nil {
				logger.Error("Cant marshal payload on enqueue")
				return err
			}
		}

		err = redisClient.RPush(fmt.Sprintf("%squeue:%s", workerSettings.Namespace, queue), bufferList...).Err()
		if err != nil {
			logger.Error("Cant push to queue")
			return err
		}
	}
	return nil
}
