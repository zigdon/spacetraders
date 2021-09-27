package cli

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/zigdon/spacetraders"
)

func init() {
	for _, c := range []cmd{
		{
			Section: "Account",
			Name:    "Account",
			Usage:   "Account",
			Help:    "Get details about the logged in account",
			Do:      doAccount,
		},
		{
			Section: "Account",
			Name:    "Login",
			Usage:   "Login [path/to/file]",
			Help:    "Load username and token from saved file, $HOME/.config/spacetraders.io by default",
			Do:      doLogin,
			MinArgs: 0,
			MaxArgs: 1,
		},
		{
			Section: "Account",
			Name:    "Logout",
			Usage:   "Logout",
			Help:    "Expire the current logged in token.",
			Do:      doLogout,
		},
		{
			Section: "Account",
			Name:    "Claim",
			Usage:   "Claim <username> <path/to/file>",
			Help:    "Claims a username, saves token to specified file",
			Do:      doClaim,
			MinArgs: 2,
			MaxArgs: 2,
		},
	} {
		if err := Register(c); err != nil {
			log.Fatalf("Can't register %q: %v", c.Name, err)
		}
	}
}

// Command implementations
func doAccount(c *spacetraders.Client, args []string) error {
	u, err := c.Account()
	if err != nil {
		return err
	}
	Out("%s", u)
	return nil
}

func doLogin(c *spacetraders.Client, args []string) error {
	path := filepath.Join(os.Getenv("HOME"), ".config/spacetraders.io")
	if len(args) > 0 {
		path = args[0]
	}
	if err := c.Load(path); err != nil {
		ErrMsg("Error loading token: %v", err)
	}

	return nil
}

func doClaim(c *spacetraders.Client, args []string) error {
	username := args[0]
	path := args[1]
	if _, err := os.Stat(args[1]); err == nil {
		return fmt.Errorf("%q already exists, aborting.", path)
	}

	token, _, err := c.Claim(username)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(
		path,
		[]byte(fmt.Sprintf("%s\n%s\n", username, token)),
		0600); err != nil {
		return fmt.Errorf("Error writing new token %q to %q: %v", token, path, err)
	}
	log.Printf("Got token %q for %q", token, username)

	return nil
}

func doLogout(c *spacetraders.Client, args []string) error {
	return c.Logout()
}
