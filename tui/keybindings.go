package tui

import (
	"fmt"
	"log"
	"strings"

	"github.com/jroimartin/gocui"
)

func (t *TUI) keybindings() error {
	if err := t.g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := t.g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, addLine); err != nil {
		return err
	}
	if err := t.g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, scrollDown); err != nil {
		return err
	}
	if err := t.g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, scrollUp); err != nil {
		return err
	}

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
