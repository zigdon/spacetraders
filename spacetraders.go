package spacetraders

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	token  string
	server string
}

func New(path string) *Client {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Can't read token from %q: %v", path, err)
	}

	token := string(data)
	log.Printf("Token loaded from %q.", path)

	return &Client{
		token:  strings.TrimSpace(token),
		server: "https://api.spacetraders.io",
	}
}

func (c *Client) Get(base string, args url.Values) (string, error) {
	var uri string
	if c.server != "" {
		uri = c.server + base
	} else {
		uri = base
	}
	if len(args) > 0 {
		uri += "?" + args.Encode()
	}
	resp, err := http.Get(uri)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			return "", fmt.Errorf("error in GET %q: %q %v", base, resp.Status, err)
		} else {
			return "", fmt.Errorf("error in GET %q: %v", base, err)
		}
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body), nil
}

func (c *Client) Status() error {
	status, err := c.Get("/game/status", nil)
	if err != nil {
		return err
	}
	log.Printf("Status: %q", status)

	return nil
}

func decodeJSON(data string, obj interface{}) error {
	dec := json.NewDecoder(strings.NewReader(data))

	if err := dec.Decode(obj); err != nil {
		return fmt.Errorf("error decoding json: %v\n%s", err, data)
	}

	return nil
}
