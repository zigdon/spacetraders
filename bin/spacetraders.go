package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/zigdon/spacetraders"
	"github.com/zigdon/spacetraders/cli"
	"github.com/zigdon/spacetraders/tasks"
	"github.com/zigdon/spacetraders/tui"
)

var (
	echo        = flag.Bool("echo", false, "If true, echo commands back to stdout")
	logFile     = flag.String("logfile", "/tmp/spacetraders.log", "Where should the log file be saved")
	errorsFatal = flag.Bool("errors_fatal", false, "If false, API errors are caught")
	historyFile = flag.String("history", filepath.Join(os.Getenv("HOME"), ".spacetraders.history"), "If not empty, save history between sessions")
)

// Main input loop
func loop(c *spacetraders.Client) {
	t, err := tui.Create()
	if err != nil {
		log.Fatalf("Can't init TUI: %v", err)
	}
	defer t.Close()
	cli.SetTUI(t)

	go func() {
		log.Print("TUI goroutine starting...")
		if err := t.MainLoop(); err != nil {
			log.Printf("TUI error: %v", err)
			t.Quit()
		}
		log.Print("TUI goroutine ended.")
	}()

	tq := tasks.GetTaskQueue()
	tq.SetClient(c)

	quitTQ := make(chan (bool))
	go func(q chan (bool)) {
		log.Print("TaskQueue goroutine starting...")
		for {
			select {
			case <-time.After(time.Second):
				msgs, err := tq.ProcessTasks()
				if err != nil {
					cli.Warn("Error processing background tasks: %v", err)
				}
				for _, m := range msgs {
					cli.Out(m)
				}
			case <-quitTQ:
				log.Print("TaskQueue goroutine ended.")
				break
			}
		}
	}(quitTQ)
	defer func() { quitTQ <- true }()

	log.Print("Input handling starting...")
	for line := range t.GetLine() {
		cmd, args, err := cli.ParseLine(c, line)
		if err != nil {
			cli.ErrMsg(err.Error())
			continue
		}

		if err := cmd.Do(c, args); err != nil {
			if err == cli.ErrExit {
				break
			}
			cli.ErrMsg("Error: %v", err)
			if *errorsFatal {
				log.Fatal(err)
			}
		}
		cli.Out("")
	}
	cli.Out("")
	log.Print("Input handling ended.")
}

func main() {
	flag.Parse()
	f, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Can't open log file %q: %v", *logFile, err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.Print("CLI starting...")
	c := spacetraders.New()

	if err := c.Status(); err != nil {
		log.Fatalf("Game down: %v", err)
	}

	if flag.NArg() > 0 {
		if err := c.Load(flag.Arg(0)); err != nil {
			log.Fatalf("Can't login with %q: %v", flag.Arg(0), err)
		}
	}

	loop(c)
	log.Print("Exiting CLI.\n\n")
}
