package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/zigdon/spacetraders"
)

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
		case "account":
			u, err := c.Account()
			if err != nil {
				log.Printf("Error: %v", err)
				break
			}
			log.Printf("%s", u)
		default:
			log.Printf("Unknown command %q.", cmd)
		}
	}
}

func main() {
	c := spacetraders.New(os.Args[1])

	if err := c.Status(); err != nil {
		log.Fatalf("Game down: %v", err)
	}

	doLoop(c)
}
