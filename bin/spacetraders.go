package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/zigdon/spacetraders"
	"github.com/zigdon/spacetraders/cli"
)

var (
	echo        = flag.Bool("echo", false, "If true, echo commands back to stdout")
	logFile     = flag.String("logfile", "/tmp/spacetraders.log", "Where should the log file be saved")
	useCache    = flag.Bool("cache", true, "If true, echo commands back to stdout")
	errorsFatal = flag.Bool("errors_fatal", false, "If false, API errors are caught")
)

// Main input loop
func loop(c *spacetraders.Client) {
	r, err := readline.New("> ")
	if err != nil {
		log.Fatalf("Can't readline: %v", err)
	}

	// Load all known ocmmands
	commands, aliases, allCommands := cli.GetCommands()

	mq := cli.GetMsgQueue()
	for {
		if mq.HasMsgs() {
			for _, m := range mq.Read() {
				cli.Out(m)
			}
		}
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
		matches := filter(allCommands, words[0], true)
		switch {
		case len(matches) == 0:
			cli.ErrMsg("Unknown command %v. Try 'help'.", words[0])
			continue
		case len(matches) > 1:
			cli.ErrMsg("%q could mean %v. Try again.", words[0], matches)
			continue
		}
		if alias, ok := aliases[matches[0]]; ok {
			words[0] = alias
		} else {
			words[0] = matches[0]
		}
		cmd, ok := commands[words[0]]
		if !ok {
			cli.ErrMsg("Unknown command %v. Try 'help'.", words[0])
		}

		if len(words)-1 < cmd.MinArgs || len(words)-1 > cmd.MaxArgs {
			cli.ErrMsg("Invalid arguments for %q", words[0])
			words = []string{"help", words[0]}
			cmd = commands["help"]
		}
		if err := validate(c, words[1:], cmd.Validators); err != nil {
			cli.ErrMsg("Invalid arguments: %v", err)
			continue
		}
		if err := cmd.Do(c, words[1:]); err != nil {
			if *errorsFatal {
				log.Fatal(err)
			}
			cli.ErrMsg("Error: %v", err)
		}
		cli.Out("")
	}
}

// Helpers
func filter(list []string, substr string, onlyPrefix bool) []string {
	res := []string{}
	lowered := strings.ToLower(substr)
	var f func(string, string) bool
	if onlyPrefix {
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
	matching := filter(validOpts, bit, false)
	switch len(matching) {
	case 0:
		return "", fmt.Errorf("No matching %ss: %v", kind, validOpts)
	case 1:
		if bit != matching[0] {
			cli.Warn("Using %q for %q", matching[0], bit)
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

	if flag.NArg() > 0 {
		if err := c.Load(flag.Arg(0)); err != nil {
			log.Fatalf("Can't login with %q: %v", flag.Arg(0), err)
		}
	}

	loop(c)
}
