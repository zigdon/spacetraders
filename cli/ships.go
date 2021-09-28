package cli

import (
	"fmt"
	"log"
	"time"

	"github.com/zigdon/spacetraders"
)

func init() {
	for _, c := range []cmd{
		{
			Section:    "Ships",
			Name:       "ListShips",
			Usage:      "ListShips <system> [filter]",
			Validators: []string{"system"},
			Help: "Show available ships at all the locations in a system. If filter is provided, " +
				"only show ships that match in type, manufacturer, or class",
			Do:      doListShips,
			MinArgs: 1,
			MaxArgs: 2,
		},
		{
			Section:    "Ships",
			Name:       "BuyShip",
			Usage:      "BuyShip <location> <type>",
			Validators: []string{"location"},
			Help:       "Buy the given ship in the specified location",
			Do:         doBuyShip,
			MinArgs:    2,
			MaxArgs:    2,
		},
		{
			Section:    "Ships",
			Name:       "MyShips",
			Usage:      "MyShips [filter]",
			Validators: []string{"ships"},
			Help:       "List owned ships, with an optional filter",
			Do:         doMyShips,
			MinArgs:    0,
			MaxArgs:    1,
			Aliases:    []string{"lsShips"},
		},
		{
			Section:    "Flight Plans",
			Name:       "CreateFlightPlan",
			Usage:      "CreateFlightPlan <shipID> <destination>",
			Validators: []string{"ships", "location"},
			Help:       "Create a flight plan for given ship to specified destination",
			Do:         doCreateFlight,
			MinArgs:    2,
			MaxArgs:    2,
			Aliases:    []string{"go", "fly"},
		},
		{
			Section:    "Flight Plans",
			Name:       "ShowFlightPlan",
			Usage:      "ShowFlightPlan <flightPlanID>",
			Validators: []string{"flights"},
			Help:       "Show the flight plan identified",
			Do:         doShowFlight,
			MinArgs:    1,
			MaxArgs:    1,
			Aliases:    []string{"lsFlights"},
		},
		{
			Section:    "Flight Plans",
			Name:       "Wait",
			Usage:      "Wait <flightPlanID>",
			Validators: []string{"flights"},
			Help:       "Wait until specified flight arrives",
			Do:         doWaitForFlight,
			MinArgs:    1,
			MaxArgs:    1,
		},
	} {
		if err := Register(c); err != nil {
			log.Fatalf("Can't register %q: %v", c.Name, err)
		}
	}
}

func doListShips(c *spacetraders.Client, args []string) error {
	ships, err := c.ListShips(args[0])
	if err != nil {
		return fmt.Errorf("error listing ships in %q: %v", args[0], err)
	}

	for _, s := range ships {
		if len(args) > 1 && !s.Filter(args[1]) {
			continue
		}
		Out(s.Listing())
	}

	return nil
}

func doBuyShip(c *spacetraders.Client, args []string) error {
	ship, err := c.BuyShip(args[0], args[1])
	if err != nil {
		return fmt.Errorf("error buying ship %q at %q: %v", args[1], args[0], err)
	}

	Out("New ship ID: %s (%s)", ship.ShortID, ship.ID)

	return nil
}

func doMyShips(c *spacetraders.Client, args []string) error {
	ships, err := c.MyShips()
	if err != nil {
		return fmt.Errorf("error listing my ships: %v", err)
	}

	res := []spacetraders.Ship{}
	for _, s := range ships {
		if len(args) > 0 {
			if !s.Filter(args[0]) {
				continue
			}
		}
		res = append(res, s)
	}

	switch len(res) {
	case 0:
		Out("No ships found.")
	case 1:
		Out(res[0].String())
	default:
		for _, s := range res {
			Out(s.Short())
		}
	}

	return nil
}

func doCreateFlight(c *spacetraders.Client, args []string) error {
	flight, err := c.CreateFlight(args[0], args[1])
	if err != nil {
		return fmt.Errorf("error creating flight plan to %q: %v", args[1], err)
	}

	Out("Created flight plan: %s", flight.Short())
	tq.Add(
		flight.ShortID,
		fmt.Sprintf("%s: %s arrived at %s", flight.ShortID, flight.ShortShipID, flight.Destination),
		nil,
		flight.ArrivesAt)

	return nil
}

func doShowFlight(c *spacetraders.Client, args []string) error {
	flight, err := c.ShowFlight(args[0])
	if err != nil {
		return fmt.Errorf("error listing flight plan %q: %v", args[0], err)
	}

	Out(flight.String())

	return nil
}

func doWaitForFlight(c *spacetraders.Client, args []string) error {
	flight, err := c.ShowFlight(args[0])
	if err != nil {
		return fmt.Errorf("error listing flight plan %q: %v", args[0], err)
	}

	if flight.ArrivesAt.Before(time.Now()) {
		return fmt.Errorf("flight %s (%s) already arrived", flight.ShortID, flight.ID)
	}

	// TODO: Allow interrupting the wait
	delay := flight.ArrivesAt.Sub(time.Now()).Truncate(time.Second)
	Out("Waiting %s for %s (%s) to arrive...", delay, flight.ShortID, flight.ID)
	select {
	case <-time.After(delay):
	case <-time.After(time.Minute):
		Out("... still waiting for %s", flight.ShortID)
	}
	Out("... %s arrived!", flight.ShortID)

	return nil
}
