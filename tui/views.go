package tui

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jroimartin/gocui"
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
	mu *sync.Mutex
	sidebar map[string]string
	account string
}

func GetUI() *TUI {
  return t
}

func init() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatalf("can't create gui: %v", err)
	}

	t = &TUI{g: g, lines: lines, mu: &sync.Mutex{}, sidebar: make(map[string]string)}
	t.g.SetManagerFunc(t.mainView)
	t.g.Cursor = true

	if err := t.keybindings(); err != nil {
		log.Fatalf("can't set keybindings: %v", err)
	}

	// Heartbeat
	go func() {
	  for !t.quit {
		select {
		  case <- time.After(time.Second):
		  g.Update(func(_ *gocui.Gui) error {return nil})
		}
	  }
	}()
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

func (t *TUI) SetAccount(msg string) {
  t.account = msg
}

func (t *TUI) ClearSidebar() {
  t.mu.Lock()
  defer t.mu.Unlock()
  t.sidebar = make(map[string]string)
}

func (t *TUI) AddSidebar(key, msg string) {
  t.mu.Lock()
  defer t.mu.Unlock()
  t.sidebar[key] = msg
}

func (t *TUI) DelSidebar(key string) {
  t.mu.Lock()
  defer t.mu.Unlock()
  delete(t.sidebar,key)
}

func (t *TUI) GetSidebar() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	var keys []string
	for k := range t.sidebar {
	  keys = append(keys, k)
	}
	sort.Strings(keys)
	res := []string{}
	for _, k := range keys {
	  res = append(res, t.sidebar[k])
	}
	return res
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
	nv := func(name string, x0, y0, x1, y1 int,
			fNew func(*gocui.View) error,
			fUpdate func(*gocui.View) error) error {
		if v, err := t.g.SetView(name, x0, y0, x1, y1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Frame = true
			v.Autoscroll = false
			if fNew != nil {
				return fNew(v)
			}
		} else if fUpdate != nil {
		  return fUpdate(v)
		}
		return nil
	}

	var err error
	err = nv("main", 0, 3, maxX-51, maxY-4, func(v *gocui.View) error {
		t.GetView("main").Autoscroll = true
		return nil
	}, nil)
	if err != nil {
		return fmt.Errorf("can't create main view: %v", err)
	}

	err = nv("account", 0, 0, maxX-1, 2, nil, func(v *gocui.View) error {
	  v.Clear()
	  fmt.Fprint(v, t.account)
	  return nil
	})
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
	}, func(v *gocui.View) error {
	  x, _ := v.Cursor()
	  if x < len(prompt) {
		v.SetCursor(len(prompt), 0)
	  }
	  return nil
	})
	if err != nil {
		return fmt.Errorf("can't create input view: %v", err)
	}

	err = nv("sidebar", maxX-50, 3, maxX-1, maxY-4, nil, func(v *gocui.View) error {
	  v.Title = time.Now().Format("15:04:05")
	  v.Clear()
	  fmt.Fprint(v,  strings.Join(t.GetSidebar(), "\n"))
	  return nil
	})
	if err != nil {
		return fmt.Errorf("can't create input view: %v", err)
	}

	return nil
}
