package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zigdon/spacetraders"
)

type cmd struct {
	usage string
	help  string
	do    func(*spacetraders.Client, []string) error
}

var commands = map[string]cmd{}

func doLoop(c *spacetraders.Client) {
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("> ")
		line, err := r.ReadString(byte('\n'))
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Printf("Error while reading input: %v", err)
			break
		}

		words := strings.Split(strings.TrimSpace(line), " ")
		switch cmd := words[0]; cmd {
		case "exit":
			return
		default:
			if cmd, ok := commands[words[0]]; ok {
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
	for cmd := range commands {
		cmds = append(cmds, cmd)
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
	log.Printf("%s", u)
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
		log.Printf("amt: %d, needs collateral: %v, rate: %d, term (days): %d, type: %s",
			l.Amount, l.CollateralRequired, l.Rate, l.TermInDays, l.Type)
	}

	return nil
}

func doTakeLoan(c *spacetraders.Client, args []string) error {
	loan, err := c.TakeLoan(args[0])
	if err != nil {
		return fmt.Errorf("error taking out loan: %v", err)
	}

	log.Printf("Loan taken, id=%s, due: %s", loan.ID, loan.Due)

	return nil
}

func doMyLoans(c *spacetraders.Client, args []string) error {
	loans, err := c.MyLoans()
	if err != nil {
		return fmt.Errorf("error querying loans: %v", err)
	}

	for _, l := range loans {
		log.Printf("id: %s, due: %s, amt: %d, status: %s, type: %s",
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
		log.Println("All systems:")
		for _, sym := range sys {
			log.Println(cache[sym])
		}
		return nil
	}

	log.Println(cache[args[0]].Details())
	return nil
}

func main() {
	c := spacetraders.New()

	if err := c.Status(); err != nil {
		log.Fatalf("Game down: %v", err)
	}

	commands = map[string]cmd{
		"help": {
			usage: "help [command]",
			help:  "List all commands, or get information on a specific command",
			do:    doHelp,
		},

		"account": {
			usage: "account",
			help:  "Get details about the logged in account",
			do:    doAccount,
		},
		"login": {
			usage: "login path/to/file",
			help:  "Load username and token from saved file, $HOME/.config/spacetraders.io by default",
			do:    doLogin,
		},
		"logout": {
			usage: "logout",
			help:  "Expire the current logged in token.",
			do:    doLogout,
		},
		"claim": {
			usage: "claim username path/to/file",
			help:  "Claims a username, saves token to specified file",
			do:    doClaim,
		},

		"availableLoans": {
			usage: "availableLoans",
			help:  "Display currently available loans",
			do:    doLoans,
		},
		"takeLoan": {
			usage: "takeLoan type",
			help:  "Take out one of the available loans",
			do:    doTakeLoan,
		},
		"myLoans": {
			usage: "myLoans",
			help:  "List outstanding loans",
			do:    doMyLoans,
		},

		"system": {
			usage: "system [symbol]",
			help:  "Get details about a system, or all systems if not specified",
			do:    doListSystems,
		},

		"listShips": {},
	}

	doLoop(c)
}
