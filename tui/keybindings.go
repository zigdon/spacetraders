package tui

import (
	"fmt"
	"log"
	"strings"

	"github.com/jroimartin/gocui"
)

func (t *TUI) keybindings() error {
	type binding struct {
		view string
		key gocui.Key
		mod gocui.Modifier
		f func(g *gocui.Gui, v *gocui.View) error
	}
	for _, b := range []binding{
		{"", gocui.KeyCtrlC, gocui.ModNone, quit},
		{"", gocui.KeyEnter, gocui.ModNone, addLine},
		{"", gocui.KeyBackspace, gocui.ModNone, backspace},
		{"", gocui.KeyBackspace2, gocui.ModNone, backspace},
		{"", gocui.KeyArrowDown, gocui.ModNone, scrollDown},
		{"", gocui.KeyArrowUp, gocui.ModNone, scrollUp},
	} {
		if err := t.g.SetKeybinding(b.view, b.key, b.mod, b.f); err != nil {
			return err
		}
	}

	return nil
}

func backspace(g *gocui.Gui, v *gocui.View) error {
  x, _ := v.Cursor()
  if x <= len(prompt) {
	return nil
  }
  v.EditDelete(true)
  return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	log.Printf("Quitting")
	return gocui.ErrQuit
}

func addLine(g *gocui.Gui, v *gocui.View) error {
	line := v.Buffer()
	if strings.HasPrefix(line, prompt) {
		line = line[len(prompt):]
	}
	line = strings.TrimSpace(line)
	if len(line) > 0 {
		lines.AddLine(strings.TrimSpace(line))
	}
	v.Clear()
	fmt.Fprint(v, prompt)
	v.SetCursor(len(prompt), 0)

	return nil
}

func scrollDown(g *gocui.Gui, _ *gocui.View) error {
	return scroll(g, 1)
}

func scrollUp(g *gocui.Gui, _ *gocui.View) error {
	return scroll(g, -1)
}

func scroll(g *gocui.Gui, offset int) error {
	v, err := g.View("main")
	if err != nil {
		return err
	}
	v.Autoscroll = false
	x, y := v.Origin()
	y += offset
	if y < 0 {
		y = 0
	}
	if y > len(v.BufferLines()) {
		y = len(v.BufferLines())
	}

	return v.SetOrigin(x, y)
}
