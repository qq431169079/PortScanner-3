package goworker

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
)

type Queue struct {
	Name   string
	PerNum int // 每次请求从队列中获取的任务数（真正获取的时候，如果队列中任务数不足，实际获取到的会小于这个值）
}

type process struct {
	Hostname string
	Pid      int
	ID       string
	Queues   []Queue
}

func newProcess(id string, queues []Queue) (*process, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &process{
		Hostname: hostname,
		Pid:      os.Getpid(),
		ID:       id,
		Queues:   queues,
	}, nil
}

func (p *process) String() string {
	l := make([]string, len(p.Queues))
	for i, v := range p.Queues {
		l[i] = v.Name
	}
	return fmt.Sprintf("%s:%d-%s:%s", p.Hostname, p.Pid, p.ID, strings.Join(l, ","))
}

func (p *process) queues(strict bool) []Queue {
	// If the queues order is strict then just return them.
	if strict {
		return p.Queues
	}

	// If not then we want to to shuffle the queues before returning them.
	queues := make([]Queue, len(p.Queues))
	for i, v := range rand.Perm(len(p.Queues)) {
		queues[i] = p.Queues[v]
	}
	return queues
}
