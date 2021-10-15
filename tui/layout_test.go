package tui

import (
	"fmt"
	"testing"
	"time"

	"github.com/awesome-gocui/gocui"
)

type size struct {
	x0, y0, x1, y1 int
}

func (s *size) String() string {
	return fmt.Sprintf("(%d,%d)-(%d,%d)", s.x0, s.y0, s.x1, s.y1)
}

func TestLayout(t *testing.T) {
	tests := []struct {
		desc    string
		layout  *layoutLevel
		size    size
		want    map[string]size
		wantErr bool
	}{
		{
			desc: "simple",
			layout: &layoutLevel{
				direction: layoutHorizontal,
				items: []*layoutItem{
					{
						ratio: 1,
						name:  "test",
					},
				},
			},
			want: map[string]size{
				"test": {0, 0, 79, 24},
			},
		},
		{
			desc: "column",
			layout: &layoutLevel{
				direction: layoutVertical,
				items: []*layoutItem{
					{
						ratio: 1,
						name:  "test",
					},
					{
						ratio: 1,
						name:  "test2",
					},
					{
						ratio: 1,
						name:  "test3",
					},
				},
			},
			want: map[string]size{
				"test":  {0, 0, 79, 7},
				"test2": {0, 8, 79, 15},
				"test3": {0, 16, 79, 24},
			},
		},
		{
			desc: "row 2:1",
			layout: &layoutLevel{
				direction: layoutHorizontal,
				items: []*layoutItem{
					{
						ratio: 2,
						name:  "test",
					},
					{
						ratio: 1,
						name:  "test2",
					},
				},
			},
			want: map[string]size{
				"test":  {0, 0, 51, 24},
				"test2": {52, 0, 79, 24},
			},
		},
		{
			desc: "grid{11, 12; 21, 22, 23}",
			layout: &layoutLevel{
				direction: layoutHorizontal,
				items: []*layoutItem{
					{
						ratio: 1,
						inner: &layoutLevel{
							direction: layoutVertical,
							items: []*layoutItem{
								{
									ratio: 1,
									name:  "test11",
								},
								{
									ratio: 1,
									name:  "test12",
								},
							},
						},
					},
					{
						ratio: 1,
						inner: &layoutLevel{
							direction: layoutVertical,
							items: []*layoutItem{
								{
									ratio: 1,
									name:  "test21",
								},
								{
									ratio: 1,
									name:  "test22",
								},
								{
									ratio: 1,
									name:  "test23",
								},
							},
						},
					},
				},
			},
			want: map[string]size{
				"test11": {0, 0, 39, 11},
				"test12": {0, 12, 39, 24},
				"test21": {40, 0, 79, 7},
				"test22": {40, 8, 79, 15},
				"test23": {40, 16, 79, 24},
			},
		},
		{
			desc: "fixed",
			layout: &layoutLevel{
				direction: layoutHorizontal,
				items: []*layoutItem{
					{
						ratio: 1,
						name:  "test1",
						fixed: 10,
					},
					{
						ratio: 1,
						name:  "test2",
					},
					{
						ratio: 4,
						name:  "test3",
					},
				},
			},
			want: map[string]size{
				"test1": {0, 0, 9, 24},
				"test2": {10, 0, 23, 24},
				"test3": {24, 0, 79, 24},
			},
		},
	}

	g, err := gocui.NewGui(gocui.OutputSimulator, true)
	if err != nil {
		t.Fatalf("Can't create gui: %v", err)
	}
	testingScreen := g.GetTestingScreen()
	cleanup := testingScreen.StartGui()
	defer cleanup()
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			mgr := &ratioLayout{definition: tc.layout}
			g.SetManager(mgr)

			<-time.After(50 * time.Millisecond)
			found := make(map[string]bool)
			for _, v := range g.Views() {
				name := v.Name()
				found[name] = true
				if s, ok := tc.want[name]; !ok {
					x0, y0, x1, y1 := v.Dimensions()
					got := size{x0, y0, x1, y1}
					t.Errorf("Found unexpected view %q %s", name, got.String())
				} else {
					x0, y0, x1, y1 := v.Dimensions()
					got := size{x0, y0, x1, y1}
					if got != s {
						t.Errorf("Unexpected size for %q: got %s, want %s", name, got.String(), s.String())
					}
				}
			}

			for w := range tc.want {
				if !found[w] {
					t.Errorf("Expected view %q not found, views: %v", w, g.Views())
				}
			}
		})
	}
}
