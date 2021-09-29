package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/chzyer/readline"
	"github.com/zigdon/spacetraders"
	"github.com/zigdon/spacetraders/cli"
)

var (
	echo        = flag.Bool("echo", false, "If true, echo commands back to stdout")
	logFile     = flag.String("logfile", "/tmp/spacetraders.log", "Where should the log file be saved")
	errorsFatal = flag.Bool("errors_fatal", false, "If false, API errors are caught")
	historyFile = flag.String("history", filepath.Join(os.Getenv("HOME"), ".spacetraders.history"), "If not empty, save history between sessions")
)

// Main input loop
func loop(c *spacetraders.Client) {
	config := &readline.Config{Prompt: "> "}
	if *historyFile != "" {
		config.HistoryFile = *historyFile
	}
	r, err := readline.NewEx(config)
	if err != nil {
		log.Fatalf("Can't readline: %v", err)
	}

	tq := cli.GetTaskQueue()
	tq.SetClient(c)
	for {
		msgs, err := tq.ProcessTasks()
		if err != nil {
			cli.Warn("Error processing background tasks: %v", err)
		}
		for _, m := range msgs {
			cli.Out(m)
		}
		line, stop := getLine(r)
		if stop {
			break
		}
		if line == "" {
			continue
		}

		cmd, args, err := cli.ParseLine(c, line)
		if err != nil {
			continue
		}

		if err := cmd.Do(c, args); err != nil {
			if *errorsFatal {
				log.Fatal(err)
			}
			cli.ErrMsg("Error: %v", err)
		}
		cli.Out("")
	}
}

// Main input utilities
func getLine(r *readline.Instance) (string, bool) {
	for {
		line, err := r.Readline()
		if err != nil {
			if err == io.EOF {
				return "", true
			}
			cli.ErrMsg("Error while reading input: %v", err)
			return "", true
		}
		if *echo {
			fmt.Printf("> %s\n", line)
		}
		return line, false
	}
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
