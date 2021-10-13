package tui

import (
	"fmt"
	"log"

	"github.com/awesome-gocui/gocui"
)

type layoutDirection bool

const (
	layoutHorizontal layoutDirection = true
	layoutVertical   layoutDirection = false
)

type layoutItem struct {
	ratio   int
	fixed     int
	item    interface{}
	fNew    func(*gocui.View) error
	fUpdate func(*gocui.View) error
}

func (l *layoutItem) get() (string, *layoutLevel) {
	switch l.item.(type) {
	case string:
		return l.item.(string), nil
	case *layoutLevel:
		return "", l.item.(*layoutLevel)
	default:
		log.Fatalf("Bad item in layout (%T): %+v", l.item, l.item)
	}
	return "", nil
}

type layoutLevel struct {
	direction layoutDirection
	items     []*layoutItem
}

func (l *layoutLevel) layout(g *gocui.Gui, x0, y0, x1, y1 int) error {
	log.Printf("layout: (%d,%d)-(%d,%d)", x0, y0, x1, y1)
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
	log.Printf("%d segments of %d, %d left over", segments, unit, left)

	for idx, item := range l.items {
		var err error
		var assignment int
		name, layout := item.get()
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

		if layout != nil {
			log.Printf("Creating sublayout at (%d,%d)-(%d,%d)", ix0, iy0, ix1, iy1)
			err = layout.layout(g, ix0, iy0, ix1, iy1)
		} else {
			log.Printf("createView(%q at (%d,%d)-(%d,%d))", name, ix0, iy0, ix1, iy1)
			err = createView(g, name, ix0, iy0, ix1, iy1, item.fNew, item.fUpdate)
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
