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

func New() *Client {
	return &Client{
		server: "https://api.spacetraders.io",
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
		args["token"] = c.token
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

func (c *Client) Get(base string, args url.Values) (string, error) {
	var uri string
	if args == nil {
		args = make(url.Values)
	}
	if c.server != "" {
		uri = c.server + base
	} else {
		uri = base
	}
	if c.token != "" {
		args["token"] = []string{c.token}
	}
	if len(args) > 0 {
		uri += "?" + args.Encode()
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

func decodeJSON(data string, obj interface{}) error {
	dec := json.NewDecoder(strings.NewReader(data))

	if err := dec.Decode(obj); err != nil {
		return fmt.Errorf("error decoding json: %v\n%s", err, data)
	}

	return nil
}

func (c *Client) Status() error {
	res, err := c.Get("/game/status", nil)
	if err != nil {
		return err
	}
	sr := &StatusRes{}
	if err := decodeJSON(res, sr); err != nil {
		return fmt.Errorf("Can't decode status: %v\n%s", err, res)
	}
	log.Printf("Status: %s", sr.Status)

	return nil
}

func (c *Client) Claim(username string) (string, *User, error) {
	if c.username != "" {
		return "", nil, fmt.Errorf("Can't claim while already logged in as %q", c.username)
	}
	res, err := c.Post(fmt.Sprintf("/users/%s/claim", username), nil)
	if err != nil {
		return "", nil, err
	}
	cr := &ClaimRes{}
	if err := decodeJSON(res, cr); err != nil {
		return "", nil, fmt.Errorf("Can't claim %q: %v\n%s", username, err, res)
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
	res, err := c.Get("/my/account", nil)
	if err != nil {
		return nil, err
	}
	ar := &AccountRes{}
	if err := decodeJSON(res, ar); err != nil {
		return nil, fmt.Errorf("Can't decode account: %v\n%s", err, res)
	}

	return &ar.User, nil
}
