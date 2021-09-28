package cli

import (
	"fmt"
	"log"
	"strconv"

	"github.com/zigdon/spacetraders"
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

func doBuy(c *spacetraders.Client, args []string) error {
	qty, err := strconv.Atoi(args[2])
	if err != nil {
		return err
	}

	order, err := c.BuyCargo(args[0], args[1], qty)
	if err != nil {
		return fmt.Errorf("error buying goods: %v", err)
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
