package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

var pwd string

const REDIR_MODE_1 int = 1
const REDIR_MODE_2 int = 2
const REDIR_MODE_APPEND_1 int = 3
const REDIR_MODE_APPEND_2 int = 4

func isValidEscChar(char byte) bool {
	return char == '$' || char == '"' || char == '\\'
}

func parseArgs(line string) []string {
	args := make([]string, 0)
	var in_single_quotes bool
	var in_double_quotes bool
	var is_escaped bool

	var tmp_arg string

	for idx, char := range line {
		if is_escaped {
			tmp_arg += string(char)
			is_escaped = false
			continue
		}

		if char == '"' && !in_single_quotes {
			in_double_quotes = !in_double_quotes
		} else if char == '\'' && !in_double_quotes {
			in_single_quotes = !in_single_quotes
		} else if char == '\\' && !in_single_quotes {
			if !in_double_quotes {
				is_escaped = true
			} else if in_double_quotes && idx < len(line)-1 && isValidEscChar(line[idx+1]) {
				is_escaped = true
			} else {
				tmp_arg += string(char)
			}
		} else if char == ' ' && !in_single_quotes && !in_double_quotes {
			if tmp_arg != "" {
				args = append(args, tmp_arg)
			}
			tmp_arg = ""
		} else {
			tmp_arg += string(char)
		}
	}
	args = append(args, tmp_arg)

	return args
}

func fileForRedirect(args []string) (string, int, int) {
	for idx, arg := range args {
		if arg == ">" || arg == "1>" {
			if idx+1 < len(args) {
				return args[idx+1], idx, REDIR_MODE_1
			}
		} else if arg == "2>" {
			if idx+1 < len(args) {
				return args[idx+1], idx, REDIR_MODE_2
			}
		} else if arg == ">>" || arg == "1>>" {
			if idx+1 < len(args) {
				return args[idx+1], idx, REDIR_MODE_APPEND_1
			}
		} else if arg == "2>>" {
			if idx+1 < len(args) {
				return args[idx+1], idx, REDIR_MODE_APPEND_2
			}
		}
	}
	return "", -1, -1
}

func writeResultToFile(result, file string, is_append bool) {
	var r_file *os.File
	var err error

	if is_append {
		r_file, err = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	} else {
		r_file, err = os.Create(file)
	}

	if err != nil {
		log.Fatal(err)
	}
	defer r_file.Close()

	writer := bufio.NewWriter(r_file)
	fmt.Fprint(writer, result)

	writer.Flush()
}

func isValidCommand(command string, allowed []string) bool {
	for i := 0; i < len(allowed); i++ {
		if allowed[i] == command {
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

func updatePwdIfExists(new_path, command string) string {
	//nested .. case
	if new_path[len(new_path)-2:] == ".." || new_path[len(new_path)-3:] == "../" {
		updatePwdIfExists(path.Dir(new_path), command)
		return ""
	}

	err := os.Chdir(new_path)
	if err != nil {
		return fmt.Sprintf("%s: %s: No such file or directory\n", command, new_path)
	}
	pwd = new_path
	return ""
}

func resolvePathForCd(new_path string) string {
	if new_path[0] == '~' {
		// home path
		home_path := os.Getenv("HOME")
		new_path = strings.Replace(new_path, "~", home_path, 1)
	} else if new_path[:2] == "./" {
		// relative path
		new_path = pwd + new_path[1:]
	} else if new_path == ".." {
		// just parent directory
		new_path = path.Dir(pwd)
	} else if new_path[:3] == "../" {
		// parent directory + further
		new_pwd := pwd
		for len(new_path) > 2 && new_path[:3] == "../" {
			// ../../(..abc) case
			new_pwd = path.Dir(new_pwd)
			new_path = new_path[3:]
		}
		if len(new_path) > 0 {
			new_path = new_pwd + "/" + new_path
		} else {
			new_path = new_pwd
		}
	} else if new_path[0] == '/' {
		// absolute path new_path remains same
	} else {
		// check if arg exists as directory in current folder
		new_path = pwd + "/" + new_path
	}
	return new_path
}

func execInBuiltCmd(command string, args, allowed_prompts []string) string {
	switch command {
	case "echo":
		return fmt.Sprint(strings.Join(args, " "), "\n")

	case "type":
		if len(args) > 0 && isValidCommand(args[0], allowed_prompts) {
			return fmt.Sprintf("%s is a shell builtin\n", args[0])
		} else {
			found := isInPath(args[0])
			if found != "" {
				return fmt.Sprintf("%s is %s\n", args[0], found)
			} else {
				return fmt.Sprintf("%s: not found\n", args[0])
			}
		}

	case "pwd":
		return fmt.Sprintln(pwd)

	case "cd":
		if len(args) != 1 {
			return fmt.Sprintln("Insufficient arguments")
		}

		new_path := args[0]
		new_path = resolvePathForCd(new_path)
		return updatePwdIfExists(new_path, command)
	}
	return ""
}

// return stdout and stderr as string
func execPathCmd(command string, args []string) (string, string) {
	cmd := exec.Command(command, args...)

	var out strings.Builder
	var err_out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &err_out

	_ = cmd.Run()

	return out.String(), err_out.String()
}

func execREPL(allowed_prompts []string) {
	var command string
	var args []string

	var is_print_to_file bool
	var result string
	var result_err string

	for {
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		prompt_newline, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		line := strings.Split(prompt_newline, "\n")[0]

		all_args := parseArgs(line)
		command = all_args[0]
		args = all_args[1:]

		redirect_file, idx, redir_mode := fileForRedirect(args)
		is_print_to_file = redirect_file != ""

		if is_print_to_file {
			args = args[:idx]
		}

		if isValidCommand(command, allowed_prompts) {
			if command == "exit" && len(args) > 0 && args[0] == "0" {
				break
			}
			result = execInBuiltCmd(command, args, allowed_prompts)
		} else {
			found := isInPath(command)
			if found != "" {
				result, result_err = execPathCmd(command, args)
			} else {
				result = fmt.Sprintf("%s: command not found\n", command)
			}
		}

		if is_print_to_file {
			if redir_mode == REDIR_MODE_1 || redir_mode == REDIR_MODE_APPEND_1 {
				// Print stderr and write stdout
				writeResultToFile(result, redirect_file, redir_mode == REDIR_MODE_APPEND_1)

				if result_err != "" {
					fmt.Print(result_err)
				}
			} else if redir_mode == REDIR_MODE_2 || redir_mode == REDIR_MODE_APPEND_2 {
				// Print stdout and write stderr
				writeResultToFile(result_err, redirect_file, redir_mode == REDIR_MODE_APPEND_2)

				if result != "" {
					fmt.Print(result)
				}
			}
		} else {
			fmt.Print(result)
		}

		// clear cache
		result = ""
		result_err = ""
		is_print_to_file = false
	}
}

func main() {
	allowed_prompts := []string{"exit", "echo", "type", "pwd", "cd"}

	pwd, _ = os.Getwd()

	execREPL(allowed_prompts)
}
