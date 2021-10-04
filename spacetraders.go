package spacetraders

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var useDebug = flag.Bool("debug", false, "Print out all debug statements")

type Client struct {
	username    string
	token       string
	server      string
	flightDests map[string]string
	cache       *Cache
}

// Utils
func debug(format string, args ...interface{}) {
	if !*useDebug {
		return
	}
	log.Output(2, fmt.Sprintf(format, args...))
}

func decodeJSON(data string, obj interface{}) error {
	dec := json.NewDecoder(strings.NewReader(data))
	dec.DisallowUnknownFields()

	if err := dec.Decode(obj); err != nil {
		return fmt.Errorf("error decoding json into %#v: %v\n%s", obj, err, data)
	}

	return nil
}

func New() *Client {
	return &Client{
		server: "https://api.spacetraders.io",
		cache: GetCache(),
		flightDests: make(map[string]string),
	}
}

func (c *Client) Load(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Can't read token from %q: %v", path, err)
	}

	lines := strings.Split(string(data), "\n")
	log.Printf("Token for %q loaded from %q.", lines[0], path)

	c.username = strings.TrimSpace(lines[0])
	c.token = strings.TrimSpace(lines[1])

	return nil
}

func (c *Client) UpdateCache(key CacheKey) error {
	switch key {
	case LOCATIONS, SYSTEMS:
		_, err := c.ListSystems()
		return err
	case MYLOCATIONS, FLIGHTS, FLIGHTDESTS:
		_, err := c.MyShips()
		return err
	case CARGO:
		return nil
	default:
		return fmt.Errorf("don't know how to cache %q", key)
	}
}

// Low level REST functions
var calls []time.Time

const (
	burstCount = 8
	burstReset = 10
	callRate   = 2
)

func rateLimit() {
	defer func() {
		calls = append(calls, time.Now())
	}()
	if len(calls) == 0 {
		return
	}
	// Remove expired calls
	for len(calls) > 0 && calls[0].Add(burstReset*time.Second).Before(time.Now()) {
		calls = calls[1:]
	}

	// Count how many calls in the last burst, and how many in the last second
	rate := 0
	var wait time.Time
	for _, c := range calls {
		if c.Add(time.Second).After(time.Now()) {
			if wait.IsZero() {
				wait = c
			}
			rate++
		}
	}

	// If we have no calls in the last second, or we haven't filled the burst yet
	if wait.IsZero() || len(calls) < burstCount {
		return
	}
	delay := time.Now().Sub(wait)
	log.Printf("Waiting %s for rate limit", delay.Truncate(time.Millisecond))
	select {
	case <-time.After(delay):
	}
}

type httpMethod string

const (
	post httpMethod = "POST"
	get  httpMethod = "GET"
)

func (c *Client) useAPI(method httpMethod, url string, args map[string]string, obj interface{}) error {
	rateLimit()
	var f func(string, map[string]string) (string, error)
	if method == post {
		f = c.Post
	} else if method == get {
		f = c.Get
	} else {
		return fmt.Errorf("Unknown method %q", method)
	}
	debug("Calling %q with %+v...", url, args)
	res, err := f(url, args)
	debug("... %v\n%s", err, res)
	if err != nil {
		return fmt.Errorf("error calling %q [%+v]: %v", url, args, err)
	}
	if err := decodeJSON(res, obj); err != nil {
		return fmt.Errorf("can't decode json: %v\n%s", err, res)
	}

	return nil
}

func backoff(f func() (*http.Response, error)) (*http.Response, error) {
	wait := 1.0
	start := time.Now()
	timeout := start.Add(time.Minute)
	retryable := map[int]int{
		// 422: 5,  // Unprocessable Entity
		429: 30, // Too many requests"
	}
	for {
		res, err := f()
		if err != nil {
			return res, err
		}

		if timeout.Before(time.Now()) {
			return res, fmt.Errorf("backoff deadline exceeded")
		}

		if ec, ok := retryable[res.StatusCode]; ok {
			timeout = start.Add(time.Duration(ec) * time.Second)
			log.Printf("%d: waiting %s seconds before retrying, %s to deadline",
				res.StatusCode, time.Duration(wait)*time.Second, timeout.Sub(time.Now()).Truncate(time.Second))
			select {
			case <-time.After(time.Duration(wait) * time.Second):
			}
			wait *= 1.5
			continue
		}

		return res, nil
	}
}

func (c *Client) Post(base string, args map[string]string) (string, error) {
	var uri string
	if args == nil {
		args = make(map[string]string)
	}
	if c.server != "" {
		uri = c.server + base
	} else {
		uri = base
	}
	if c.token != "" {
		uri += "?" + url.Values{"token": []string{c.token}}.Encode()
	}
	jsonBody, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("Can't encode %+v: %v", args, err)
	}
	body := bytes.NewBuffer(jsonBody)

	resp, err := backoff(func() (*http.Response, error) {
		return http.Post(uri, "application/json", body)
	})

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		defer resp.Body.Close()
		resBody, _ := ioutil.ReadAll(resp.Body)
		return string(resBody), nil
	}

	if resp != nil {
		return "", fmt.Errorf("error in POST %q (rc=%d %q): %v", base, resp.StatusCode, resp.Status, err)
	}

	return "", fmt.Errorf("error in POST %q: (nil response) %v", base, err)
}

func (c *Client) Get(base string, args map[string]string) (string, error) {
	var uri string
	var values = make(url.Values)
	if args != nil {
		for k, v := range args {
			values[k] = []string{v}
		}
	}
	if c.server != "" {
		uri = c.server + base
	} else {
		uri = base
	}
	if c.token != "" {
		values["token"] = []string{c.token}
	}
	if len(values) > 0 {
		uri += "?" + values.Encode()
	}
	resp, err := backoff(func() (*http.Response, error) {
		return http.Get(uri)
	})
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		return string(body), nil
	}

	if resp != nil {
		return "", fmt.Errorf("error in GET %q: rc=%d, %q %v", base, resp.StatusCode, resp.Status, err)
	}
	return "", fmt.Errorf("error in GET %q: rc=%d, %v", base, resp.StatusCode, err)
}

// ##ENDPOINT Game status - `/game/status`
func (c *Client) Status() error {
	sr := &StatusRes{}
	if err := c.useAPI(get, "/game/status", nil, sr); err != nil {
		return err
	}
	log.Printf("Status: %s", sr.Status)

	return nil
}

// Account
// ##ENDPOINT Claim username - `/users/USERNAME/claim`
func (c *Client) Claim(username string) (string, *User, error) {
	if c.username != "" {
		return "", nil, fmt.Errorf("Can't claim while already logged in as %q", c.username)
	}

	cr := &ClaimRes{}
	if err := c.useAPI(post, fmt.Sprintf("/users/%s/claim", username), nil, cr); err != nil {
		return "", nil, err
	}

	c.token = cr.Token
	c.username = username

	return cr.Token, &cr.User, nil
}

func (c *Client) Logout() error {
	if c.username == "" {
		return fmt.Errorf("Already logged out.")
	}

	log.Printf("Logging out %q", c.username)

	c.username = ""
	c.token = ""

	return nil
}

// ##ENDPOINT Account details - `/my/account`
func (c *Client) Account() (*User, error) {
	ar := &AccountRes{}
	if err := c.useAPI(get, "/my/account", nil, ar); err != nil {
		return nil, err
	}

	return &ar.User, nil
}

// Loans
// ##ENDPOINT Available loans - `/types/loans`
func (c *Client) AvailableLoans() ([]Loan, error) {
	lr := &LoanRes{}

	if err := c.useAPI(get, "/types/loans", nil, lr); err != nil {
		return nil, err
	}

	loans := []string{}
	for _, l := range lr.Loans {
		loans = append(loans, l.ID)
	}

	return lr.Loans, nil
}

// ##ENDPOINT Take out loan - `/my/loans`
func (c *Client) TakeLoan(name string) (*Loan, error) {
	tlr := &TakeLoanRes{}

	if err := c.useAPI(post, "/my/loans", map[string]string{"type": name}, tlr); err != nil {
		return nil, err
	}
	tlr.Loan.ShortID = makeShort(LOANS, tlr.Loan.ID)
	c.cache.Add(LOANS, tlr.Loan.ID)

	return &tlr.Loan, nil
}

// ##ENDPOINT List outstanding loans - `/my/loans`
func (c *Client) MyLoans() ([]Loan, error) {
	mlr := &MyLoansRes{}

	if err := c.useAPI(get, "/my/loans", nil, mlr); err != nil {
		return nil, err
	}

	ids := []string{}
	shorts := []string{}
	for i, l := range mlr.Loans {
		ids = append(ids, l.ID)
		mlr.Loans[i].ShortID = makeShort(LOANS, l.ID)
		shorts = append(shorts, l.ID)
	}
	c.cache.Store(LOANS, ids, shorts)

	return mlr.Loans, nil
}

// ##ENDPOINT Pay off a loan - `/my/loans/LOANID`
func (c *Client) PayLoan(loanID string) error {
	plr := &PayLoanRes{}

	if err := c.useAPI(post, fmt.Sprintf("/my/loans/%s", loanID), nil, plr); err != nil {
		return err
	}

	return nil
}

// Systems
// ##ENDPOINT List all systems - `/game/systems`
func (c *Client) ListSystems() ([]System, error) {
	sr := &SystemsRes{}

	if err := c.useAPI(get, "/game/systems", nil, sr); err != nil {
		return nil, err
	}

	systems := []string{}
	locations := []string{}
	for _, s := range sr.Systems {
		systems = append(systems, s.Symbol)
		for _, l := range s.Locations {
			l.SystemSymbol = s.Symbol
			locations = append(locations, l.Symbol)
		}
	}
	c.cache.Store(SYSTEMS, systems, nil)
	c.cache.Store(LOCATIONS, locations, nil)

	return sr.Systems, nil
}

// ##ENDPOINT List locations in a system - `/systems/SYSTEM/locations`
func (c *Client) ListLocations(system string, kind string) ([]Location, error) {
	lr := &LocationsRes{}

	args := map[string]string{
		"type": kind,
	}

	if err := c.useAPI(get, fmt.Sprintf("/systems/%s/locations", system), args, lr); err != nil {
		return nil, err
	}

	for _, l := range lr.Locations {
		l.SystemSymbol = system
	}

	return lr.Locations, nil
}

// Ships
// ##ENDPOINT List ships for purchase - `/systems/LOCATION/ship-listing`
func (c *Client) ListShips(system string) ([]Ship, error) {
	slr := &ShipListingRes{}

	if err := c.useAPI(get, fmt.Sprintf("/systems/%s/ship-listings", system), nil, slr); err != nil {
		return nil, err
	}

	return slr.Ships, nil
}

// ##ENDPOINT Buy ship - `/my/ships`
func (c *Client) BuyShip(location, kind string) (*Ship, error) {
	bsr := &BuyShipRes{}
	args := map[string]string{
		"location": location,
		"type":     kind,
	}

	if err := c.useAPI(post, "/my/ships", args, bsr); err != nil {
		return nil, err
	}
	bsr.Ship.ShortID = makeShort(SHIPS, bsr.Ship.ID)
	c.cache.Add(SHIPS, bsr.Ship.ID)

	return &bsr.Ship, nil
}

// ##ENDPOINT List my ship - `/my/ships`
func (c *Client) MyShips() ([]Ship, error) {
	msr := &MyShipsRes{}

	if err := c.useAPI(get, "/my/ships", nil, msr); err != nil {
		return nil, err
	}

	ids := []string{}
	shorts := []string{}
	locs := []string{}
	flights := []string{}
	for i, s := range msr.Ships {
		ids = append(ids, s.ID)
		msr.Ships[i].ShortID = makeShort(SHIPS, s.ID)
		if s.FlightPlanID != "" {
			msr.Ships[i].ShortFlightPlanID = makeShort(FLIGHTS, s.FlightPlanID)
			msr.Ships[i].FlightPlanDest = c.getFlightDest(s.FlightPlanID)
			flights = append(flights, s.FlightPlanID)
		}
		shorts = append(shorts, s.ShortID)
		locs = append(locs, s.LocationName)
	}
	c.cache.Store(SHIPS, ids, shorts)
	c.cache.Store(MYLOCATIONS, locs, nil)
	c.cache.Store(FLIGHTS, flights, nil)

	return msr.Ships, nil
}

// ##ENDPOINT Create flight plan - `/my/flight-plans`
func (c *Client) CreateFlight(shipID, destination string) (*FlightPlan, error) {
	shipID = makeLong(shipID)
	fpr := &FlightPlanRes{}
	args := map[string]string{
		"shipId":      shipID,
		"destination": destination,
	}

	if err := c.useAPI(post, "/my/flight-plans", args, fpr); err != nil {
		return nil, err
	}
	fp := fpr.FlightPlan
	fp.ShortID = makeShort(FLIGHTS, fp.ID)
	fp.ShortShipID = makeShort(SHIPS, fp.ShipID)
	c.cache.Add(FLIGHTS, fp.ID)

	return &fp, nil
}

// ##ENDPOINT Show flight plans - `/my/flight-plans/FLIGHTID`
func (c *Client) ShowFlight(flightID string) (*FlightPlan, error) {
	flightID = makeLong(flightID)
	fpr := &FlightPlanRes{}

	if err := c.useAPI(get, fmt.Sprintf("/my/flight-plans/%s", flightID), nil, fpr); err != nil {
		return nil, err
	}
	fp := fpr.FlightPlan
	fp.ShortID = makeShort(FLIGHTS, fp.ID)
	fp.ShortShipID = makeShort(SHIPS, fp.ShipID)

	return &fp, nil
}

func (c *Client) getFlightDest(flightID string) string {
	if d, ok := c.flightDests[flightID]; ok {
		return d
	}
	fp, err := c.ShowFlight(flightID)
	if err != nil {
		log.Printf("Error looking up %s: %v", flightID, err)
		return "Unknown"
	}
	c.flightDests[flightID] = fp.Destination

	return fp.Destination
}

// Goods and Cargo
// ##ENDPOINT Buy cargo - `/my/purchase-orders`
func (c *Client) BuyCargo(shipID, good string, qty int) (*Order, error) {
	shipID = makeLong(shipID)
	br := &BuyRes{}

	args := map[string]string{
		"shipId":   shipID,
		"good":     good,
		"quantity": fmt.Sprintf("%d", qty),
	}

	if err := c.useAPI(post, "/my/purchase-orders", args, br); err != nil {
		return nil, err
	}

	// Didn't error, must be real
	c.cache.Extend(CARGO, []string{good}, nil)

	return &br.Order, nil
}

// ##ENDPOINT Sell cargo - `/my/sell-orders`
func (c *Client) SellCargo(shipID, good string, qty int) (*Order, error) {
	shipID = makeLong(shipID)
	sr := &SellRes{}

	args := map[string]string{
		"shipId":   shipID,
		"good":     good,
		"quantity": fmt.Sprintf("%d", qty),
	}

	if err := c.useAPI(post, "/my/sell-orders", args, sr); err != nil {
		return nil, err
	}

	// Didn't error, must be real
	c.cache.Extend(CARGO, []string{good}, nil)

	return &sr.Order, nil
}

// ##ENDPOINT Available offers - `/locations/LOCATION/marketplace`
func (c *Client) Marketplace(loc string) ([]Offer, error) {
	mr := &MarketplaceRes{}

	if err := c.useAPI(get, fmt.Sprintf("/locations/%s/marketplace", loc), nil, mr); err != nil {
		return nil, err
	}
	cargoType := []string{}
	for _, o := range mr.Offers {
		cargoType = append(cargoType, o.Symbol)
	}
	c.cache.Extend(CARGO, cargoType, nil)

	return mr.Offers, nil
}
