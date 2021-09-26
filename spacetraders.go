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
	"sort"
	"strings"
	"time"
)

var useDebug = flag.Bool("debug", false, "Print out all debug statements")

type Client struct {
	username string
	token    string
	server   string
	cache    map[CacheKey]*cacheItem
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

	if err := dec.Decode(obj); err != nil {
		return fmt.Errorf("error decoding json: %v\n%s", err, data)
	}

	return nil
}

func New() *Client {
	return &Client{
		server: "https://api.spacetraders.io",
		cache:  make(map[CacheKey]*cacheItem),
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

// Caching
type CacheKey string

const (
	LOANS       CacheKey = "loans"
	SHIPS       CacheKey = "ships"
	MYLOCATIONS CacheKey = "my locations"
	LOCATIONS   CacheKey = "all locations"
	SYSTEMS     CacheKey = "systems"
	FLIGHTS     CacheKey = "flight"
)

type cacheItem struct {
	expiresOn time.Time
	data      []string
	shorts    []string
}

var (
	shortToID  = make(map[string]string)
	idToShort  = make(map[string]string)
	shortIndex = make(map[CacheKey]int)
)

func makeShort(key CacheKey, data string) string {
	short, ok := idToShort[data]
	if ok {
		return short
	}
	var prefix string
	switch key {
	case LOANS:
		prefix = "ln"
	case SHIPS:
		prefix = "s"
	case FLIGHTS:
		prefix = "f"
	default:
		log.Printf("Unknown prefix for %s", key)
		prefix = "X"
	}

	shortIndex[key]++
	short = fmt.Sprintf("%s-%d", prefix, shortIndex[key])
	idToShort[data] = short
	shortToID[short] = data
	log.Printf("Created short %q in %q for %q", short, key, data)
	return short
}

func makeLong(id string) string {
	if long, ok := shortToID[id]; ok {
		return long
	}
	return id
}

func getShorts(key CacheKey, data []string) []string {
	res := []string{}
	for _, d := range data {
		res = append(res, makeShort(key, d))
	}

	return res
}

func (c *Client) Add(key CacheKey, data string) {
	short := makeShort(key, data)
	if _, ok := c.cache[key]; !ok {
		c.cache[key] = &cacheItem{}
	}
	c.cache[key].data = sort.StringSlice(append(c.cache[key].data, data))
	c.cache[key].shorts = sort.StringSlice(append(c.cache[key].shorts, short))
}

func (c *Client) Store(key CacheKey, validFor time.Duration, data []string, shorts []string) {
	sort.Strings(data)
	c.cache[key] = &cacheItem{expiresOn: time.Now().Add(validFor), data: data, shorts: shorts}
}

func (c *Client) Restore(key CacheKey) []string {
	cached, ok := c.cache[key]
	if !ok || cached.expiresOn.Before(time.Now()) {
		log.Printf("Cache miss: %q", key)
		if err := c.Cache(key); err != nil {
			log.Printf("Error caching %q: %v", key, err)
			return []string{}
		}
		cached = c.cache[key]
	} else {
		log.Printf("Cache hit: %q", key)
	}
	if cached.shorts != nil {
		return append(cached.shorts, cached.data...)
	}
	return cached.data
}

func (c *Client) Cache(key CacheKey) error {
	switch key {
	case LOCATIONS, SYSTEMS:
		_, err := c.ListSystems()
		return err
	case MYLOCATIONS, FLIGHTS:
		_, err := c.MyShips()
		return err
	default:
		return fmt.Errorf("don't know how to cache %q", key)
	}
}

// Low level REST functions
type httpMethod string

const (
	post httpMethod = "POST"
	get  httpMethod = "GET"
)

func (c *Client) useAPI(method httpMethod, url string, args map[string]string, obj interface{}) error {
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
		return fmt.Errorf("error calling %q: %v", url, err)
	}
	if err := decodeJSON(res, obj); err != nil {
		return fmt.Errorf("can't decode json: %v\n%s", err, res)
	}

	return nil
}

func backoff(deadline int, f func() (*http.Response, error)) (*http.Response, error) {
	wait := 1.0
	start := time.Now()
	for {
		res, err := f()
		if err != nil {
			return nil, err
		}

		if res.StatusCode != 429 { // Too many requests
			return res, nil
		}

		if start.Add(time.Duration(deadline)).After(time.Now()) {
			return nil, fmt.Errorf("backoff deadline of %d seconds exceeded", deadline)
		}

		log.Printf("Too many requests, waiting %0.0f seconds, deadline %d", wait, deadline)
		select {
		case <-time.After(time.Duration(wait) * time.Second):
		}
		wait *= 1.5
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

	resp, err := backoff(30, func() (*http.Response, error) {
		return http.Post(uri, "application/json", body)
	})

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		defer resp.Body.Close()
		resBody, _ := ioutil.ReadAll(resp.Body)
		return string(resBody), nil
	}

	if resp != nil {
		return "", fmt.Errorf("error in POST %q: (rc=%d) %q %v", base, resp.StatusCode, resp.Status, err)
	}

	return "", fmt.Errorf("error in POST %q: (rc=%d) %v", base, resp.StatusCode, err)
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
	resp, err := backoff(30, func() (*http.Response, error) {
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
	c.Add(LOANS, tlr.Loan.ID)

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
	c.Store(LOANS, time.Minute, ids, shorts)

	return mlr.Loans, nil
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
			locations = append(locations, l.Symbol)
		}
	}
	c.Store(SYSTEMS, time.Hour, systems, nil)
	c.Store(LOCATIONS, time.Hour, locations, nil)

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

	return lr.Locations, nil
}

// Ships
// ##ENDPOINT List ships for purchase - `/systems/LOCATION/ship-listing`
func (c *Client) ListShips(system string) ([]ShipListing, error) {
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
	c.Add(SHIPS, bsr.Ship.ID)

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
		shorts = append(shorts, s.ShortID)
		locs = append(locs, s.LocationName)
		locs = append(locs, s.FlightPlanID)
	}
	c.Store(SHIPS, time.Minute, ids, shorts)
	c.Store(MYLOCATIONS, time.Minute, locs, nil)
	c.Store(FLIGHTS, time.Hour, flights, nil)

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
	fpr.FlightPlan.ShortID = makeShort(FLIGHTS, fpr.FlightPlan.ID)
	c.Add(FLIGHTS, fpr.FlightPlan.ID)

	return &fpr.FlightPlan, nil
}

// ##ENDPOINT Show flight plans - `/my/flight-plans/FLIGHTID`
func (c *Client) ShowFlight(flightID string) (*FlightPlan, error) {
	flightID = makeLong(flightID)
	fpr := &FlightPlanRes{}

	if err := c.useAPI(get, fmt.Sprintf("/my/flight-plans/%s", flightID), nil, fpr); err != nil {
		return nil, err
	}
	fpr.FlightPlan.ShortID = makeShort(FLIGHTS, fpr.FlightPlan.ID)
	fpr.FlightPlan.ShortShipID = makeShort(SHIPS, fpr.FlightPlan.ShipID)

	return &fpr.FlightPlan, nil
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

	return &sr.Order, nil
}

// ##ENDPOINT Available offers - `/locations/LOCATION/marketplace`
func (c *Client) Marketplace(loc string) ([]Offer, error) {
	mr := &MarketplaceRes{}

	if err := c.useAPI(get, fmt.Sprintf("/locations/%s/marketplace", loc), nil, mr); err != nil {
		return nil, err
	}

	return mr.Offers, nil
}
