package tui

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
)

type layoutDirection bool

const (
	layoutHorizontal layoutDirection = true
	layoutVertical   layoutDirection = false
)

type layoutItem struct {
	ratio   int
	fixed   int
	name    string
	inner   *layoutLevel
	fNew    func(*gocui.View) error
	fUpdate func(*gocui.View) error
}

type layoutLevel struct {
	direction layoutDirection
	items     []*layoutItem
}

func (l *layoutLevel) layout(g *gocui.Gui, x0, y0, x1, y1 int) error {
	var length, acc int

	// Figure out which dimention we care about
	if l.direction == layoutHorizontal {
		length = x1 - x0 + 1
		acc = x0
	} else {
		length = y1 - y0 + 1
		acc = y0
	}

	// Add up all the fixed sizes, as they're not available for assignment
	fixed := 0
	segments := 0
	for _, item := range l.items {
		if item.fixed > 0 {
		  fixed += item.fixed
		} else {
		  segments += item.ratio
		}
	}
	if length < fixed {
		return fmt.Errorf("window too small for fixed sizes: %d < %d", length, fixed)
	}
	length -= fixed

	// The rest of the space gets split between the segments
	unit := length / segments
	left := length % segments

	if unit == 0 {
	  return fmt.Errorf("window too small for allocated units: length=%d, segments=%d", length, segments)
	}

	for idx, item := range l.items {
		var err error
		var assignment int
		if item.fixed == 0 {
		  assignment = unit * item.ratio
		} else {
		  assignment = item.fixed
		}

		// The last item gets the leftovers
		if idx == len(l.items) {
			assignment += length % segments
		}

		var ix0, ix1, iy0, iy1 int
		if l.direction == layoutHorizontal {
			ix0 = acc
			ix1 = acc + assignment -1
			if idx == len(l.items)-1 {
			  ix1 += left
			}
			iy0, iy1 = y0, y1
		} else {
			ix0, ix1 = x0, x1
			iy0 = acc
			iy1 = acc + assignment -1
			if idx == len(l.items)-1 {
			  iy1 += left
			}
		}
		acc += assignment

		if item.inner != nil {
			err = item.inner.layout(g, ix0, iy0, ix1, iy1)
		} else {
			err = createView(g, item.name, ix0, iy0, ix1, iy1, item.fNew, item.fUpdate)
		}

		if err != nil {
			return fmt.Errorf("error creating layout: %v", err)
		}
	}

	return nil
}

type ratioLayout struct {
	definition *layoutLevel
}

/* Layout views by ratio, recursively
	  +-------------------------------+
	  | v1                            |
	  +-------------------------------+
	  | v2     |  v3                  |
	  |        +----------------------+
	  |        |  v4                  |
	  |        +----------------------+
	  |        |  v5                  |
	  +-------------------------------+

  { VERTICAL [
	1, "v1"
	3, { HORIZONTAL [
	  1, "v2", 5
	  1, { VERTICAL [
		1, "v3"
		1, "v4"
		1, "v5"
	  ]},
	]},
  ]}
*/

func (l *ratioLayout) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	return l.definition.layout(g, 0, 0, maxX-1, maxY-1)
}

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

