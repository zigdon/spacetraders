package tui

import (
	"fmt"
	"time"

	"github.com/awesome-gocui/gocui"
)

var sideBars = &layoutLevel{
	direction: layoutVertical,
	items: []*layoutItem{
		{
			ratio:   1,
			name:    "sidebar",
			fUpdate: sidebarUpdate,
		},
		{
			ratio: 1,
			name:  "msgs",
			fNew:  msgsNew,
		},
	},
}

var content = &layoutLevel{
	direction: layoutHorizontal,
	items: []*layoutItem{
		{
			ratio: 3,
			name:  "main",
			fNew:  mainNew,
		},
		{
			ratio: 1,
			inner: sideBars,
		},
	},
}

var mainLayout = &ratioLayout{
	&layoutLevel{
		direction: layoutVertical,
		items: []*layoutItem{
			{
				fixed: 3,
				name:  "account",
			},
			{
				ratio: 1,
				inner: content,
			},
			{
				fixed: 3,
				name:  "input",
				fNew:  inputNew,
			},
		},
	},
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
