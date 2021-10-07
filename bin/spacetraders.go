package main

import (
	"flag"
	"log"
	"os"
	"sort"
	"strings"
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
	saveFile    = flag.String("savefile", "spacetraders.save", "What is the file to use as the default save")
)

func loop(c *spacetraders.Client) {
	t := tui.GetUI()
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
}

func runUI() {
	t := tui.GetUI()
	go func() {
		log.Print("TUI goroutine starting...")
		if err := t.MainLoop(); err != nil {
			log.Printf("TUI error: %v", err)
			t.Quit()
		}
		log.Print("TUI goroutine ended.")
	}()
}

func runTQ(c *spacetraders.Client) chan (bool) {
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
				if len(msgs) > 0 {
					cli.Out("")
				}
			case <-quitTQ:
				log.Print("TaskQueue goroutine ended.")
				break
			}
		}
	}(quitTQ)
	return quitTQ
}

func createTasks() {
	tq := tasks.GetTaskQueue()
	t := tui.GetUI()
	cache := spacetraders.GetCache()

	t.SetView("account", func() string {
		us := cache.RestoreObjs(spacetraders.USEROBJ)
		if len(us) == 0 {
			return ""
		}
		u := us[0].(*spacetraders.User)
		return u.Short()
	})

	tq.Add("updateAccount", "", time.Now(), time.Minute, func(c *spacetraders.Client) error {
		_, err := c.Account()
		if err != nil {
			log.Printf("Can't display account info: %v", err)
		}
		return nil
	})

	t.SetView("sidebar", func() string {
		msg := []string{}
		so := cache.RestoreObjs(spacetraders.SHIPOBJ)
		if len(so) == 0 {
			return ""
		}
		var ships []*spacetraders.Ship
		for _, s := range so {
			ships = append(ships, s.(*spacetraders.Ship))
		}
		sort.Slice(ships, func(i, j int) bool {
			return ships[i].ShortID < ships[j].ShortID
		})
		for _, s := range ships {
			msg = append(msg, s.Sidebar())
		}

		return strings.Join(msg, "\n")
	})

	tq.Add("updateShips", "", time.Now(), time.Minute, func(c *spacetraders.Client) error {
		_, err := c.MyShips()
		if err != nil {
			return nil
		}
		return nil
	})

	tq.Add("processRoutes", "", time.Now().Add(10*time.Second), 20*time.Second, func(c *spacetraders.Client) error {
	  return cli.ProcessRoutes(c)
	})
}

func autoLoad() {
	if *saveFile == "" {
		return
	}

	if _, err := os.Stat(*saveFile); os.IsNotExist(err) {
		log.Printf("No autosave %q found, skipping", *saveFile)
		return
	}

	if err := cli.Load(*saveFile); err != nil {
		log.Fatalf("Failed to autoload %q: %v", *saveFile, err)
	}
}
func autoSave() {
	if *saveFile == "" {
		return
	}

	if err := cli.Save(*saveFile); err != nil {
		log.Fatalf("Failed to autosave %q: %v", *saveFile, err)
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

	t := tui.GetUI()
	defer t.Close()

	for _, l := range t.GetInitLogs() {
		log.Printf("init: %s", l)
	}

	cli.SetTUI(t)
	runUI()
	quitTQ := runTQ(c)
	createTasks()
	autoLoad()
	loop(c)
	quitTQ <- true
	log.Print("Exiting CLI.\n\n")
	autoSave()
}
