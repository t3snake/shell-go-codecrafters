package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
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
	allowed_prompts := []string{"exit", "echo", "type", "pwd"}

	_, filename, _, _ := runtime.Caller(0)
	// pwd would be same as root of current project if called with .sh file
	// current file is in root/cmd/myshell so remove cmd/myshell to get root
	build_dir := path.Dir(filename)
	pwd = strings.ReplaceAll(build_dir, "cmd/myshell", "")

	execREPL(allowed_prompts)
}
