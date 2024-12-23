package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var pwd string

func isValidCommand(command string, allowed []string) bool {
	for i := 0; i < len(allowed); i++ {
		if allowed[i] == command {
			// TODO
			return true
		}
	}
	return false
}

func isInPath(command string) string {
	path, err := exec.LookPath(command)
	if err != nil {
		return ""
	}

	return path
}

func execInBuiltCmd(command string, args, allowed_prompts []string) {
	switch command {
	case "echo":
		fmt.Print(strings.Join(args, " "), "\n")
	case "type":
		if len(args) > 0 && isValidCommand(args[0], allowed_prompts) {
			fmt.Printf("%s is a shell builtin\n", args[0])
		} else {
			found := isInPath(args[0])
			if found != "" {
				fmt.Printf("%s is %s\n", args[0], found)
			} else {
				fmt.Printf("%s: not found\n", args[0])
			}
		}
	case "pwd":
		fmt.Println(pwd)
	case "cd":
		if args[0][:2] == "./" {
			// TODO: relative path
		} else if args[0][:3] == "../" {
			// TODO: parent directory
		} else if args[0][0] == '/' {
			// absolute path
			err := os.Chdir(args[0])
			if err != nil {
				fmt.Printf("%s: %s: No such file or directory", command, args[0])
			}
			pwd = args[0]
		} else {
			// check if arg exists as directory in current folder
		}
	}
}

func execPathCmd(command string, args []string) {
	cmd := exec.Command(command, args...)

	var out strings.Builder
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Print(out.String())
	}
}

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

		if isValidCommand(command, allowed_prompts) {
			if command == "exit" && len(args) > 0 && args[0] == "0" {
				break
			}
			execInBuiltCmd(command, args, allowed_prompts)
		} else {
			found := isInPath(command)
			if found != "" {
				execPathCmd(command, args)
			} else {
				fmt.Printf("%s: command not found\n", command)
			}
		}
	}
}

func main() {
	allowed_prompts := []string{"exit", "echo", "type", "pwd", "cd"}

	pwd, _ = os.Getwd()

	execREPL(allowed_prompts)
}
