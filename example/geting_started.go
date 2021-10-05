package main

// Implement the steps from the getting started guide

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/zigdon/spacetraders"
)

func main() {
	c := spacetraders.New()
	flag.Set("debug", "true")
	if err := c.Status(); err != nil {
		log.Fatal(err)
	}

	log.Print("*** Generate An Access Token")
	rand.Seed(int64(time.Now().Nanosecond()))
	username := fmt.Sprintf("example_%d", rand.Int31())
	_, _, err := c.Claim(username)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("*** View Your User Account")
	_, err = c.Account()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("*** View Available Loans")
	_, err = c.AvailableLoans()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("*** Take Out A Loan")
	_, err = c.TakeLoan("STARTUP")
	if err != nil {
		log.Fatal(err)
	}

	log.Print("*** View Ships To Purchase")
	_, err = c.ListShips("OE")
	if err != nil {
		log.Fatal(err)
	}

	log.Print("*** Purchase A Ship")
	ship, err := c.BuyShip("OE-PM-TR", "JW-MK-I")
	if err != nil {
		log.Fatal(err)
	}

	log.Print("*** Purchase Ship Fuel")
	_, err = c.BuyCargo(ship.ID, "FUEL", 20)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("*** View Marketplace")
	_, err = c.Marketplace("OE-PM-TR")
	if err != nil {
		log.Fatal(err)
	}

	log.Print("*** Buy Cargo")
	_, err = c.BuyCargo(ship.ID, "METALS", 25)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("*** Find Nearby Planet")
	_, err = c.ListLocations("OE", "PLANET")
	if err != nil {
		log.Fatal(err)
	}

	log.Print("*** Create Flight Plan")
	fp, err := c.CreateFlight(ship.ID, "OE-PM")
	if err != nil {
		log.Fatal(err)
	}

	log.Print("*** View Flight Plan")
	_, err = c.ShowFlight(fp.ID)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("*** Wait for flight to arrive")
	select {
	case <-time.After(fp.ArrivesAt.Sub(time.Now())):
	}

	log.Print("*** Sell Trade Goods")
	_, err = c.SellCargo(ship.ID, "METALS", 25)
	if err != nil {
		log.Fatal(err)
	}
}
