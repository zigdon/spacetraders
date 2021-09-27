package cli

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/zigdon/spacetraders"
)

// Command implementations
func doAccount(c *spacetraders.Client, args []string) error {
	u, err := c.Account()
	if err != nil {
		return err
	}
	Out("%s", u)
	return nil
}

func doLogin(c *spacetraders.Client, args []string) error {
	path := filepath.Join(os.Getenv("HOME"), ".config/spacetraders.io")
	if len(args) > 0 {
		path = args[0]
	}
	if err := c.Load(path); err != nil {
		ErrMsg("Error loading token: %v", err)
	}

	return nil
}

func doClaim(c *spacetraders.Client, args []string) error {
	username := args[0]
	path := args[1]
	if _, err := os.Stat(args[1]); err == nil {
		return fmt.Errorf("%q already exists, aborting.", path)
	}

	token, _, err := c.Claim(username)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(
		path,
		[]byte(fmt.Sprintf("%s\n%s\n", username, token)),
		0600); err != nil {
		return fmt.Errorf("Error writing new token %q to %q: %v", token, path, err)
	}
	log.Printf("Got token %q for %q", token, username)

	return nil
}

func doLogout(c *spacetraders.Client, args []string) error {
	return c.Logout()
}

func doLoans(c *spacetraders.Client, args []string) error {
	loans, err := c.AvailableLoans()
	if err != nil {
		return fmt.Errorf("error getting loans: %v", err)
	}

	for _, l := range loans {
		Out("amt: %d, needs collateral: %v, rate: %d, term (days): %d, type: %s",
			l.Amount, l.CollateralRequired, l.Rate, l.TermInDays, l.Type)
	}

	return nil
}

func doTakeLoan(c *spacetraders.Client, args []string) error {
	loan, err := c.TakeLoan(args[0])
	if err != nil {
		return fmt.Errorf("error taking out loan: %v", err)
	}

	Out("Loan taken, %s (%s), due: %s (in %s)",
		loan.ShortID, loan.ID, loan.Due.Local(), loan.Due.Sub(time.Now()).Truncate(time.Second))

	return nil
}

func doMyLoans(c *spacetraders.Client, args []string) error {
	loans, err := c.MyLoans()
	if err != nil {
		return fmt.Errorf("error querying loans: %v", err)
	}

	for _, l := range loans {
		Out(l.String())
	}

	return nil
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

func doListShips(c *spacetraders.Client, args []string) error {
	ships, err := c.ListShips(args[0])
	if err != nil {
		return fmt.Errorf("error listing ships in %q: %v", args[0], err)
	}

	if len(args) > 1 {
		for _, s := range ships {
			if !s.Filter(args[1]) {
				continue
			}
			Out(s.String())
		}
		return nil
	}

	for _, s := range ships {
		Out(s.String())
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

func doBuy(c *spacetraders.Client, args []string) error {
	qty, err := strconv.Atoi(args[2])
	if err != nil {
		return err
	}

	order, err := c.BuyCargo(args[0], args[1], qty)
	if err != nil {
		return fmt.Errorf("error selling goods: %v", err)
	}

	Out("Bought %d of %s for %d", order.Quantity, order.Good, order.Total)

	return nil
}

func doSell(c *spacetraders.Client, args []string) error {
	qty, err := strconv.Atoi(args[2])
	if err != nil {
		return err
	}

	order, err := c.SellCargo(args[0], args[1], qty)
	if err != nil {
		return fmt.Errorf("error selling goods: %v", err)
	}

	Out("Sold %d of %s for %d", order.Quantity, order.Good, order.Total)

	return nil
}

func doMarket(c *spacetraders.Client, args []string) error {
	offers, err := c.Marketplace(args[0])
	if err != nil {
		return fmt.Errorf("error querying marketplace at %q: %v", args[0], err)
	}

	Out("%d offers at %q:", len(offers), args[0])
	for _, offer := range offers {
		Out(offer.String())
	}

	return nil
}
