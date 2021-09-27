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

	loop(c)
}

// Main input loop
func loop(c *spacetraders.Client) {
	r, err := readline.New("> ")
	if err != nil {
		log.Fatalf("Can't readline: %v", err)
	}
	commands := cli.GetCommands()
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
				continue
			}
			cli.ErrMsg("Unknown command %v. Try 'help'.", words)
		}
	}
}

// Helpers
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
