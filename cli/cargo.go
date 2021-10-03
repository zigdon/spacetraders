package cli

import (
	"fmt"
	"log"
	"strconv"

	"github.com/zigdon/spacetraders"
	"github.com/zigdon/spacetraders/tasks"
)

func init() {
	for _, c := range []cmd{
		{
			Section:    "Goods and Cargo",
			Name:       "Buy",
			Usage:      "Buy <shipID> <good> <quantity>",
			Validators: []string{"ships"},
			Help:       "Buy the specified quantiy of good for the ship identified",
			Do:         doBuy,
			MinArgs:    3,
			MaxArgs:    3,
		},
		{
			Section:    "Goods and Cargo",
			Name:       "Sell",
			Usage:      "Sell <shipID> <good> <quantity>",
			Validators: []string{"ships"},
			Help:       "Sell the specified quantiy of good from the ship identified",
			Do:         doSell,
			MinArgs:    3,
			MaxArgs:    3,
		},
		{
			Section:    "Goods and Cargo",
			Name:       "Market",
			Usage:      "Market <location>",
			Validators: []string{"mylocation"},
			Help:       "List all goods offered at location.",
			Do:         doMarket,
			MinArgs:    1,
			MaxArgs:    1,
		}} {
		if err := Register(c); err != nil {
			log.Fatalf("Error registering %q: %v", c.Name, err)
		}
	}
}

type commerceType bool

var (
	commerceBuy  commerceType = false
	commerceSell commerceType = true
)

func doBuy(c *spacetraders.Client, args []string) error {
	return cargoTx(c, args, commerceBuy)
}

func doSell(c *spacetraders.Client, args []string) error {
	return cargoTx(c, args, commerceSell)
}

func cargoTx(c *spacetraders.Client, args []string, kind commerceType) error {
	qty, err := strconv.Atoi(args[2])
	if err != nil {
		return err
	}

	shipName := args[0]
	ship, err := getShip(c, shipName)
	if err != nil {
		return fmt.Errorf("unknown ship %q: %v", shipName, err)
	}

	var doing, done string
	var f func(string, string, int) (*spacetraders.Order, error)
	if kind == commerceBuy {
		doing = "buying"
		done = "bought"
		f = c.BuyCargo
	} else if kind == commerceSell {
		doing = "selling"
		done = "sold"
		f = c.SellCargo
	}

	var total, handled int
	for handled < qty {
		left := qty - handled
		if left > ship.LoadingSpeed {
			left = ship.LoadingSpeed
		}
		order, err := f(shipName, args[1], left)
		if err != nil {
			return fmt.Errorf("error %s %d goods, only %s %d: %v", doing, left, done, handled, err)
		}
		handled += order.Quantity
		total += order.Total
	}

	Out("%s %s %d of %s for %d", ship.ShortID, done, handled, args[1], total)
	tasks.GetTaskQueue().Run("updateShips")

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
