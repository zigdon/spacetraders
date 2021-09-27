package cli

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

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
	Aliases    []string
}

var (
	commands    map[string]cmd
	aliases     map[string]string
	allCommands []string
	mq          *msgQueue
)

func GetMsgQueue() *msgQueue {
	return mq
}

func GetCommands() (map[string]cmd, map[string]string, []string) {
	return commands, aliases, allCommands
}

func init() {
	mq = NewMessageQueue()
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
			Aliases: []string{"lsLoans"},
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
			Aliases:    []string{"lsSys"},
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
			Aliases:    []string{"lsLocations", "lsLocs"},
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
			Aliases:    []string{"lsShips"},
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
			Aliases:    []string{"go", "fly"},
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
			Aliases:    []string{"lsFlights"},
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
	aliases = make(map[string]string)
	allCommands = []string{}
	for name, cmd := range commands {
		allCommands = append(allCommands, name)
		if len(cmd.Aliases) > 0 {
			aliases[strings.ToLower(name)] = name

			for _, a := range cmd.Aliases {
				aliases[strings.ToLower(a)] = name
				allCommands = append(allCommands, a)
			}
		}
	}
}

func doHelp(c *spacetraders.Client, args []string) error {
	if len(args) > 0 {
		cmd, ok := commands[args[0]]
		if !ok {
			cmd, ok = commands[aliases[args[0]]]
		}
		if ok {
			a := ""
			if len(cmd.Aliases) > 0 {
				a = fmt.Sprintf("\nAliases: %s", strings.Join(cmd.Aliases, ", "))
			}
			Out("%s: %s\n%s%s", cmd.Name, cmd.Usage, cmd.Help, a)
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
			if len(cm.Aliases) > 0 {
				res = append(res, fmt.Sprintf("    %s (%s): %s", cm.Name, strings.Join(cm.Aliases, ", "), cm.Usage))
			} else {
				res = append(res, fmt.Sprintf("    %s: %s", cm.Name, cm.Usage))
			}
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

type msg struct {
	when time.Time
	msg  string
}

type msgQueue struct {
	mu   sync.Mutex
	msgs map[string]msg
}

func NewMessageQueue() *msgQueue {
	return &msgQueue{
		msgs: make(map[string]msg),
	}
}

func (m *msgQueue) HasMsgs() bool {
	for _, v := range m.msgs {
		if v.when.Before(time.Now()) {
			return true
		}
	}

	return false
}

func (m *msgQueue) Add(key, text string, when time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.msgs[key]; ok {
		return
	}
	m.msgs[key] = msg{
		msg:  text,
		when: when,
	}
}

func (m *msgQueue) Read() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	msgs := []string{}
	for k, v := range m.msgs {
		if v.when.After(time.Now()) {
			continue
		}
		msgs = append(msgs, v.msg)
		delete(m.msgs, k)
	}
	return msgs
}
