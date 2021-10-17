package tui

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
)

type layoutDirection bool
type hideLayout bool

var NotFound = fmt.Errorf("Item not found")

const (
	layoutHorizontal layoutDirection = true
	layoutVertical   layoutDirection = false

	layoutHidden  hideLayout = true
	layoutVisible hideLayout = false
)

type layoutItem struct {
	ratio   int
	fixed   int
	name    string
	hidden  hideLayout
	inner   *layoutLevel
	fNew    func(*gocui.View) error
	fUpdate func(*gocui.View) error
}

func (l *layoutItem) Equal(other *layoutItem) bool {
	if l.ratio != other.ratio {
		return false
	}
	if l.fixed != other.fixed {
		return false
	}
	if l.name != other.name {
		return false
	}
	if l.hidden != other.hidden {
		return false
	}
	if l.inner != nil {
		return l.inner.Equal(other.inner)
	}
	return true
}

func (l *layoutItem) isHidden() hideLayout {
	if l.hidden {
		return layoutHidden
	}
	if l.inner != nil {
		return l.inner.allHidden()
	}

	return layoutVisible
}

type layoutLevel struct {
	direction layoutDirection
	items     []*layoutItem
}

func (l *layoutLevel) findItem(name string) (*layoutItem, error) {
	for _, item := range l.items {
		if item.name == name {
			return item, nil
		}
		if item.inner != nil {
			found, err := item.inner.findItem(name)
			if err == nil {
				return found, err
			}
			if err != NotFound {
				return nil, err
			}
		}
	}
	return nil, NotFound
}

func (l *layoutLevel) Equal(other *layoutLevel) bool {
	if l.direction != other.direction {
		return false
	}
	if len(l.items) != len(other.items) {
		return false
	}
	for i := range l.items {
		if !l.items[i].Equal(other.items[i]) {
			return false
		}
	}

	return true
}

func (l *layoutLevel) HideItem(name string, hidden hideLayout) error {
	i, err := l.findItem(name)
	if err != nil {
		return err
	}

	i.hidden = hidden

	return nil
}

func (l *layoutLevel) ResizeItem(name string, ratio, fixed int) error {
	i, err := l.findItem(name)
	if err != nil {
		return err
	}

	i.ratio = ratio
	i.fixed = fixed

	return nil
}

func (l *layoutLevel) ReplaceItem(name string, newItem *layoutItem) bool {
	for i, item := range l.items {
		if item.name == name {
			l.items[i] = newItem
			return true
		}
		if item.inner != nil {
			found := item.inner.ReplaceItem(name, newItem)
			if found {
				return true
			}
		}
	}
	return false
}

func (l *layoutLevel) allHidden() hideLayout {
	for _, item := range l.items {
		if !item.isHidden() {
			return layoutVisible
		}
	}
	return layoutHidden
}

func (l *layoutLevel) layout(g *gocui.Gui, x0, y0, x1, y1 int, forceHidden hideLayout) error {
	var length, acc int
	var overlap int
	if !g.SupportOverlaps {
		overlap = 1
	}

	// Figure out which dimention we care about
	if l.direction == layoutHorizontal {
		length = x1 - x0 + 1
		acc = x0
	} else {
		length = y1 - y0 + 1
		acc = y0
	}

	// Add up all the (visible) fixed sizes, as they're not available for assignment
	fixed := 0
	segments := 0
	lastVisible := 0
	for i, item := range l.items {
		if forceHidden || item.isHidden() {
			continue
		}
		if item.fixed > 0 {
			fixed += item.fixed
		} else {
			segments += item.ratio
		}
		lastVisible = i
	}
	if length < fixed {
		return fmt.Errorf("window too small for fixed sizes: %d < %d", length, fixed)
	}
	length -= fixed

	// The rest of the space gets split between the segments
	unit := -1
	left := -1
	if segments > 0 {
		unit = length / segments
		left = length % segments
	}

	if unit == 0 {
		return fmt.Errorf("window too small for allocated units: length=%d, segments=%d", length, segments)
	}

	for idx, item := range l.items {
		// Make sure we still create all the views, even if they're not visible
		var err error
		if forceHidden || item.isHidden() {
			if item.inner != nil {
				err = item.inner.layout(g, x0, y0, x1, y1, layoutHidden)
			} else {
				err = createView(g, item.name, x0, y0, x1, y1, 0, item.fNew, item.fUpdate)
				g.SetViewOnBottom(item.name)
			}
			if err != nil {
				return fmt.Errorf("error creating layout: %v", err)
			}
			continue
		}

		var assignment int
		if item.fixed == 0 {
			assignment = unit * item.ratio
		} else {
			assignment = item.fixed
		}

		// The last item gets the leftovers
		if idx == lastVisible {
			assignment += left
		}

		ix0, ix1, iy0, iy1 := x0, x1, y0, y1
		if l.direction == layoutHorizontal {
			ix0 = acc
			ix1 = acc + assignment - overlap
			if ix1 > x1 {
				ix1 = x1
			}
		} else {
			iy0 = acc
			iy1 = acc + assignment - overlap
			if iy1 > y1 {
				iy1 = y1
			}
		}
		acc += assignment

		if item.inner != nil {
			err = item.inner.layout(g, ix0, iy0, ix1, iy1, layoutVisible)
		} else {
			err = createView(g, item.name, ix0, iy0, ix1, iy1, 0, item.fNew, item.fUpdate)
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
	  2, { VERTICAL [
		1, "v3"
		1, "v4"
		1, "v5"
	  ]},
	]},
  ]}
*/

func (l *ratioLayout) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	return l.definition.layout(g, 0, 0, maxX-1, maxY-1, layoutVisible)
}

// Helper to generate a simple row/column of a layout. If sizes are negative,
// they're used as "fixed", otherwise as "ratio".
func GenerateLine(dir layoutDirection, name string, sizes []int) *layoutLevel {
	items := []*layoutItem{}
	for i, r := range sizes {
		if r > 0 {
			items = append(items, &layoutItem{
				ratio: r,
				name:  fmt.Sprintf("%s%d", name, i+1),
			})
		} else {
			items = append(items, &layoutItem{
				fixed: -r,
				name:  fmt.Sprintf("%s%d", name, i+1),
			})
		}
	}
	return &layoutLevel{direction: dir, items: items}
}

func createView(g *gocui.Gui, name string, x0, y0, x1, y1 int, overlaps byte,
	fNew func(*gocui.View) error,
	fUpdate func(*gocui.View) error) error {
	if v, err := g.SetView(name, x0, y0, x1, y1, overlaps); err != nil {
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
