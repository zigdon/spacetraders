package tasks

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
	when   time.Time
	repeat time.Duration
	f      func(c *spacetraders.Client) error
	msg    string
}

type taskQueue struct {
	mu       sync.Mutex
	tasks    map[string]*task
	c        *spacetraders.Client
	nextTime time.Time
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
	if tq.nextTime.After(time.Now()) {
		return nil, nil
	}

	var msgs []string
	var errs []error
	var err error
	for k, t := range tq.tasks {
		if t.when.Before(time.Now()) {
			log.Printf("executing task %q", k)
			msg, err := tq.Run(k)
			if msg != "" {
				msgs = append(msgs, msg)
			}
			if err != nil {
				errs = append(errs, err)
			}
			if t.repeat == 0 {
				delete(tq.tasks, k)
				continue
			}

			log.Printf("requeuing task %q in %s", k, t.repeat.Truncate(time.Second))
			t.when = time.Now().Add(t.repeat)
			if t.when.Before(tq.nextTime) {
				tq.nextTime = t.when
			}
		}
	}

	if len(errs) > 0 {
		err = fmt.Errorf("%d errors while processing background tasks: %v", len(errs), errs)
	}

	tq.findNext()

	return msgs, err
}

func (tq *taskQueue) Run(key string) (string, error) {
	t, ok := tq.tasks[key]
	if !ok {
		return "", fmt.Errorf("unknown task %q", key)
	}
	msg := t.msg
	var err error
	if t.f != nil {
		err = t.f(tq.c)
	}

	return msg, err

}

func (tq *taskQueue) Add(key, msg string, when time.Time, repeat time.Duration, f func(*spacetraders.Client) error) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	log.Printf("Adding task %q at %s (in %s): %q (f: %v)",
		key, when, when.Sub(time.Now()).Truncate(time.Second), msg, f != nil)
	if _, ok := tq.tasks[key]; ok {
		return
	}
	tq.tasks[key] = &task{
		when:   when,
		repeat: repeat,
		msg:    msg,
		f:      f,
	}

	if when.Before(tq.nextTime) {
		tq.nextTime = when
	}
}

func (tq *taskQueue) findNext() {
	var next time.Time
	for _, t := range tq.tasks {
		if next.IsZero() || t.when.Before(next) {
			next = t.when
		}
	}

	tq.nextTime = next
}

func (tq *taskQueue) GetNext() time.Time {
	return tq.nextTime
}
