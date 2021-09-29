package cli

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/zigdon/spacetraders"
)

var (
	useCache = flag.Bool("cache", true, "If true, echo commands back to stdout")
)

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
	commands    = map[string]*cmd{}
	aliases     = map[string]string{}
	allCommands = []string{}
)

func Register(c cmd) error {
	lower := strings.ToLower
	if _, ok := commands[lower(c.Name)]; ok {
		return fmt.Errorf("there is already a %q command", c.Name)
	}

	commands[lower(c.Name)] = &c
	allCommands = append(allCommands, c.Name)
	for _, a := range c.Aliases {
		aliases[lower(a)] = lower(c.Name)
		allCommands = append(allCommands, a)
	}

	return nil
}

func ParseLine(c *spacetraders.Client, line string) (*cmd, []string, error) {
	words := strings.Split(strings.TrimSpace(line), " ")
	matches := filter(allCommands, words[0], filterPrefix)
	switch {
	case len(matches) == 0:
		return nil, nil, fmt.Errorf("Unknown command %v. Try 'help'.", words[0])
	case len(matches) > 1:
		return nil, nil, fmt.Errorf("%q could mean %v. Try again.", words[0], matches)
	}
	if alias, ok := aliases[matches[0]]; ok {
		words[0] = alias
	} else {
		words[0] = matches[0]
	}
	cmd, ok := commands[strings.ToLower(words[0])]
	if !ok {
		log.Fatalf("Command %q not found!", words[0])
	}
	if len(words)-1 < cmd.MinArgs || len(words)-1 > cmd.MaxArgs {
		ErrMsg("Invalid arguments for %q", words[0])
		words = []string{"help", words[0]}
		cmd = commands["help"]
	}

	args := words[1:]
	if err := validate(c, args, cmd.Validators); err != nil {
		return nil, nil, fmt.Errorf("Invalid arguments: %v", err)
	}

	return cmd, args, nil
}

type filterType bool

var filterPrefix filterType = true
var filterContains filterType = false

func filter(list []string, substr string, kind filterType) []string {
	res := []string{}
	lowered := strings.ToLower(substr)
	var f func(string, string) bool
	if kind == filterPrefix {
		f = strings.HasPrefix
	} else {
		f = strings.Contains
	}
	for _, s := range list {
		// If it's an exact match, don't bother with the rest
		if strings.ToLower(s) == lowered {
			return []string{s}
		}
		if f(strings.ToLower(s), lowered) {
			res = append(res, s)
		}
	}

	return res
}

func valid(c *spacetraders.Client, kind spacetraders.CacheKey, bit string) (string, error) {
	validOpts := c.Restore(kind)
	matching := filter(validOpts, bit, filterContains)
	switch len(matching) {
	case 0:
		return "", fmt.Errorf("No matching %ss: %v", kind, validOpts)
	case 1:
		if bit != matching[0] {
			Warn("Using %q for %q", matching[0], bit)
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

func init() {
	for _, c := range []cmd{
		{
			Name:    "Help",
			Usage:   "Help [command]",
			Help:    "List all commands, or get information on a specific command",
			Do:      doHelp,
			MaxArgs: 1,
		},
	} {
		if err := Register(c); err != nil {
			log.Fatalf("Can't register %q: %v", c.Name, err)
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
	cmds := make(map[string][]*cmd)
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
		log.Printf("[err] %s", l)
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
