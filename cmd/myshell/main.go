package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

func execREPL(allowed_prompts []string) {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		prompt, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		command := strings.Split(strings.Split(prompt, "\n")[0], " ")[0]

		found := false

		for i := 0; i < len(allowed_prompts); i++ {
			if allowed_prompts[i] == command {
				// TODO
				found = true
			}
		}

		if !found {
			fmt.Printf("%s: not found\n", command)
		}
	}
}

func main() {
	allowed_prompts := make([]string, 0)

	execREPL(allowed_prompts)
}
