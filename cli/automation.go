package cli

import (
	"encoding/json"
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
	} {
		if err := Register(c); err != nil {
			log.Fatalf("Can't register %q: %v", c.Name, err)
		}
	}

	if err := RegisterPersistence("automation", save, load); err != nil {
		log.Fatalf("Can't register load/save for automation: %v", err)
	}
}

type route struct {
	Name         string
	Ships        map[string]int
	Destinations []string
	Cargos       []string
	AutoFuel     bool
	Balance      int
	LogEntries   []string
}

type saveData struct {
	Routes map[string]*route `json:"routes"`
}

func save() string {
	out := strings.Builder{}
	enc := json.NewEncoder(&out)

	mySave := saveData{Routes: routes}

	if err := enc.Encode(mySave); err != nil {
		log.Fatalf("error saving automation %#v: %v", mySave, err)
	}

	return out.String()
}

func load(data string) error {
	dec := json.NewDecoder(strings.NewReader(data))
	dec.DisallowUnknownFields()

	saved := &saveData{}
	if err := dec.Decode(saved); err != nil {
		return fmt.Errorf("error decoding json into %#v: %v\n%s", saved, err, data)
	}
	routes = saved.Routes

	return nil
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
		Name:         name,
		AutoFuel:     true,
		Destinations: []string{},
		Cargos:       []string{},
		Ships:        make(map[string]int),
		LogEntries:   []string{},
	}
	pairs := args[1:]

	for len(pairs) > 0 {
		if len(pairs) < 2 {
			return fmt.Errorf("Arguments must match: <name> <location, cargo>...; Got %q", args)
		}
		if err := validate(c, pairs[:2], []string{"location", "cargo"}); err != nil {
			return fmt.Errorf("invalid pair %v for route: %v", pairs[:2], err)
		}
		r.Destinations = append(r.Destinations, pairs[0])
		r.Cargos = append(r.Cargos, pairs[1])
		pairs = pairs[2:]
	}

	routes[strings.ToLower(name)] = r
	Out("Created route:\n%s", r.String())

	return nil
}

func doShowTradeRoute(c *spacetraders.Client, args []string) error {
	var rs []string
	for _, r := range routes {
		rs = append(rs, r.Name)
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
			rs = append(rs, r.Name)
		}
		sort.Strings(rs)
		return fmt.Errorf("can't find route %q. Available routes: %s", name, strings.Join(rs, ", "))
	}

	ship, err := getShip(c, args[1])
	if err != nil {
		return fmt.Errorf("can't find ship %q: %v", args[1], err)
	}

	if _, ok := r.Ships[ship.ID]; ok {
		return fmt.Errorf("ship %s is already on route %s.", ship.ShortID, r.Name)
	}

	r.Log("%s: Adding ship %q to route", ship.ShortID, ship.ID)
	r.Ships[ship.ID] = 0
	if ship.LocationName != r.Destinations[0] {
		fp, err := c.CreateFlight(ship.ID, r.Destinations[0])
		if err != nil {
			return fmt.Errorf("can't send %q to %q: %v", ship.ShortID, r.Destinations[0], err)
		}
		r.Log("%s: Created flight plan to %q: %s (%s)", ship.ShortID, r.Destinations[0], fp.ShortID, fp.ID)
	}

	return nil
}

func ProcessRoutes(c *spacetraders.Client) error {
	for k, r := range routes {
		if err := r.HandlePending(c); err != nil {
			return fmt.Errorf("error handling route %s: %v", k, err)
		}
	}

	return nil
}

func (r *route) Log(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	ui.Msg("%s: %s", r.Name, msg)
	r.LogEntries = append(r.LogEntries, msg)
	if len(r.LogEntries) > 10 {
		purge := len(r.LogEntries) - 10
		r.LogEntries = r.LogEntries[purge:]
	}
}

func (r *route) String() string {
	res := []string{fmt.Sprintf("Route %q  Profit: %d", r.Name, r.Balance), "Locations:"}
	for i := range r.Destinations {
		res = append(res, fmt.Sprintf("  %s: %s", r.Destinations[i], r.Cargos[i]))
	}
	if len(r.Ships) > 0 {
		res = append(res, "Ships:")
		for s, i := range r.Ships {
			res = append(res, fmt.Sprintf("  %s: -> %s", s, r.Destinations[i]))
		}
	}
	if len(r.LogEntries) > 0 {
		res = append(res, "Recent log entries:")
		for _, l := range r.LogEntries {
			res = append(res, fmt.Sprintf("  %s", l))
		}
	}

	return strings.Join(res, "\n")
}

func (r *route) Short() string {
	return fmt.Sprintf("%s: %d ships, trading: %v, locations: %v, profit: %d", r.Name, len(r.Ships), r.Cargos, r.Destinations, r.Balance)
}

func (r *route) AddShip(ship string) error {
	if i, ok := r.Ships[ship]; ok {
		return fmt.Errorf("%s is already using this route, heading to %s", ship, r.Destinations[i])
	}
	r.Ships[ship] = 0

	return nil
}

func (r *route) DelShip(ship string) error {
	if _, ok := r.Ships[ship]; !ok {
		return fmt.Errorf("%s isn't using this route.", ship)
	}
	delete(r.Ships, ship)

	return nil
}

func (r *route) HandlePending(c *spacetraders.Client) error {
	for s, i := range r.Ships {
		ship, err := getShip(c, s)
		if err != nil {
			return fmt.Errorf("can't find ship %q for route %q: %v", s, r.Name, err)
		}

		// Ship is still in flight
		if ship.FlightPlanID != "" {
			continue
		}

		var expectedLocation string
		expectedLocation = r.Destinations[i]
		if ship.LocationName != expectedLocation {
			return fmt.Errorf("%s isn't at the expected location for %s[%d]: at %q rather than %q",
				ship.ShortID, r.Name, i, ship.LocationName, expectedLocation)
		}

		// Sell cargo
		if err := r.SellAll(c, ship); err != nil {
			return fmt.Errorf("can't sell cargo from %q at %q: %v", ship.ShortID, ship.LocationName, err)
		}

		// Buy fuel for the next hop
		ship, err = getShip(c, s)
		if err != nil {
			return fmt.Errorf("can't find ship %q for route %q: %v", s, r.Name, err)
		}
		if err := r.BuyFuel(c, ship, r.Destinations[i]); err != nil {
			return fmt.Errorf("can't buy fuel for %q: %v", ship.ShortID, err)
		}

		// Buy cargo
		if r.Cargos[i] == "NONE" {
			r.Log("%s: Not buying cargo at %s", ship.ShortID, ship.LocationName)
		} else {
			ship, err = getShip(c, s)
			if err != nil {
				return fmt.Errorf("can't find ship %q for route %q: %v", s, r.Name, err)
			}
			if err := r.BuyCargo(c, ship, r.Cargos[i]); err != nil {
				return fmt.Errorf("can't buy cargo %s for %q: %v", r.Cargos[i], ship.ShortID, err)
			}
		}

		// Fly on
		var nextDest string
		nextDest = r.Destinations[(i+1)%len(r.Destinations)]
		fp, err := c.CreateFlight(ship.ID, nextDest)
		if err != nil {
			return fmt.Errorf("can't send %s to %s: %v", ship.ShortID, nextDest, err)
		}
		r.Ships[s] = (i + 1) % len(r.Destinations)
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

func (r *route) BuySell(c *spacetraders.Client, ship *spacetraders.Ship, bs bsType, good string, qty int) error {
	var f func(string, string, int) (*spacetraders.Order, error)
	var verb string
	if bs == bsSell {
		f = c.SellCargo
		verb = "selling"
	} else {
		f = c.BuyCargo
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
		r.Balance += o.Total
	}
	return nil
}

func (r *route) SellAll(c *spacetraders.Client, ship *spacetraders.Ship) error {
	for _, g := range ship.Cargo {
		if g.Good == "FUEL" {
			continue
		}
		qty := g.Quantity
		r.Log("%s: selling %d %s at %s", ship.ShortID, qty, g.Good, ship.LocationName)
		if err := r.BuySell(c, ship, bsSell, g.Good, qty); err != nil {
			return err
		}
	}
	return nil
}

func (r *route) BuyFuel(c *spacetraders.Client, ship *spacetraders.Ship, dest string) error {
	curFuel := 0
	if len(ship.Cargo) > 0 {
		good := ship.Cargo[0]
		if good.Good != "FUEL" {
			return fmt.Errorf("%q had some cargo that isn't fuel (%q)", ship.ShortID, good.Good)
		}
		curFuel = good.Quantity
	}

	curLoc, err := getLocation(c, ship.LocationName)
	if err != nil {
		return fmt.Errorf("error getting current location %q for %q: %v", ship.LocationName, ship.ShortID, err)
	}
	nextLoc, err := getLocation(c, dest)
	if err != nil {
		return fmt.Errorf("error getting next location %q for %q: %v", dest, ship.ShortID, err)
	}

	fuelNeeded := ship.FuelNeeded(curLoc, nextLoc) - curFuel
	if fuelNeeded <= 0 {
		r.Log("%s: Already have enoughd fuel for trip to %s", ship.ShortID, dest)
		return nil
	}

	r.Log("%s: Buying %d fuel for trip to %s", ship.ShortID, fuelNeeded, dest)
	return r.BuySell(c, ship, bsBuy, "FUEL", fuelNeeded)
}

func (r *route) BuyCargo(c *spacetraders.Client, ship *spacetraders.Ship, cargo string) error {
	market, err := c.Marketplace(ship.LocationName)
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

	acc, err := c.Account()
	if err != nil {
		return fmt.Errorf("can't get account info: %v", err)
	}
	if buy*offer.PricePerUnit > acc.Credits {
		return fmt.Errorf("can't afford to buy %d of %s at %s, only %d credits available.",
			buy, cargo, ship.LocationName, acc.Credits)
	}
	r.Log("%s: Buying %d of %s at %s", ship.ShortID, buy, cargo, ship.LocationName)
	return r.BuySell(c, ship, bsBuy, cargo, buy)
}
