package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chzyer/readline"
	"github.com/zigdon/spacetraders"
)

type cmd struct {
	name    string
	usage   string
	help    string
	do      func(*spacetraders.Client, []string) error
	minArgs int
	maxArgs int
}

var commands = map[string]cmd{}

func doLoop(c *spacetraders.Client) {
	r, err := readline.New("> ")
	if err != nil {
		log.Fatalf("Can't readline: %v", err)
	}
	for {
		line, err := r.Readline()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Printf("Error while reading input: %v", err)
			break
		}

		words := strings.Split(strings.TrimSpace(line), " ")
		switch cmd := strings.ToLower(words[0]); cmd {
		case "exit":
			return
		default:
			if cmd, ok := commands[words[0]]; ok {
				if len(words)-1 < cmd.minArgs || len(words)-1 > cmd.maxArgs {
					log.Printf("Invalid arguments for %q", words[0])
					words = []string{"help", words[0]}
					cmd = commands["help"]
				}
				if err := cmd.do(c, words[1:]); err != nil {
					log.Printf("Error: %v", err)
				}
				fmt.Println()
				continue
			}
			log.Printf("Unknown command %q. Try 'help'.", cmd)
		}
	}
}

func doHelp(c *spacetraders.Client, args []string) error {
	if len(args) > 0 {
		cmd, ok := commands[args[0]]
		if ok {
			fmt.Printf("%s:\n%s", cmd.usage, cmd.help)
			return nil
		}
	}
	cmds := []string{}
	for _, cmd := range commands {
		cmds = append(cmds, cmd.name)
	}
	sort.Strings(cmds)
	fmt.Printf("Available commands: %s",
		strings.Join(cmds, ", "))
	return nil
}

func doAccount(c *spacetraders.Client, args []string) error {
	u, err := c.Account()
	if err != nil {
		return err
	}
	fmt.Printf("%s", u)
	return nil
}

func doLogin(c *spacetraders.Client, args []string) error {
	path := filepath.Join(os.Getenv("HOME"), ".config/spacetraders.io")
	if len(args) > 0 {
		path = args[0]
	}
	if err := c.Load(path); err != nil {
		log.Print(err)
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

func doLoans(c *spacetraders.Client, args []string) error {
	loans, err := c.AvailableLoans()
	if err != nil {
		return fmt.Errorf("error getting loans: %v", err)
	}

	for _, l := range loans {
		fmt.Printf("amt: %d, needs collateral: %v, rate: %d, term (days): %d, type: %s",
			l.Amount, l.CollateralRequired, l.Rate, l.TermInDays, l.Type)
	}

	return nil
}

func doTakeLoan(c *spacetraders.Client, args []string) error {
	loan, err := c.TakeLoan(args[0])
	if err != nil {
		return fmt.Errorf("error taking out loan: %v", err)
	}

	fmt.Printf("Loan taken, id=%s, due: %s", loan.ID, loan.Due)

	return nil
}

func doMyLoans(c *spacetraders.Client, args []string) error {
	loans, err := c.MyLoans()
	if err != nil {
		return fmt.Errorf("error querying loans: %v", err)
	}

	for _, l := range loans {
		fmt.Printf("id: %s, due: %s, amt: %d, status: %s, type: %s",
			l.ID, l.Due, l.RepaymentAmount, l.Status, l.Type)
	}

	return nil
}

func doListSystems(c *spacetraders.Client, args []string) error {
	systems, err := c.ListSystems()
	if err != nil {
		return fmt.Errorf("error listing systems: %v", err)
	}

	sys := []string{}
	cache := make(map[string]spacetraders.System)
	for _, s := range systems {
		sys = append(sys, s.Symbol)
		cache[s.Symbol] = s
	}
	sort.Strings(sys)
	if len(args) == 0 {
		fmt.Println("All systems:")
		for _, sym := range sys {
			fmt.Println(cache[sym])
		}
		return nil
	}

	fmt.Println(cache[args[0]].Details())
	return nil
}

func doListShips(c *spacetraders.Client, args []string) error {
	ships, err := c.ListShips(args[0])
	if err != nil {
		return fmt.Errorf("error listing ships in %q: %v", args[0], err)
	}

	if len(args) > 1 {
		for _, s := range ships {
			if !s.Filter(args[1]) {
				continue
			}
			fmt.Println(s)
		}
		return nil
	}

	for _, s := range ships {
		fmt.Println(s)
	}

	return nil
}

func doBuyShip(c *spacetraders.Client, args []string) error {
	ship, err := c.BuyShip(args[0], args[1])
	if err != nil {
		return fmt.Errorf("error buying ship %q at %q: %v", args[1], args[0], err)
	}

	fmt.Printf("New ship ID: %s", ship.ID)

	return nil
}

func doMyShips(c *spacetraders.Client, args []string) error {
	ships, err := c.MyShips()
	if err != nil {
		return fmt.Errorf("error listing my ships: %v", err)
	}

	res := []spacetraders.Ship{}
	for _, s := range ships {
		if len(args) > 0 {
			if !s.Filter(args[0]) {
				continue
			}
		}
		res = append(res, s)
	}

	switch len(res) {
	case 0:
		fmt.Println("No ships found.")
	case 1:
		fmt.Println(res[0].String())
	default:
		for _, s := range res {
			fmt.Println(s.Short())
		}
	}

	return nil
}

func main() {
	c := spacetraders.New()

	if err := c.Status(); err != nil {
		log.Fatalf("Game down: %v", err)
	}

	commands = map[string]cmd{
		"help": {
			name:    "Help",
			usage:   "help [command]",
			help:    "List all commands, or get information on a specific command",
			do:      doHelp,
			maxArgs: 1,
		},

		"account": {
			name:  "Account",
			usage: "account",
			help:  "Get details about the logged in account",
			do:    doAccount,
		},
		"login": {
			name:    "Login",
			usage:   "login [path/to/file]",
			help:    "Load username and token from saved file, $HOME/.config/spacetraders.io by default",
			do:      doLogin,
			minArgs: 0,
			maxArgs: 1,
		},
		"logout": {
			name:  "Logout",
			usage: "logout",
			help:  "Expire the current logged in token.",
			do:    doLogout,
		},
		"claim": {
			name:    "Claim",
			usage:   "claim username path/to/file",
			help:    "Claims a username, saves token to specified file",
			do:      doClaim,
			minArgs: 2,
			maxArgs: 2,
		},

		"availableloans": {
			name:  "AvailableLoans",
			usage: "availableLoans",
			help:  "Display currently available loans",
			do:    doLoans,
		},
		"takeloan": {
			name:    "TakeLoan",
			usage:   "takeLoan type",
			help:    "Take out one of the available loans",
			do:      doTakeLoan,
			minArgs: 1,
			maxArgs: 1,
		},
		"myloans": {
			name:  "MyLoans",
			usage: "myLoans",
			help:  "List outstanding loans",
			do:    doMyLoans,
		},

		"system": {
			name:    "System",
			usage:   "system [symbol]",
			help:    "Get details about a system, or all systems if not specified",
			do:      doListSystems,
			maxArgs: 1,
		},

		"listships": {
			name:  "ListShips",
			usage: "listShips location [filter]",
			help: "Show available ships at location. If filter is provided, " +
				"only show ships that match in type, manufacturer, or class",
			do:      doListShips,
			minArgs: 1,
			maxArgs: 2,
		},
		"buyship": {
			name:    "BuyShip",
			usage:   "buyShip location type",
			help:    "Buy the given ship in the specified location",
			do:      doBuyShip,
			minArgs: 2,
			maxArgs: 2,
		},
		"myships": {
			name:    "MyShips",
			usage:   "myShips [filter]",
			help:    "List owned ships, with an optional filter",
			do:      doMyShips,
			minArgs: 0,
			maxArgs: 1,
		},
	}

	doLoop(c)
}
