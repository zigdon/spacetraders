package cli

import (
	"fmt"
	"log"
	"sort"

	"github.com/zigdon/spacetraders"
)

func init() {
	for _, c := range []cmd{
		{
			Section:    "Locations",
			Name:       "System",
			Usage:      "System [system]",
			Validators: []string{"system"},
			Help:       "Get details about a system, or all systems if not specified",
			Do:         doListSystems,
			MaxArgs:    1,
			Aliases:    []string{"lsSys"},
		},
		{
			Section:    "Locations",
			Name:       "Locations",
			Usage:      "Locations <system> [type]",
			Validators: []string{"system"},
			Help:       "Show all locations in a system",
			Do:         doListLocations,
			MinArgs:    1,
			MaxArgs:    2,
			Aliases:    []string{"lsLocations", "lsLocs"},
		},
	} {
		if err := Register(c); err != nil {
			log.Fatalf("Can't register %q: %v", c.Name, err)
		}
	}
}

func doListSystems(c *spacetraders.Client, args []string) error {
	systems, err := c.ListSystems()
	if err != nil {
		return fmt.Errorf("error listing systems: %v", err)
	}

	sys := []string{}
	cache := make(map[string]spacetraders.System)
	for _, s := range systems {
		sys = append(sys, s.Symbol)
		cache[s.Symbol] = s
	}
	sort.Strings(sys)
	if len(args) == 0 {
		Out("All systems:")
		for _, sym := range sys {
			Out(cache[sym].String())
		}
		return nil
	}

	Out(cache[args[0]].Details(0))
	return nil
}

func doListLocations(c *spacetraders.Client, args []string) error {
	filter := ""
	if len(args) > 1 {
		filter = args[1]
	}
	locs, err := c.ListLocations(args[0], filter)
	if err != nil {
		return fmt.Errorf("error listing locations in %q: %v", args[0], err)
	}

	Out("%d locations in %q:", len(locs), args[0])
	for _, l := range locs {
		Out(l.Details(1))
	}

	return nil
}
