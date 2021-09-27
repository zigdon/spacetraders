package cli

import (
	"fmt"
	"log"
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
	commands    = map[string]cmd{}
	aliases     = map[string]string{}
	allCommands = []string{}
	mq          *msgQueue
)

func GetMsgQueue() *msgQueue {
	return mq
}

func GetCommands() (map[string]cmd, map[string]string, []string) {
	return commands, aliases, allCommands
}

func Register(c cmd) error {
	lower := strings.ToLower
	if _, ok := commands[lower(c.Name)]; ok {
		return fmt.Errorf("there is already a %q command", c.Name)
	}

	commands[lower(c.Name)] = c
	allCommands = append(allCommands, c.Name)
	for _, a := range c.Aliases {
		aliases[lower(a)] = lower(c.Name)
		allCommands = append(allCommands, a)
	}

	return nil
}

func init() {
	mq = NewMessageQueue()
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
