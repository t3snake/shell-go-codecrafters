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
	var command string
	var args []string

	for {
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		prompt_newline, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		prompt := strings.Split(strings.Split(prompt_newline, "\n")[0], " ")

		command = prompt[0]
		args = prompt[1:]

		found := false

		for i := 0; i < len(allowed_prompts); i++ {
			if allowed_prompts[i] == command {
				// TODO
				found = true
			}
		}

		if found {
			if command == "exit" && len(args) > 0 && args[0] == "0" {
				break
			} else if command == "echo" {
				fmt.Print(strings.Join(args, " "), "\n")
			}
		} else {
			fmt.Printf("%s: not found\n", command)
		}
	}
}

func main() {
	allowed_prompts := []string{"exit", "echo"}

	execREPL(allowed_prompts)
}
