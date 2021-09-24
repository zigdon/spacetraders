package spacetraders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func decodeJSON(data string, obj interface{}) error {
	dec := json.NewDecoder(strings.NewReader(data))

	if err := dec.Decode(obj); err != nil {
		return fmt.Errorf("error decoding json: %v\n%s", err, data)
	}

	return nil
}

type httpMethod string

const (
	post httpMethod = "POST"
	get  httpMethod = "GET"
)

// Utils
func New() *Client {
	return &Client{
		server: "https://api.spacetraders.io",
	}
}

func (c *Client) useAPI(method httpMethod, url string, args map[string]string, obj interface{}) error {
	var f func(string, map[string]string) (string, error)
	if method == post {
		f = c.Post
	} else if method == get {
		f = c.Get
	} else {
		return fmt.Errorf("Unknown method %q", method)
	}
	res, err := f(url, args)
	if err != nil {
		return fmt.Errorf("error calling %q: %v", url, err)
	}
	if err := decodeJSON(res, obj); err != nil {
		return fmt.Errorf("can't decode json: %v\n%s", err, res)
	}

	return nil
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

	resp, err := http.Post(uri, "application/json", body)
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
	resp, err := http.Get(uri)
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

func (c *Client) Status() error {
	sr := &StatusRes{}
	if err := c.useAPI(get, "/game/status", nil, sr); err != nil {
		return err
	}
	log.Printf("Status: %s", sr.Status)

	return nil
}

// Account
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

func (c *Client) Account() (*User, error) {
	ar := &AccountRes{}
	if err := c.useAPI(get, "/my/account", nil, ar); err != nil {
		return nil, err
	}

	return &ar.User, nil
}

// Loans
func (c *Client) AvailableLoans() ([]Loan, error) {
	lr := &LoanRes{}

	if err := c.useAPI(get, "/types/loans", nil, lr); err != nil {
		return nil, err
	}

	return lr.Loans, nil
}

func (c *Client) TakeLoan(name string) (*Loan, error) {
	tlr := &TakeLoanRes{}

	if err := c.useAPI(post, "/my/loans", map[string]string{"type": name}, tlr); err != nil {
		return nil, err
	}

	return &tlr.Loan, nil
}

func (c *Client) MyLoans() ([]Loan, error) {
	mlr := &MyLoansRes{}

	if err := c.useAPI(get, "/my/loans", nil, mlr); err != nil {
		return nil, err
	}

	return mlr.Loans, nil
}

// Systems
func (c *Client) ListSystems() ([]System, error) {
	sr := &SystemsRes{}

	if err := c.useAPI(get, "/game/systems", nil, sr); err != nil {
		return nil, err
	}

	return sr.Systems, nil
}

// Ships
func (c *Client) ListShips(system string) ([]ShipListing, error) {
	slr := &ShipListingRes{}

	if err := c.useAPI(get, fmt.Sprintf("/systems/%s/ship-listings", system), nil, slr); err != nil {
		return nil, err
	}

	return slr.Ships, nil
}

func (c *Client) BuyShip(location, kind string) (*Ship, error) {
	bsr := &BuyShipRes{}
	args := map[string]string{
		"location": location,
		"type":     kind,
	}

	if err := c.useAPI(post, "/my/ships", args, bsr); err != nil {
		return nil, err
	}

	return &bsr.Ship, nil
}

func (c *Client) MyShips() ([]Ship, error) {
	msr := &MyShipsRes{}

	if err := c.useAPI(get, "/my/ships", nil, msr); err != nil {
		return nil, err
	}

	return msr.Ships, nil
}
