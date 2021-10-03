package tui

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jroimartin/gocui"
)

var prompt = "> "

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
}

func Create() (*TUI, error) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, fmt.Errorf("can't create gui: %v", err)
	}

	log.Printf("Creating TUI")
	t := &TUI{g: g, lines: lines}
	t.g.SetManagerFunc(t.mainView)
	t.g.Cursor = true

	log.Printf("Setting keybindings")
	if err := t.keybindings(); err != nil {
		return nil, fmt.Errorf("can't set keybindings: %v", err)
	}

	return t, nil
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

func (t *TUI) GetView(name string) *gocui.View {
	v, err := t.g.View(name)
	if err != nil {
		log.Fatalf("Error getting view %q: %v", name, err)
	}
	return v
}

func (t *TUI) PrintMsg(buf, prefix, format string, args ...interface{}) {
	t.g.Update(func(g *gocui.Gui) error {
		output, err := g.View(buf)
		if err != nil {
			return fmt.Errorf("can't get view %q: %v", buf, err)
		}
		if buf == "main" {
			output.Autoscroll = true
		}
		if format == "" {
			fmt.Fprintln(output)
			return nil
		}
		for i, l := range strings.Split(fmt.Sprintf(format, args...), "\n") {
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

func (t *TUI) mainView(g *gocui.Gui) error {
	t.g = g
	maxX, maxY := t.g.Size()
	nv := func(name string, x0, y0, x1, y1 int, f func(*gocui.View) error) error {
		if v, err := t.g.SetView(name, x0, y0, x1, y1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Frame = true
			v.Title = name
			v.Autoscroll = false
			if f != nil {
				return f(v)
			}
		}
		return nil
	}

	var err error
	err = nv("main", 0, 3, maxX-31, maxY-4, func(v *gocui.View) error {
		t.GetView("main").Autoscroll = true
		return nil
	})
	if err != nil {
		return fmt.Errorf("can't create main view: %v", err)
	}

	err = nv("account", 0, 0, maxX-1, 2, nil)
	if err != nil {
		return fmt.Errorf("can't create account view: %v", err)
	}

	err = nv("input", 0, maxY-3, maxX-1, maxY-1, func(v *gocui.View) error {
		v.Editable = true
		fmt.Fprintf(v, prompt)
		if err := v.SetCursor(len(prompt), 0); err != nil {
			return fmt.Errorf("can't set cursor: %v", err)
		}
		if _, err := t.g.SetCurrentView("input"); err != nil {
			return fmt.Errorf("can't focus input: %v", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("can't create input view: %v", err)
	}

	err = nv("sidebar", maxX-30, 3, maxX-1, maxY-4, nil)
	if err != nil {
		return fmt.Errorf("can't create input view: %v", err)
	}

	return nil
}
