package cli

import (
	"fmt"
	"log"
	"sort"
	"strings"

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
		{
			Section:    "Locations",
			Name:       "Distance",
			Usage:      "Distance <loc1> <loc2>",
			Validators: []string{"location", "location"},
			Help:       "Calculate the distance between two locations",
			Do:         doDistance,
			MinArgs:    2,
			MaxArgs:    2,
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

func getLocation(c *spacetraders.Client, loc string) (*spacetraders.Location, error) {
	loc = strings.ToUpper(loc)
	i := strings.Index(loc, "-")
	if i == -1 {
		return nil, fmt.Errorf("can't figure out system of %q", loc)
	}
	sysName := loc[:i]

	systems, err := c.ListSystems()
	if err != nil {
		return nil, fmt.Errorf("can't load systems: %v", err)
	}

	var sys *spacetraders.System
	for _, s := range systems {
		if s.Symbol == sysName {
			sys = &s
			break
		}
	}

	if sys == nil {
		return nil, fmt.Errorf("can't find system %q!", sysName)
	}

	for _, l := range sys.Locations {
		if l.Symbol == loc {
			return &l, nil
		}
	}

	return nil, fmt.Errorf("can't find location %q in %q!", loc, sysName)
}

func doDistance(c *spacetraders.Client, args []string) error {
	loc1, err := getLocation(c, args[0])
	if err != nil {
		return fmt.Errorf("can't find location %q: %v", args[0], err)
	}
	loc2, err := getLocation(c, args[1])
	if err != nil {
		return fmt.Errorf("can't find location %q: %v", args[1], err)
	}

	if loc1.SystemSymbol != loc2.SystemSymbol {
		return fmt.Errorf("locations must be in the same system, not %q and %q", loc1.SystemSymbol, loc2.SystemSymbol)
	}

	Out("Distance between %q and %q: %.2f", loc1.Symbol, loc2.Symbol, loc1.Distance(loc2))

	return nil
}
