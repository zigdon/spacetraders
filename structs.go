package spacetraders

import (
	"fmt"
	"strings"
	"time"
)

// JSON responses
type StatusRes struct {
	Status string `json:"status"`
}

type ClaimRes struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type AccountRes struct {
	User User `json:"user"`
}

type LoanRes struct {
	Loans []Loan `json:"loans"`
}

type TakeLoanRes struct {
	Credits int  `json:"credits"`
	Loan    Loan `json:"loan"`
}

type MyLoansRes struct {
	Loans []Loan `json:"loans"`
}

type BuyShipRes struct {
	User User `json:"user"`
	Ship Ship `json:"ship"`
}

type MyShipsRes struct {
	Ships []Ship `json:"ships"`
}

type ShipListingRes struct {
	Ships []ShipListing `json:"shipListings"`
}

type SystemsRes struct {
	Systems []System `json:"systems"`
}

type LocationsRes struct {
	Locations []Location `json:"locations"`
}

type BuyRes struct {
	User  User  `json:"user"`
	Order Order `json:"order"`
	Ship  Ship  `json:"ship"`
}

type SellRes struct {
	Credits int   `json:"credits"`
	Order   Order `json:"order"`
	Ship    Ship  `json:"ship"`
}

type MarketplaceRes struct {
	Offers []Offer `json:"marketplace"`
}

type FlightPlanRes struct {
	FlightPlan FlightPlan `json:"flightPlan"`
}

// Core types

type Loan struct {
	Due                time.Time `json:"due"`
	ID                 string    `json:"id"`
	ShortID            string
	RepaymentAmount    int    `json:"repaymentAmount"`
	Status             string `json:"status"`
	Amount             int    `json:"amount"`
	CollateralRequired bool   `json:"collateralRequired"`
	Rate               int    `json:"rate"`
	TermInDays         int    `json:"termInDays"`
	Type               string `json:"type"`
}

func (l *Loan) String() string {
	if l.Due.After(time.Now()) {
		return fmt.Sprintf("id: %s, due in: %s, amt: %d, status: %s, type: %s",
			l.ShortID, l.Due.Sub(time.Now()).Truncate(time.Second), l.RepaymentAmount, l.Status, l.Type)
	}
	return fmt.Sprintf("id: %s, was due: %s, amt: %d, status: %s, type: %s",
		l.ShortID, l.Due.Local(), l.RepaymentAmount, l.Status, l.Type)
}

type User struct {
	Credits        int       `json:"credits"`
	JoinedAt       time.Time `json:"joinedAt"`
	ShipCount      int       `json:"shipCount"`
	StructureCount int       `json:"structureCount"`
	Username       string    `json:"username"`
}

func (u *User) String() string {
	return fmt.Sprintf("%s: Credits: %d, Ships: %d, Structures: %d, Joined: %s",
		u.Username, u.Credits, u.ShipCount, u.StructureCount, u.JoinedAt.Local())
}

type Ship struct {
	Cargo             []Cargo `json:"cargo"`
	Class             string  `json:"class"`
	FlightPlanID      string  `json:"flightPlanId,omitempty"`
	ShortFlightPlanID string
	ID                string `json:"id"`
	ShortID           string
	LocationName      string `json:"location"`
	Manufacturer      string `json:"manufacturer"`
	MaxCargo          int    `json:"maxCargo"`
	LoadingSpeed      int    `json:"loadingSpeed"`
	Plating           int    `json:"plating"`
	SpaceAvailable    int    `json:"spaceAvailable"`
	Speed             int    `json:"speed"`
	Type              string `json:"type"`
	Weapons           int    `json:"weapons"`
	X                 int    `json:"x"`
	Y                 int    `json:"y"`
}

func (s *Ship) Filter(word string) bool {
	word = strings.ToLower(word)
	for _, bit := range []string{s.ShortID, s.ShortFlightPlanID, s.Class, s.LocationName, s.Type} {
		if strings.ToLower(bit) == word {
			return true
		}
	}

	for _, id := range []string{s.ID, s.FlightPlanID} {
		if strings.HasPrefix(id, word) {
			return true
		}
	}

	return false
}

func (s *Ship) String() string {
	res := []string{}
	i := func(format string, args ...interface{}) { res = append(res, fmt.Sprintf(format, args...)) }
	i("%s: %s %s (%s)", s.ShortID, s.Manufacturer, s.Class, s.Type)
	i("ID: %s", s.ID)
	i("Speed: %d, Max cargo: %d, Available space: %d, Weapons: %d, Plating: %d",
		s.Speed, s.MaxCargo, s.SpaceAvailable, s.Weapons, s.Plating)
	if s.FlightPlanID == "" {
		i("At %s (%d, %d)", s.LocationName, s.X, s.Y)
	} else {
		i("In flight: %s", s.FlightPlanID)
	}
	if len(s.Cargo) > 0 {
		i("Cargo:")
		for _, c := range s.Cargo {
			i("  %s", c)
		}
	}

	return strings.Join(res, "\n")
}

func (s *Ship) Short() string {
	if s.FlightPlanID == "" {
		return fmt.Sprintf("%s: %s %s (%s): Loc: %s (%d, %d), Space: %d",
			s.ShortID, s.Manufacturer, s.Class, s.Type, s.LocationName, s.X, s.Y, s.SpaceAvailable)
	}
	return fmt.Sprintf("%s: %s %s (%s): Flight plan: %s, Space: %d",
		s.ShortID, s.Manufacturer, s.Class, s.Type, s.FlightPlanID, s.SpaceAvailable)
}

type Cargo struct {
	Good        string `json:"good"`
	Quantity    int    `json:"quantity"`
	TotalVolume int    `json:"totalVolume"`
}

func (c Cargo) String() string {
	return fmt.Sprintf("%d of %s (%d)", c.Quantity, c.Good, c.TotalVolume)
}

type ShipListing struct {
	Class             string `json:"class"`
	Manufacturer      string `json:"manufacturer"`
	MaxCargo          int    `json:"maxCargo"`
	Plating           int    `json:"plating"`
	PurchaseLocations []struct {
		LocationName string `json:"location"`
		Price        int    `json:"price"`
	} `json:"purchaseLocations"`
	Speed   int    `json:"speed"`
	Type    string `json:"type"`
	Weapons int    `json:"weapons"`
}

func (s ShipListing) String() string {
	res := []string{}
	i := func(st string) {
		res = append(res, st)
	}
	i(fmt.Sprintf("%s: %s %s", s.Type, s.Manufacturer, s.Class))
	i(fmt.Sprintf("speed: %d, cargo: %d, weapons: %d, plating: %d", s.Speed, s.MaxCargo, s.Weapons, s.Plating))
	for _, l := range s.PurchaseLocations {
		i(fmt.Sprintf("  %s: %d", l.LocationName, l.Price))
	}

	return strings.Join(res, "\n")
}

func (s ShipListing) Filter(word string) bool {
	for _, bit := range []string{s.Type, s.Manufacturer, s.Class} {
		if bit == word {
			return true
		}
	}
	return false
}

type System struct {
	Symbol    string     `json:"symbol"`
	Name      string     `json:"name"`
	Locations []Location `json:"locations"`
}

func (s System) String() string {
	structs := 0
	hasMsgs := ""
	for _, l := range s.Locations {
		structs += len(l.Structures)
		if len(l.Messages) > 0 {
			hasMsgs = "(see details for more)"
		}
	}
	return fmt.Sprintf("%s: %s\nLocations: %d, Structures: %d %s\n",
		s.Symbol, s.Name, len(s.Locations), structs, hasMsgs)
}

func (s System) Details(indent int) string {
	res := []string{fmt.Sprintf("%s: %s", s.Symbol, s.Name)}
	for _, l := range s.Locations {
		res = append(res, l.Short(indent))
	}
	return strings.Join(res, "\n")
}

type Location struct {
	Symbol             string      `json:"symbol"`
	Type               string      `json:"type"`
	Name               string      `json:"name"`
	X                  int         `json:"x"`
	Y                  int         `json:"y"`
	AllowsConstruction bool        `json:"allowsConstruction"`
	Structures         []Structure `json:"structures"`
	Traits             []string    `json:"traits"`
	Messages           []string    `json:"messages,omitempty"`
}

func (l Location) Short(indent int) string {
	prefix := strings.Repeat("  ", indent)
	res := fmt.Sprintf("%s%s: %s (%d, %d)", prefix, l.Name, l.Type, l.X, l.Y)
	if len(l.Structures) > 0 {
		res += fmt.Sprintf(" %d structures", len(l.Structures))
	}
	if len(l.Messages) > 0 {
		res += " (see details for more)"
	}
	return res
}

func (l Location) Details(indent int) string {
	prefix := strings.Repeat("  ", indent)
	var res []string
	i := func(s string) { res = append(res, fmt.Sprintf("%s%s", prefix, s)) }
	i(fmt.Sprintf("%s: %s", l.Symbol, l.Name))
	prefix = strings.Repeat("  ", indent+1)
	i(fmt.Sprintf("Type: %s  (%d, %d)", l.Type, l.X, l.Y))
	if l.AllowsConstruction {
		i("Allows construction.")
	}
	if len(l.Traits) > 0 {
		i(fmt.Sprintf("Traits: %v", l.Traits))
	}
	if len(l.Structures) > 0 {
		i(fmt.Sprintf("%d structures:", len(l.Structures)))
		for _, st := range l.Structures {
			res = append(res, st.Details(indent+1))
		}
		i("")
	}
	if len(l.Messages) > 0 {
		for _, m := range l.Messages {
			i(m)
		}
	}
	return strings.Join(res, "\n")
}

type Structure struct {
	ID       string `json:"id"`
	ShortID  string
	Type     string `json:"type"`
	Location string `json:"location"`
}

func (st Structure) Details(indent int) string {
	prefix := strings.Repeat("  ", indent)
	return fmt.Sprintf("%s%s: %s", prefix, st.ID, st.Type)
}

type Order struct {
	Good         string `json:"good"`
	PricePerUnit int    `json:"pricePerUnit"`
	Quantity     int    `json:"quantity"`
	Total        int    `json:"total"`
}

type Offer struct {
	Symbol               string `json:"symbol"`
	VolumePerUnit        int    `json:"volumePerUnit"`
	PricePerUnit         int    `json:"pricePerUnit"`
	Spread               int    `json:"spread"`
	PurchasePricePerUnit int    `json:"purchasePricePerUnit"`
	SellPricePerUnit     int    `json:"sellPricePerUnit"`
	QuantityAvailable    int    `json:"quantityAvailable"`
}

func (o Offer) String() string {
	return fmt.Sprintf("%6d x %-30s Buy: %-6d  Sell: %-6d  Spread: %-4d  Volume per unit: %d",
		o.QuantityAvailable, o.Symbol, o.PurchasePricePerUnit, o.SellPricePerUnit, o.Spread, o.VolumePerUnit)
}

type FlightPlan struct {
	ArrivesAt              time.Time `json:"arrivesAt"`
	Departure              string    `json:"departure"`
	Destination            string    `json:"destination"`
	Distance               int       `json:"distance"`
	FuelConsumed           int       `json:"fuelConsumed"`
	FuelRemaining          int       `json:"fuelRemaining"`
	ID                     string    `json:"id"`
	ShortID                string
	ShipID                 string `json:"shipId"`
	ShortShipID            string
	TerminatedAt           time.Time `json:"terminatedAt"`
	TimeRemainingInSeconds int       `json:"timeRemainingInSeconds"`
}

func (f FlightPlan) Short() string {
	return fmt.Sprintf("%s: %s %s->%s, ETA: %s",
		f.ShortID, f.ShortShipID, f.Departure, f.Destination,
		f.ArrivesAt.Sub(time.Now()).Truncate(time.Second))
}

func (f FlightPlan) String() string {
	var arrives string
	if f.ArrivesAt.After(time.Now()) {
		arrives = fmt.Sprintf("  Arrives at: %s, ETA: %s",
			f.ArrivesAt.Local(), f.ArrivesAt.Sub(time.Now()).Truncate(time.Second))
	} else {
		arrives = fmt.Sprintf("  Arrived at %s", f.ArrivesAt.Local())
	}

	return strings.Join([]string{
		fmt.Sprintf("%s: %s %s->%s", f.ShortID, f.ShortShipID, f.Departure, f.Destination),
		fmt.Sprintf("  ID: %s", f.ID),
		fmt.Sprintf("  ShipID: %s", f.ShipID),
		arrives,
		fmt.Sprintf("  Fuel consumed: %d, remaining: %d", f.FuelConsumed, f.FuelRemaining),
		fmt.Sprintf("  Distance: %d, Terminated: %s", f.Distance, f.TerminatedAt),
	}, "\n")
}
