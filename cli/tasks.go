package cli

import (
	"fmt"
	"sync"
	"time"

	"github.com/zigdon/spacetraders"
)

var (
	tq *taskQueue
)

type task struct {
	when time.Time
	f    func(c *spacetraders.Client) error
	msg  string
}

type taskQueue struct {
	mu    sync.Mutex
	tasks map[string]*task
	c     *spacetraders.Client
}

func NewTaskQueue(c *spacetraders.Client) *taskQueue {
	once := sync.Once{}
	once.Do(func() {
		tq = &taskQueue{
			tasks: make(map[string]*task),
			c:     c,
		}
	})
	return tq
}

func GetTaskQueue() *taskQueue {
	return tq
}

func (tq *taskQueue) ProcessTasks() ([]string, error) {
	var msgs []string
	var errs []error
	var err error
	for _, t := range tq.tasks {
		if t.when.Before(time.Now()) {
			msgs = append(msgs, t.msg)
			if err := t.f(tq.c); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		err = fmt.Errorf("%d errors while processing background tasks: %v", errs)
	}

	return msgs, err
}

func (tq *taskQueue) Add(key, msg string, f func(*spacetraders.Client) error, when time.Time) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if _, ok := tq.tasks[key]; ok {
		return
	}
	tq.tasks[key] = &task{
		when: when,
		msg:  msg,
		f:    f,
	}
}
