package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zigdon/spacetraders"
)

var ()

type cmd struct {
	Name       string
	Section    string
	Usage      string
	Validators []string
	Help       string
	Do         func(*spacetraders.Client, []string) error
	MinArgs    int
	MaxArgs    int
}

var commands = map[string]cmd{}

func GetCommands() map[string]cmd {
	return commands
}

func init() {
	commands = map[string]cmd{
		"help": {
			Name:    "Help",
			Usage:   "Help [command]",
			Help:    "List all commands, or get information on a specific command",
			Do:      doHelp,
			MaxArgs: 1,
		},

		"account": {
			Section: "Account",
			Name:    "Account",
			Usage:   "Account",
			Help:    "Get details about the logged in account",
			Do:      doAccount,
		},
		"login": {
			Section: "Account",
			Name:    "Login",
			Usage:   "Login [path/to/file]",
			Help:    "Load username and token from saved file, $HOME/.config/spacetraders.io by default",
			Do:      doLogin,
			MinArgs: 0,
			MaxArgs: 1,
		},
		"logout": {
			Section: "Account",
			Name:    "Logout",
			Usage:   "Logout",
			Help:    "Expire the current logged in token.",
			Do:      doLogout,
		},
		"claim": {
			Section: "Account",
			Name:    "Claim",
			Usage:   "Claim <username> <path/to/file>",
			Help:    "Claims a username, saves token to specified file",
			Do:      doClaim,
			MinArgs: 2,
			MaxArgs: 2,
		},

		"availableloans": {
			Section: "Loans",
			Name:    "AvailableLoans",
			Usage:   "AvailableLoans",
			Help:    "Display currently available loans",
			Do:      doLoans,
		},
		"takeloan": {
			Section: "Loans",
			Name:    "TakeLoan",
			Usage:   "TakeLoan <type>",
			Help:    "Take out one of the available loans",
			Do:      doTakeLoan,
			MinArgs: 1,
			MaxArgs: 1,
		},
		"myloans": {
			Section: "Loans",
			Name:    "MyLoans",
			Usage:   "MyLoans",
			Help:    "List outstanding loans",
			Do:      doMyLoans,
		},

		"system": {
			Section:    "Locations",
			Name:       "System",
			Usage:      "System [system]",
			Validators: []string{"system"},
			Help:       "Get details about a system, or all systems if not specified",
			Do:         doListSystems,
			MaxArgs:    1,
		},
		"locations": {
			Section:    "Locations",
			Name:       "Locations",
			Usage:      "Locations <system> [type]",
			Validators: []string{"system"},
			Help:       "Show all locations in a system",
			Do:         doListLocations,
			MinArgs:    1,
			MaxArgs:    2,
		},

		"listships": {
			Section:    "Ships",
			Name:       "ListShips",
			Usage:      "ListShips <system> [filter]",
			Validators: []string{"system"},
			Help: "Show available ships at all the locations in a system. If filter is provided, " +
				"only show ships that match in type, manufacturer, or class",
			Do:      doListShips,
			MinArgs: 1,
			MaxArgs: 2,
		},
		"buyship": {
			Section:    "Ships",
			Name:       "BuyShip",
			Usage:      "BuyShip <location> <type>",
			Validators: []string{"location"},
			Help:       "Buy the given ship in the specified location",
			Do:         doBuyShip,
			MinArgs:    2,
			MaxArgs:    2,
		},
		"myships": {
			Section:    "Ships",
			Name:       "MyShips",
			Usage:      "MyShips [filter]",
			Validators: []string{"ships"},
			Help:       "List owned ships, with an optional filter",
			Do:         doMyShips,
			MinArgs:    0,
			MaxArgs:    1,
		},

		"createflightplan": {
			Section:    "Flight Plans",
			Name:       "CreateFlightPlan",
			Usage:      "CreateFlightPlan <shipID> <destination>",
			Validators: []string{"ships", "location"},
			Help:       "Create a flight plan for given ship to specified destination",
			Do:         doCreateFlight,
			MinArgs:    2,
			MaxArgs:    2,
		},
		"showflightplan": {
			Section:    "Flight Plans",
			Name:       "ShowFlightPlan",
			Usage:      "ShowFlightPlan <flightPlanID>",
			Validators: []string{"flights"},
			Help:       "Show the flight plan identified",
			Do:         doShowFlight,
			MinArgs:    1,
			MaxArgs:    1,
		},
		"wait": {
			Section:    "Flight Plans",
			Name:       "Wait",
			Usage:      "Wait <flightPlanID>",
			Validators: []string{"flights"},
			Help:       "Wait until specified flight arrives",
			Do:         doWaitForFlight,
			MinArgs:    1,
			MaxArgs:    1,
		},

		"buy": {
			Section:    "Goods and Cargo",
			Name:       "Buy",
			Usage:      "Buy <shipID> <good> <quantity>",
			Validators: []string{"ships"},
			Help:       "Buy the specified quantiy of good for the ship identified",
			Do:         doBuy,
			MinArgs:    3,
			MaxArgs:    3,
		},
		"sell": {
			Section:    "Goods and Cargo",
			Name:       "Sell",
			Usage:      "Sell <shipID> <good> <quantity>",
			Validators: []string{"ships"},
			Help:       "Sell the specified quantiy of good from the ship identified",
			Do:         doSell,
			MinArgs:    3,
			MaxArgs:    3,
		},
		"market": {
			Section:    "Goods and Cargo",
			Name:       "Market",
			Usage:      "Market <location>",
			Validators: []string{"mylocation"},
			Help:       "List all goods offered at location.",
			Do:         doMarket,
			MinArgs:    1,
			MaxArgs:    1,
		},
	}
}

func doHelp(c *spacetraders.Client, args []string) error {
	if len(args) > 0 {
		cmd, ok := commands[args[0]]
		if ok {
			Out("%s: %s\n%s", cmd.Name, cmd.Usage, cmd.Help)
			return nil
		}
	}
	cmds := make(map[string][]cmd)
	for _, cmd := range commands {
		cmds[cmd.Section] = append(cmds[cmd.Section], cmd)
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
		sort.SliceStable(cmds[s], func(i, j int) bool { return cmds[s][i].Name < cmds[s][j].Name })
		for _, cm := range cmds[s] {
			res = append(res, fmt.Sprintf("    %s: %s", cm.Name, cm.Usage))
		}
		res = append(res, "")
	}
	Out(strings.Join(res, "\n"))
	return nil
}

func ErrMsg(format string, args ...interface{}) {
	for _, l := range strings.Split(fmt.Sprintf(format, args...), "\n") {
		fmt.Printf("! %s\n", l)
	}
}

func Warn(format string, args ...interface{}) {
	for _, l := range strings.Split(fmt.Sprintf(format, args...), "\n") {
		fmt.Printf("* %s\n", l)
	}
}

func Out(format string, args ...interface{}) {
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
