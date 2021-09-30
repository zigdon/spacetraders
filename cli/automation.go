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
			Section: "Automation",
			Name:    "CreateTradeRoute",
			Usage:   "CreateTradeRoute <name> <location, cargo>...",
			Help: "Create a new trade route. Ships on the route will go to each " +
				"location in order, sell all their cargo, buy fuel for the next hop, " +
				"and fill the remaining space with the specified cargo.",
			Do:      doCreateTradeRoute,
			MinArgs: 3,
			MaxArgs: -1,
			Aliases: []string{"NewTrade", "NewRoute"},
		},
		{
			Section: "Automation",
			Name:    "ShowTradeRoute",
			Usage:   "ShowTradeRoute [name]",
			Help:    "Show the named trade route, or list all routes if none specified.",
			Do:      doShowTradeRoute,
			MinArgs: 0,
			MaxArgs: 1,
			Aliases: []string{"ShowRoute"},
		},
		{
			Section:    "Automation",
			Name:       "AddShipToRoute",
			Usage:      "AddShipToRoute <route name> <ship id>",
			Validators: []string{"", "ships"},
			Help:       "Add a new ship to an existing trade route.",
			Do:         doAddShipToRoute,
			MinArgs:    2,
			MaxArgs:    2,
		},
		{
			Section: "Automation",
			Name:    "ProcessRoutes",
			Usage:   "ProcessRoutes",
			Help:    "Manually trigger routes processing, for debugging only.",
			Do:      doProcessRoutes,
			MinArgs: 0,
			MaxArgs: 0,
		},
	} {
		if err := Register(c); err != nil {
			log.Fatalf("Can't register %q: %v", c.Name, err)
		}
	}
}

type route struct {
	c            *spacetraders.Client
	name         string
	ships        map[string]int
	destinations []string
	cargos       []string
	autoFuel     bool
	balance      int
	logEntries   []string
}

var routes = make(map[string]*route)

// Take a list of [location, good], create a trade route
func doCreateTradeRoute(c *spacetraders.Client, args []string) error {
	name := args[0]
	if r, ok := routes[strings.ToLower(name)]; ok {
		return fmt.Errorf("a route already exists named %q: %s", name, r.Short())
	}
	log.Printf("Creating route %q", name)
	r := &route{
		c:            c,
		name:         name,
		autoFuel:     true,
		destinations: []string{},
		cargos:       []string{},
		ships:        make(map[string]int),
		logEntries:   []string{},
	}
	pairs := args[1:]

	for len(pairs) > 0 {
		if len(pairs) < 2 {
			return fmt.Errorf("Arguments must match: <name> <location, cargo>...; Got %q", args)
		}
		if err := validate(c, pairs[:2], []string{"location", "cargo"}); err != nil {
			return fmt.Errorf("invalid pair %v for route: %v", pairs[:2], err)
		}
		r.destinations = append(r.destinations, pairs[0])
		r.cargos = append(r.cargos, pairs[1])
		pairs = pairs[2:]
	}

	routes[strings.ToLower(name)] = r
	Out("Created route:\n%s", r.String())

	return nil
}

func doShowTradeRoute(c *spacetraders.Client, args []string) error {
	var rs []string
	for _, r := range routes {
		rs = append(rs, r.name)
	}
	sort.Strings(rs)

	if len(args) == 0 {
		Out("Known routes: %s", strings.Join(rs, ", "))
		return nil
	}

	r, ok := routes[strings.ToLower(args[0])]
	if !ok {
		return fmt.Errorf("Can't find route %q. Known routes: %s", args[0], strings.Join(rs, ", "))
	}

	Out(r.String())
	return nil
}

func doAddShipToRoute(c *spacetraders.Client, args []string) error {
	name := args[0]
	r, ok := routes[strings.ToLower(name)]
	if !ok {
		rs := []string{}
		for _, r := range routes {
			rs = append(rs, r.name)
		}
		sort.Strings(rs)
		return fmt.Errorf("can't find route %q. Available routes: %s", name, strings.Join(rs, ", "))
	}

	ship, err := getShip(c, args[1])
	if err != nil {
		return fmt.Errorf("can't find ship %q: %v", args[1], err)
	}

	if _, ok := r.ships[ship.ID]; ok {
		return fmt.Errorf("ship %s is already on route %s.", ship.ShortID, r.name)
	}

	r.Log("%s: Adding ship %q to route", ship.ShortID, ship.ID)
	r.ships[ship.ID] = 0
	if ship.LocationName != r.destinations[0] {
		fp, err := c.CreateFlight(ship.ID, r.destinations[0])
		if err != nil {
			return fmt.Errorf("can't send %q to %q: %v", ship.ShortID, r.destinations[0], err)
		}
		r.Log("%s: Created flight plan to %q: %s (%s)", ship.ShortID, r.destinations[0], fp.ShortID, fp.ID)
	}

	return nil
}

func doProcessRoutes(c *spacetraders.Client, args []string) error {
	for k, r := range routes {
		Out("Running tasks for route %s...", k)
		if err := r.HandlePending(); err != nil {
			return fmt.Errorf("error handling route %s: %v", k, err)
		}
	}

	return nil
}

func (r *route) Log(format string, args ...interface{}) {
	r.logEntries = append(r.logEntries, fmt.Sprintf(format, args...))
	if len(r.logEntries) > 50 {
		purge := len(r.logEntries) - 50
		r.logEntries = r.logEntries[purge:]
	}
}

func (r *route) String() string {
	res := []string{fmt.Sprintf("Route %q  Profit: %d", r.name, r.balance), "Locations:"}
	for i := range r.destinations {
		res = append(res, fmt.Sprintf("  %s: %s", r.destinations[i], r.cargos[i]))
	}
	if len(r.ships) > 0 {
		res = append(res, "Ships:")
		for s, i := range r.ships {
			res = append(res, fmt.Sprintf("  %s: -> %s", s, r.destinations[i]))
		}
	}
	if len(r.logEntries) > 0 {
		res = append(res, "Recent log entries:")
		for _, l := range r.logEntries {
			res = append(res, fmt.Sprintf("  %s", l))
		}
	}

	return strings.Join(res, "\n")
}

func (r *route) Short() string {
	return fmt.Sprintf("%s: %d ships, trading: %v, locations: %v, profit: %d", r.name, len(r.ships), r.cargos, r.destinations, r.balance)
}

func (r *route) AddShip(ship string) error {
	if i, ok := r.ships[ship]; ok {
		return fmt.Errorf("%s is already using this route, heading to %s", ship, r.destinations[i])
	}
	r.ships[ship] = 0

	return nil
}

func (r *route) DelShip(ship string) error {
	if _, ok := r.ships[ship]; !ok {
		return fmt.Errorf("%s isn't using this route.", ship)
	}
	delete(r.ships, ship)

	return nil
}

func (r *route) HandlePending() error {
	for s, i := range r.ships {
		ship, err := getShip(r.c, s)
		if err != nil {
			return fmt.Errorf("can't find ship %q for route %q: %v", s, r.name, err)
		}

		// Ship is still in flight
		if ship.FlightPlanID != "" {
			continue
		}

		var expectedLocation string
		expectedLocation = r.destinations[i]
		if ship.LocationName != expectedLocation {
			return fmt.Errorf("%s isn't at the expected location for %s[%d]: at %q rather than %q",
				ship.ShortID, r.name, i, ship.LocationName, expectedLocation)
		}

		// Sell cargo
		if err := r.SellAll(ship); err != nil {
			return fmt.Errorf("can't sell cargo from %q at %q: %v", ship.ShortID, ship.LocationName, err)
		}

		// Buy fuel for the next hop
		ship, err = getShip(r.c, s)
		if err != nil {
			return fmt.Errorf("can't find ship %q for route %q: %v", s, r.name, err)
		}
		if err := r.BuyFuel(ship, r.destinations[i]); err != nil {
			return fmt.Errorf("can't buy fuel for %q: %v", ship.ShortID, err)
		}

		// Buy cargo
		if r.cargos[i] == "NONE" {
			r.Log("%s: Not buying cargo at %s", ship.ShortID, ship.LocationName)
		} else {
			ship, err = getShip(r.c, s)
			if err != nil {
				return fmt.Errorf("can't find ship %q for route %q: %v", s, r.name, err)
			}
			if err := r.BuyCargo(ship, r.cargos[i]); err != nil {
				return fmt.Errorf("can't buy cargo %s for %q: %v", r.cargos[i], ship.ShortID, err)
			}
		}

		// Fly on
		var nextDest string
		nextDest = r.destinations[(i+1)%len(r.destinations)]
		fp, err := r.c.CreateFlight(ship.ID, nextDest)
		if err != nil {
			return fmt.Errorf("can't send %s to %s: %v", ship.ShortID, nextDest, err)
		}
		r.ships[s] = (i + 1) % len(r.destinations)
		r.Log("%s: Created flight plan %s to %s", ship.ShortID, fp.ShortID, nextDest)
	}
	return nil
}

func (r *route) BuyAll(ship *spacetraders.Ship, cargo string) error {
	return nil
}

type bsType bool

var (
	bsBuy  bsType = true
	bsSell bsType = false
)

func (r *route) BuySell(ship *spacetraders.Ship, bs bsType, good string, qty int) error {
	var f func(string, string, int) (*spacetraders.Order, error)
	var verb string
	if bs == bsSell {
		f = r.c.SellCargo
		verb = "selling"
	} else {
		f = r.c.BuyCargo
		verb = "buying"
	}
	for qty > 0 {
		sell := qty
		if sell > ship.LoadingSpeed {
			sell = ship.LoadingSpeed
		}
		o, err := f(ship.ID, good, sell)
		if err != nil {
			return fmt.Errorf("%s: error %s %d %q: %v", ship.ShortID, verb, sell, good, err)
		}
		qty -= sell
		r.balance += o.Total
	}
	return nil
}

func (r *route) SellAll(ship *spacetraders.Ship) error {
	for _, c := range ship.Cargo {
		if c.Good == "FUEL" {
			continue
		}
		qty := c.Quantity
		r.Log("%s: selling %d %s at %s", ship.ShortID, qty, c.Good, ship.LocationName)
		if err := r.BuySell(ship, bsSell, c.Good, qty); err != nil {
			return err
		}
	}
	return nil
}

func (r *route) BuyFuel(ship *spacetraders.Ship, dest string) error {
	curFuel := ship.Cargo[0]
	if curFuel.Good != "FUEL" {
		return fmt.Errorf("%q had some cargo that isn't fuel (%q)", ship.ShortID, curFuel.Good)
	}

	curLoc, err := getLocation(r.c, ship.LocationName)
	if err != nil {
		return fmt.Errorf("error getting current location %q for %q: %v", ship.LocationName, ship.ShortID, err)
	}
	nextLoc, err := getLocation(r.c, dest)
	if err != nil {
		return fmt.Errorf("error getting next location %q for %q: %v", dest, ship.ShortID, err)
	}

	fuelNeeded := ship.FuelNeeded(curLoc, nextLoc) - curFuel.Quantity
	if fuelNeeded <= 0 {
		r.Log("%s: Already have enoughd fuel for trip to %s", ship.ShortID, dest)
		return nil
	}

	r.Log("%s: Buying %d fuel for trip to %s", ship.ShortID, fuelNeeded, dest)
	return r.BuySell(ship, bsBuy, "FUEL", fuelNeeded)
}

func (r *route) BuyCargo(ship *spacetraders.Ship, cargo string) error {
	market, err := r.c.Marketplace(ship.LocationName)
	if err != nil {
		return fmt.Errorf("can't check market at %q: %v", ship.LocationName, err)
	}
	var offer *spacetraders.Offer
	for _, o := range market {
		if o.Symbol != cargo {
			continue
		}
		offer = &o
		break
	}
	if offer == nil {
		return fmt.Errorf("can't find %q on offer at %q", cargo, ship.LocationName)
	}

	buy := ship.SpaceAvailable / offer.VolumePerUnit
	if buy > offer.QuantityAvailable {
		buy = offer.QuantityAvailable
		r.Log("%s: Only %d of %s available at %s", ship.ShortID, buy, cargo, ship.LocationName)
	}

	acc, err := r.c.Account()
	if err != nil {
		return fmt.Errorf("can't get account info: %v", err)
	}
	if buy*offer.PricePerUnit > acc.Credits {
		return fmt.Errorf("can't afford to buy %d of %s at %s, only %d credits available.",
			buy, cargo, ship.LocationName, acc.Credits)
	}
	r.Log("%s: Buying %d of %s at %s", ship.ShortID, buy, cargo, ship.LocationName)
	return r.BuySell(ship, bsBuy, cargo, buy)
}
