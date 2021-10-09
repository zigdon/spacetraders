package tui

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/awesome-gocui/gocui"
)

var prompt = "> "
var t *TUI

type input struct {
	mu    *sync.Mutex
	lines []string
}

func (i *input) GetLine() string {
	if len(i.lines) > 0 {
		i.mu.Lock()
		defer i.mu.Unlock()
		line := i.lines[0]
		i.lines = i.lines[1:]
		return line
	}

	return ""
}

func (i *input) AddLine(line string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.lines = append(i.lines, line)
}

var lines *input = &input{mu: &sync.Mutex{}}

type TUI struct {
	g         *gocui.Gui
	lines     *input
	inputChan chan (string)
	quit      bool
	viewCache map[string]func() string
	windows   map[string]bool
	initLogs  []string
	msgs      []string
}

func GetUI() *TUI {
	return t
}

func init() {
	g, err := gocui.NewGui(gocui.OutputNormal, false)
	if err != nil {
		log.Fatalf("can't create gui: %v", err)
	}

	t = &TUI{
		g:         g,
		lines:     lines,
		viewCache: make(map[string]func() string),
		windows: map[string]bool{
			"sidebar": true,
			"logs":    false,
			"msgs":    true,
		},
		initLogs: []string{},
	}
	t.g.SetManagerFunc(t.mainView)
	t.g.Cursor = true

	if err := t.keybindings(); err != nil {
		log.Fatalf("can't set keybindings: %v", err)
	}

	// Heartbeat
	go func() {
		for !t.quit {
			select {
			case <-time.After(time.Second):
				g.Update(func(_ *gocui.Gui) error { return nil })
			}
		}
	}()
}

func (t *TUI) initLog(format string, args ...interface{}) {
	t.initLogs = append(t.initLogs, fmt.Sprintf(format, args...))
}

func (t *TUI) GetInitLogs() []string {
	defer func() { t.initLogs = []string{} }()
	return t.initLogs
}

func (t *TUI) Msg(format string, args ...interface{}) {
	t.PrintMsg("msgs", "-", format, args...)
}

func (t *TUI) GetLine() <-chan (string) {
	t.inputChan = make(chan (string))
	go func() {
		log.Print("Started GetLine goroutine")
		for !t.quit {
			l := t.lines.GetLine()
			if l == "" {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			t.inputChan <- l
		}
		close(t.inputChan)
	}()
	return t.inputChan
}

func (t *TUI) Clear(buf string) {
	t.g.Update(func(g *gocui.Gui) error {
		output, err := g.View(buf)
		if err != nil {
			return fmt.Errorf("can't get view %q: %v", buf, err)
		}
		output.Clear()
		return nil
	})
}

func (t *TUI) SetView(name string, f func() string) {
	t.viewCache[name] = f
}

func (t *TUI) UpdateView(name string) string {
	if f, ok := t.viewCache[name]; ok {
		return f()
	}
	return ""
}

func (t *TUI) Toggle(name string) error {
	if name == "all" {
		open := true
		for _, s := range t.windows {
			open = open && s
		}
		for w := range t.windows {
			t.windows[w] = !open
		}
		return nil
	}
	if _, ok := t.windows[name]; !ok {
		return fmt.Errorf("unknown window %q", name)
	}
	t.windows[name] = !t.windows[name]
	return nil
}

func (t *TUI) PrintMsg(buf, prefix, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	t.g.Update(func(g *gocui.Gui) error {
		output, err := g.View(buf)
		if err != nil {
			log.Printf("can't get view %q: %v", buf, err)
			return fmt.Errorf("can't get view %q: %v", buf, err)
		}
		if buf == "main" {
			output.Autoscroll = true
		}
		if format == "" {
			fmt.Fprintln(output)
			return nil
		}
		for i, l := range strings.Split(msg, "\n") {
			if i > 0 {
				prefix = " "
			}
			fmt.Fprintf(output, "%s %s\n", prefix, l)
		}
		return nil
	})
}

func (t *TUI) Update(f func(*gocui.Gui) error) {
	t.g.Update(f)
}
func (t *TUI) Close() {
	t.g.Close()
}

func (t *TUI) Quit() {
	t.quit = true
}

func (t *TUI) MainLoop() error {
	return t.g.MainLoop()
}
