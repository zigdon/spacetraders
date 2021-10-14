package tui

import (
	"fmt"
	"time"

	"github.com/awesome-gocui/gocui"
)

func createView(g *gocui.Gui, name string, x0, y0, x1, y1 int,
	fNew func(*gocui.View) error,
	fUpdate func(*gocui.View) error) error {
	if v, err := g.SetView(name, x0, y0, x1, y1, 0); err != nil {
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
	} else {
		msg := t.UpdateView(name)
		if msg != "" {
			v.Clear()
			fmt.Fprint(v, msg)
		}
	}
	return nil
}

func mainNew(v *gocui.View) error {
	v.Autoscroll = true
	v.Wrap = true
	return nil
}

func inputNew(v *gocui.View) error {
	v.Editable = true
	fmt.Fprintf(v, prompt)
	if err := v.SetCursor(len(prompt), 0); err != nil {
		return fmt.Errorf("can't set cursor: %v", err)
	}
	if _, err := t.g.SetCurrentView("input"); err != nil {
		return fmt.Errorf("can't focus input: %v", err)
	}
	return nil
}

func inputUpdate(v *gocui.View) error {
	x, _ := v.Cursor()
	if x < len(prompt) {
		v.SetCursor(len(prompt), 0)
	}
	return nil
}

func sidebarUpdate(v *gocui.View) error {
	v.Title = time.Now().Format("15:04:05")
	v.Clear()
	fmt.Fprint(v, t.UpdateView("sidebar"))
	return nil
}

func msgsNew(v *gocui.View) error {
	v.Autoscroll = true
	v.Wrap = true
	return nil
}
