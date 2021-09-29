package cli

import (
	"fmt"
	"log"
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

func init() {
	tq = &taskQueue{
		tasks: make(map[string]*task),
	}
}

func GetTaskQueue() *taskQueue {
	return tq
}

func (tq *taskQueue) SetClient(c *spacetraders.Client) {
	tq.c = c
}

func (tq *taskQueue) ProcessTasks() ([]string, error) {
	var msgs []string
	var errs []error
	var err error
	for k, t := range tq.tasks {
		if t.when.Before(time.Now()) {
			log.Printf("executing task %q", k)
			msgs = append(msgs, t.msg)
			if t.f != nil {
				if err := t.f(tq.c); err != nil {
					errs = append(errs, err)
				}
			}
			delete(tq.tasks, k)
		}
	}

	if len(errs) > 0 {
		err = fmt.Errorf("%d errors while processing background tasks: %v", len(errs), errs)
	}

	return msgs, err
}

func (tq *taskQueue) Add(key, msg string, f func(*spacetraders.Client) error, when time.Time) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	log.Printf("Adding task %q at %s (in %s): %q (f: %v)",
		key, when, when.Sub(time.Now()).Truncate(time.Second), msg, f != nil)
	if _, ok := tq.tasks[key]; ok {
		return
	}
	tq.tasks[key] = &task{
		when: when,
		msg:  msg,
		f:    f,
	}
}
