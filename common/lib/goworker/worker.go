package goworker

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

type worker struct {
	process
}

func newWorker(id string, queues []Queue) (*worker, error) {
	process, err := newProcess(id, queues)
	if err != nil {
		return nil, err
	}
	return &worker{
		process: *process,
	}, nil
}

func (w *worker) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.String())
}

func (w *worker) fail(job *Job, err error) error {
	failure := &failure{
		FailedAt:  time.Now(),
		Payload:   job.Payload,
		Exception: "Error",
		Error:     err.Error(),
		Worker:    w,
		Queue:     job.Queue,
	}
	buffer, err := json.Marshal(failure)
	if err != nil {
		return err
	}

	return redisClient.RPush(fmt.Sprintf("%sfailed", workerSettings.Namespace), buffer).Err()
}

func (w *worker) work(jobs <-chan *Job, monitor *sync.WaitGroup) {
	monitor.Add(1)
	go func() {
		defer func() {
			defer monitor.Done()
		}()
		for job := range jobs {
			if workerFunc, ok := workers[job.Payload.Class]; ok {
				w.run(job, workerFunc)
				logger.Debugf("done: (Job{%s} | %s | %v)", job.Queue, job.Payload.Class, job.Payload.Args)
			} else {
				errorLog := fmt.Sprintf("No worker for %s in queue %s with args %v", job.Payload.Class, job.Queue, job.Payload.Args)
				logger.Error(errorLog)

				err := w.fail(job, errors.New(errorLog))
				if err != nil {
					logger.Errorf("Error on save failed job in worker %v: %v", w, err)
				}
			}
		}
	}()
}

func (w *worker) run(job *Job, workerFunc workerFunc) {
	var err error
	defer func() {
		if err != nil {
			err := w.fail(job, err)
			if err != nil {
				logger.Errorf("Error on save failed job in worker %v: %v", w, err)
			}
		}
	}()
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()

	err = workerFunc(job.Queue, job.Payload.Args...)
}
