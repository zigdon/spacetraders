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

func (t *TUI) mainView(g *gocui.Gui) error {
	var err error
	maxX, maxY := t.g.Size()

	/////// main
	mainW := maxX - 51
	if maxX < 100 || (!t.windows["sidebar"] && !t.windows["msgs"]) {
		mainW = maxX - 1
	}
	err = createView(g, "main", 0, 3, mainW, maxY-4, func(v *gocui.View) error {
		v.Autoscroll = true
		v.Wrap = true
		return nil
	}, nil)
	if err != nil {
		return fmt.Errorf("can't create main view: %v", err)
	}

	/////// account
	err = createView(g, "account", 0, 0, maxX-1, 2, nil, nil)
	if err != nil {
		return fmt.Errorf("can't create account view: %v", err)
	}

	/////// input
	err = createView(g, "input", 0, maxY-3, maxX-1, maxY-1, func(v *gocui.View) error {
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

	/////// sidebar
	sideH := maxY - 40
	if maxY < 40 || !t.windows["msgs"] {
		sideH = maxY - 4
	}
	err = createView(g, "sidebar", maxX-50, 3, maxX-1, sideH, nil, func(v *gocui.View) error {
		if t.windows["sidebar"] {
			g.SetViewOnTop("sidebar")
		} else {
			g.SetViewOnBottom("sidebar")
		}
		v.Title = time.Now().Format("15:04:05")
		v.Clear()
		fmt.Fprint(v, t.UpdateView("sidebar"))
		return nil
	})
	if err != nil {
		return fmt.Errorf("can't create input view: %v", err)
	}

	/////// messages
	startM := maxY - 39
	if !t.windows["sidebar"] {
		startM = 3
	}
	err = createView(g, "msgs", maxX-50, startM, maxX-1, maxY-4, func(v *gocui.View) error {
		v.Autoscroll = true
		v.Wrap = true
		return nil
	}, func(v *gocui.View) error {
		if t.windows["msgs"] {
			g.SetViewOnTop("msgs")
		} else {
			g.SetViewOnBottom("msgs")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("can't create msgs view: %v", err)
	}
	return nil
}
