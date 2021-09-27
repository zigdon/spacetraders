package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/zigdon/spacetraders"
)

var (
	echo        = flag.Bool("echo", false, "If true, echo commands back to stdout")
	useCache    = flag.Bool("cache", true, "If true, echo commands back to stdout")
	errorsFatal = flag.Bool("errors_fatal", false, "If false, API errors are caught")
	logFile     = flag.String("logfile", "/tmp/spacetraders.log", "Where should the log file be saved")
)

type cmd struct {
	name       string
	section    string
	usage      string
	validators []string
	help       string
	do         func(*spacetraders.Client, []string) error
	minArgs    int
	maxArgs    int
}

var commands = map[string]cmd{}

// Main input loop
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
		if line == "" {
			continue
		}
		if *echo {
			fmt.Printf("> %s\n", line)
		}

		words := strings.Split(strings.TrimSpace(line), " ")
		switch cmd := strings.ToLower(words[0]); cmd {
		case "exit":
			return
		default:
			if cmd, ok := commands[words[0]]; ok {
				if len(words)-1 < cmd.minArgs || len(words)-1 > cmd.maxArgs {
					errMsg("Invalid arguments for %q", words[0])
					words = []string{"help", words[0]}
					cmd = commands["help"]
				}
				if err := validate(c, words[1:], cmd.validators); err != nil {
					errMsg("Invalid arguments: %v", err)
					continue
				}
				if err := cmd.do(c, words[1:]); err != nil {
					if *errorsFatal {
						log.Fatal(err)
					}
					errMsg("Error: %v", err)
				}
				out("")
				continue
			}
			errMsg("Unknown command %v. Try 'help'.", words)
		}
	}
}

// Command implementations
func doHelp(c *spacetraders.Client, args []string) error {
	if len(args) > 0 {
		cmd, ok := commands[args[0]]
		if ok {
			out("%s: %s\n%s", cmd.name, cmd.usage, cmd.help)
			return nil
		}
	}
	cmds := make(map[string][]cmd)
	for _, cmd := range commands {
		cmds[cmd.section] = append(cmds[cmd.section], cmd)
	}
	res := []string{
		"Available commands:",
		"<arguments> are required, [options] are optional.",
		"",
	}
	for _, s := range []string{"", "Account", "Loans", "Ships", "Flight Plans", "Locations", "Goods and Cargo"} {
		if s != "" {
			res = append(res, fmt.Sprintf("  %s:", s))
		}
		sort.SliceStable(cmds[s], func(i, j int) bool { return cmds[s][i].name < cmds[s][j].name })
		for _, cm := range cmds[s] {
			res = append(res, fmt.Sprintf("    %s: %s", cm.name, cm.usage))
		}
		res = append(res, "")
	}
	out(strings.Join(res, "\n"))
	return nil
}

func doAccount(c *spacetraders.Client, args []string) error {
	u, err := c.Account()
	if err != nil {
		return err
	}
	out("%s", u)
	return nil
}

func doLogin(c *spacetraders.Client, args []string) error {
	path := filepath.Join(os.Getenv("HOME"), ".config/spacetraders.io")
	if len(args) > 0 {
		path = args[0]
	}
	if err := c.Load(path); err != nil {
		errMsg("Error loading token: %v", err)
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
		out("amt: %d, needs collateral: %v, rate: %d, term (days): %d, type: %s",
			l.Amount, l.CollateralRequired, l.Rate, l.TermInDays, l.Type)
	}

	return nil
}

func doTakeLoan(c *spacetraders.Client, args []string) error {
	loan, err := c.TakeLoan(args[0])
	if err != nil {
		return fmt.Errorf("error taking out loan: %v", err)
	}

	out("Loan taken, %s (%s), due: %s (in %s)",
		loan.ShortID, loan.ID, loan.Due.Local(), loan.Due.Sub(time.Now()).Truncate(time.Second))

	return nil
}

func doMyLoans(c *spacetraders.Client, args []string) error {
	loans, err := c.MyLoans()
	if err != nil {
		return fmt.Errorf("error querying loans: %v", err)
	}

	for _, l := range loans {
		out(l.String())
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
		out("All systems:")
		for _, sym := range sys {
			out(cache[sym].String())
		}
		return nil
	}

	out(cache[args[0]].Details(0))
	return nil
}

func doListLocations(c *spacetraders.Client, args []string) error {
	filter := ""
	if len(args) > 1 {
		filter = args[1]
	}
	locs, err := c.ListLocations(args[0], filter)
	if err != nil {
		return fmt.Errorf("error listing locations in %q: %v", args[0], err)
	}

	out("%d locations in %q:", len(locs), args[0])
	for _, l := range locs {
		out(l.Details(1))
	}

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
			out(s.String())
		}
		return nil
	}

	for _, s := range ships {
		out(s.String())
	}

	return nil
}

func doBuyShip(c *spacetraders.Client, args []string) error {
	ship, err := c.BuyShip(args[0], args[1])
	if err != nil {
		return fmt.Errorf("error buying ship %q at %q: %v", args[1], args[0], err)
	}

	out("New ship ID: %s (%s)", ship.ShortID, ship.ID)

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
		out("No ships found.")
	case 1:
		out(res[0].String())
	default:
		for _, s := range res {
			out(s.Short())
		}
	}

	return nil
}

func doCreateFlight(c *spacetraders.Client, args []string) error {
	flight, err := c.CreateFlight(args[0], args[1])
	if err != nil {
		return fmt.Errorf("error creating flight plan to %q: %v", args[1], err)
	}

	out("Created flight plan: %s", flight.Short())

	return nil
}

func doShowFlight(c *spacetraders.Client, args []string) error {
	flight, err := c.ShowFlight(args[0])
	if err != nil {
		return fmt.Errorf("error listing flight plan %q: %v", args[0], err)
	}

	out(flight.String())

	return nil
}

func doWaitForFlight(c *spacetraders.Client, args []string) error {
	flight, err := c.ShowFlight(args[0])
	if err != nil {
		return fmt.Errorf("error listing flight plan %q: %v", args[0], err)
	}

	if flight.ArrivesAt.Before(time.Now()) {
		return fmt.Errorf("flight %s (%s) already arrived", flight.ShortID, flight.ID)
	}

	// TODO: Allow interrupting the wait
	delay := flight.ArrivesAt.Sub(time.Now()).Truncate(time.Second)
	out("Waiting %s for %s (%s) to arrive...", delay, flight.ShortID, flight.ID)
	select {
	case <-time.After(delay):
	case <-time.After(time.Minute):
		out("... still waiting for %s", flight.ShortID)
	}
	out("... %s arrived!", flight.ShortID)

	return nil
}

func doBuy(c *spacetraders.Client, args []string) error {
	qty, err := strconv.Atoi(args[2])
	if err != nil {
		return err
	}

	order, err := c.BuyCargo(args[0], args[1], qty)
	if err != nil {
		return fmt.Errorf("error selling goods: %v", err)
	}

	out("Bought %d of %s for %d", order.Quantity, order.Good, order.Total)

	return nil
}

func doSell(c *spacetraders.Client, args []string) error {
	qty, err := strconv.Atoi(args[2])
	if err != nil {
		return err
	}

	order, err := c.SellCargo(args[0], args[1], qty)
	if err != nil {
		return fmt.Errorf("error selling goods: %v", err)
	}

	out("Sold %d of %s for %d", order.Quantity, order.Good, order.Total)

	return nil
}

func doMarket(c *spacetraders.Client, args []string) error {
	offers, err := c.Marketplace(args[0])
	if err != nil {
		return fmt.Errorf("error querying marketplace at %q: %v", args[0], err)
	}

	out("%d offers at %q:", len(offers), args[0])
	for _, offer := range offers {
		out(offer.String())
	}

	return nil
}

// Helpers
func errMsg(format string, args ...interface{}) {
	for _, l := range strings.Split(fmt.Sprintf(format, args...), "\n") {
		fmt.Printf("! %s\n", l)
	}
}

func warn(format string, args ...interface{}) {
	for _, l := range strings.Split(fmt.Sprintf(format, args...), "\n") {
		fmt.Printf("* %s\n", l)
	}
}

func out(format string, args ...interface{}) {
	if format == "" {
		fmt.Println()
		return
	}
	for i, l := range strings.Split(fmt.Sprintf(format, args...), "\n") {
		if i == 0 {
			fmt.Printf("- %s\n", l)
			continue
		}
		fmt.Printf("  %s\n", l)
	}
}

func filter(list []string, filter string) []string {
	res := []string{}
	lowered := strings.ToLower(filter)
	for _, s := range list {
		// If it's an exact match, don't bother with the rest
		if strings.ToLower(s) == lowered {
			return []string{s}
		}
		if strings.Contains(strings.ToLower(s), lowered) {
			res = append(res, s)
		}
	}

	return res
}

func valid(c *spacetraders.Client, kind spacetraders.CacheKey, bit string) (string, error) {
	validOpts := c.Restore(kind)
	matching := filter(validOpts, bit)
	switch len(matching) {
	case 0:
		return "", fmt.Errorf("No matching %ss: %v", kind, validOpts)
	case 1:
		if bit != matching[0] {
			warn("Using %q for %q", matching[0], bit)
		}
		return matching[0], nil
	default:
		return "", fmt.Errorf("Multiple matching %ss: %v", kind, matching)
	}
}

func validate(c *spacetraders.Client, words []string, validators []string) error {
	if !*useCache || len(words) == 0 {
		return nil
	}
	msgs := []string{}
	for i, v := range validators {
		if len(words) < i-1 {
			return nil
		}
		var ck spacetraders.CacheKey
		switch v {
		case "mylocation":
			ck = spacetraders.MYLOCATIONS
		case "location":
			ck = spacetraders.LOCATIONS
		case "system":
			ck = spacetraders.SYSTEMS
		case "ship":
			ck = spacetraders.SHIPS
		case "flights":
			ck = spacetraders.FLIGHTS
		default:
			continue
		}
		match, err := valid(c, ck, words[i])
		if err != nil {
			msgs = append(msgs, fmt.Sprintf("Invalid %s %q: %v", v, words[i], err))
			continue
		}
		words[i] = match
	}

	if len(msgs) > 0 {
		return fmt.Errorf("validation errors:\n%s", strings.Join(msgs, "\n"))
	}

	return nil
}

func main() {
	flag.Parse()
	f, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Can't open log file %q: %v", *logFile, err)
	}
	defer f.Close()
	log.SetOutput(f)
	c := spacetraders.New()

	if err := c.Status(); err != nil {
		log.Fatalf("Game down: %v", err)
	}

	commands = map[string]cmd{
		"help": {
			name:    "Help",
			usage:   "Help [command]",
			help:    "List all commands, or get information on a specific command",
			do:      doHelp,
			maxArgs: 1,
		},

		"account": {
			section: "Account",
			name:    "Account",
			usage:   "Account",
			help:    "Get details about the logged in account",
			do:      doAccount,
		},
		"login": {
			section: "Account",
			name:    "Login",
			usage:   "Login [path/to/file]",
			help:    "Load username and token from saved file, $HOME/.config/spacetraders.io by default",
			do:      doLogin,
			minArgs: 0,
			maxArgs: 1,
		},
		"logout": {
			section: "Account",
			name:    "Logout",
			usage:   "Logout",
			help:    "Expire the current logged in token.",
			do:      doLogout,
		},
		"claim": {
			section: "Account",
			name:    "Claim",
			usage:   "Claim <username> <path/to/file>",
			help:    "Claims a username, saves token to specified file",
			do:      doClaim,
			minArgs: 2,
			maxArgs: 2,
		},

		"availableloans": {
			section: "Loans",
			name:    "AvailableLoans",
			usage:   "AvailableLoans",
			help:    "Display currently available loans",
			do:      doLoans,
		},
		"takeloan": {
			section: "Loans",
			name:    "TakeLoan",
			usage:   "TakeLoan <type>",
			help:    "Take out one of the available loans",
			do:      doTakeLoan,
			minArgs: 1,
			maxArgs: 1,
		},
		"myloans": {
			section: "Loans",
			name:    "MyLoans",
			usage:   "MyLoans",
			help:    "List outstanding loans",
			do:      doMyLoans,
		},

		"system": {
			section:    "Locations",
			name:       "System",
			usage:      "System [system]",
			validators: []string{"system"},
			help:       "Get details about a system, or all systems if not specified",
			do:         doListSystems,
			maxArgs:    1,
		},
		"locations": {
			section:    "Locations",
			name:       "Locations",
			usage:      "Locations <system> [type]",
			validators: []string{"system"},
			help:       "Show all locations in a system",
			do:         doListLocations,
			minArgs:    1,
			maxArgs:    2,
		},

		"listships": {
			section:    "Ships",
			name:       "ListShips",
			usage:      "ListShips <system> [filter]",
			validators: []string{"system"},
			help: "Show available ships at all the locations in a system. If filter is provided, " +
				"only show ships that match in type, manufacturer, or class",
			do:      doListShips,
			minArgs: 1,
			maxArgs: 2,
		},
		"buyship": {
			section:    "Ships",
			name:       "BuyShip",
			usage:      "BuyShip <location> <type>",
			validators: []string{"location"},
			help:       "Buy the given ship in the specified location",
			do:         doBuyShip,
			minArgs:    2,
			maxArgs:    2,
		},
		"myships": {
			section:    "Ships",
			name:       "MyShips",
			usage:      "MyShips [filter]",
			validators: []string{"ships"},
			help:       "List owned ships, with an optional filter",
			do:         doMyShips,
			minArgs:    0,
			maxArgs:    1,
		},

		"createflightplan": {
			section:    "Flight Plans",
			name:       "CreateFlightPlan",
			usage:      "CreateFlightPlan <shipID> <destination>",
			validators: []string{"ships", "location"},
			help:       "Create a flight plan for given ship to specified destination",
			do:         doCreateFlight,
			minArgs:    2,
			maxArgs:    2,
		},
		"showflightplan": {
			section:    "Flight Plans",
			name:       "ShowFlightPlan",
			usage:      "ShowFlightPlan <flightPlanID>",
			validators: []string{"flights"},
			help:       "Show the flight plan identified",
			do:         doShowFlight,
			minArgs:    1,
			maxArgs:    1,
		},
		"wait": {
			section:    "Flight Plans",
			name:       "Wait",
			usage:      "Wait <flightPlanID>",
			validators: []string{"flights"},
			help:       "Wait until specified flight arrives",
			do:         doWaitForFlight,
			minArgs:    1,
			maxArgs:    1,
		},

		"buy": {
			section:    "Goods and Cargo",
			name:       "Buy",
			usage:      "Buy <shipID> <good> <quantity>",
			validators: []string{"ships"},
			help:       "Buy the specified quantiy of good for the ship identified",
			do:         doBuy,
			minArgs:    3,
			maxArgs:    3,
		},
		"sell": {
			section:    "Goods and Cargo",
			name:       "Sell",
			usage:      "Sell <shipID> <good> <quantity>",
			validators: []string{"ships"},
			help:       "Sell the specified quantiy of good from the ship identified",
			do:         doSell,
			minArgs:    3,
			maxArgs:    3,
		},
		"market": {
			section:    "Goods and Cargo",
			name:       "Market",
			usage:      "Market <location>",
			validators: []string{"mylocation"},
			help:       "List all goods offered at location.",
			do:         doMarket,
			minArgs:    1,
			maxArgs:    1,
		},
	}

	doLoop(c)
}
