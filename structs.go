package spacetraders

import (
	"fmt"
	"strings"
	"time"
)

type Client struct {
	username string
	token    string
	server   string
}

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

type Loan struct {
	Due                time.Time `json:"due"`
	ID                 string    `json:"id"`
	RepaymentAmount    int       `json:"repaymentAmount"`
	Status             string    `json:"status"`
	Amount             int       `json:"amount"`
	CollateralRequired bool      `json:"collateralRequired"`
	Rate               int       `json:"rate"`
	TermInDays         int       `json:"termInDays"`
	Type               string    `json:"type"`
}

type TakeLoanRes struct {
	Credits int  `json:"credits"`
	Loan    Loan `json:"loan"`
}

type MyLoansRes struct {
	Loans []Loan `json:"loans"`
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
		u.Username, u.Credits, u.ShipCount, u.StructureCount, u.JoinedAt)
}

type BuyShipRes struct {
	User struct {
		Credits int `json:"credits"`
	} `json:"user"`
	Ship Ship `json:"ship"`
}

type MyShipsRes struct {
	Ships []Ship `json:"ships"`
}

type Ship struct {
	Cargo          []Cargo `json:"cargo"`
	Class          string  `json:"class"`
	FlightPlanID   string  `json:"flightPlanId,omitempty"`
	ID             string  `json:"id"`
	Location       string  `json:"location"`
	Manufacturer   string  `json:"manufacturer"`
	MaxCargo       int     `json:"maxCargo"`
	Plating        int     `json:"plating"`
	SpaceAvailable int     `json:"spaceAvailable"`
	Speed          int     `json:"speed"`
	Type           string  `json:"type"`
	Weapons        int     `json:"weapons"`
	X              int     `json:"x"`
	Y              int     `json:"y"`
}

func (s *Ship) Filter(word string) bool {
	for _, bit := range []string{s.Class, s.Location, s.Type} {
		if bit == word {
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
	i("%s: %s %s (%s)", s.ID, s.Manufacturer, s.Class, s.Type)
	i("Speed: %d, Max cargo: %d, Available space: %d, Weapons: %d, Plating: %d",
		s.Speed, s.MaxCargo, s.SpaceAvailable, s.Weapons, s.Plating)
	if s.FlightPlanID == "" {
		i("At %s (%d, %d)", s.Location, s.X, s.Y)
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
			s.ID, s.Manufacturer, s.Class, s.Type, s.Location, s.X, s.Y, s.SpaceAvailable)
	}
	return fmt.Sprintf("%s: %s %s (%s): Flight plan: %s, Space: %d",
		s.ID, s.Manufacturer, s.Class, s.Type, s.FlightPlanID, s.SpaceAvailable)
}

type Cargo struct {
	Good        string `json:"good"`
	Quantity    int    `json:"quantity"`
	TotalVolume int    `json:"totalVolume"`
}

func (c Cargo) String() string {
	return fmt.Sprintf("%d of %s (%d)", c.Quantity, c.Good, c.TotalVolume)
}

type ShipListingRes struct {
	Ships []ShipListing `json:"shipListings"`
}

type ShipListing struct {
	Class             string `json:"class"`
	Manufacturer      string `json:"manufacturer"`
	MaxCargo          int    `json:"maxCargo"`
	Plating           int    `json:"plating"`
	PurchaseLocations []struct {
		Location string `json:"location"`
		Price    int    `json:"price"`
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
		i(fmt.Sprintf("  %s: %d", l.Location, l.Price))
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

type SystemsRes struct {
	Systems []System `json:"systems"`
}

type System struct {
	Symbol    string     `json:"symbol"`
	Name      string     `json:"name"`
	Locations []Location `json:"locations"`
}

func (s System) String() string {
	structs := 0
	msgs := []string{}
	for _, l := range s.Locations {
		structs += len(l.Structures)
		msgs = append(msgs, l.Messages...)
	}
	return fmt.Sprintf("%s: %s\nLocations: %d, Structures: %d\n  %s",
		s.Symbol, s.Name, len(s.Locations), structs, strings.Join(msgs, "\n  "))
}

func (s System) Details() string {
	res := []string{fmt.Sprintf("%s: %s", s.Symbol, s.Name)}
	for _, l := range s.Locations {
		res = append(res, strings.Join(l.Details(0), "\n"))
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
	Messages           []string    `json:"messages,omitempty"`
}

func (l Location) Details(indent int) []string {
	prefix := strings.Repeat("  ", indent)
	var res []string
	i := func(s string) { res = append(res, fmt.Sprintf("%s%s", prefix, s)) }
	i(fmt.Sprintf("%s: %s (%s - %d, %d)", l.Symbol, l.Name, l.Type, l.X, l.Y))
	if l.AllowsConstruction {
		i("Allows construction.")
	}
	if len(l.Structures) > 0 {
		i(fmt.Sprintf("%d structures:", len(l.Structures)))
		for _, st := range l.Structures {
			res = append(res, st.Details(indent+1)...)
		}
		i("")
	}
	return res
}

type Structure struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Location string `json:"location"`
}

func (st Structure) Details(indent int) []string {
	prefix := strings.Repeat("  ", indent)
	var res []string
	i := func(s string) { res = append(res, fmt.Sprintf("%s%s", prefix, s)) }
	i(fmt.Sprintf("%s: %s", st.ID, st.Type))
	return res
}
