package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"golang.org/x/term"
)

// ANSI escape codes

const SAVE_CURSOR_POS = "\033[s"
const RESTORE_CURSOR_POS = "\033[u"
const MOVE_CURSOR_TO_BEG = "\033[2G"  // for input line its column 2 (after '$ ')
const MOVE_CURSOR_X_LEFT = "\033[%dD" // formatter: move cursor left %d times
const TERMINAL_BELL = "\x07"
const TERMINAL_UP = "\x1b[A"
const TERMINAL_DOWN = "\x1b[B"

// Implementation of simple GNU readline in raw terminal mode
func terminalReadLine(auto_completion_db *PrefixTreeNode, history []HistoryEntry) (string, error) {
	// -1 didnt start navigation, else index of history
	cur_history := -1
	var temp_history = []byte("")

	// change terminal to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	// change terminal back to cooked mode
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	var current_buffer []byte = make([]byte, 0)
	input_char := make([]byte, 1)

	// successive tab count
	var tab_count = 0

	for {
		n, err := os.Stdin.Read(input_char)
		if err != nil || n == 0 {
			return "", err
		}

		typed_character := input_char[0]

		// handle successive tab count
		if typed_character == '\t' {
			tab_count++
		} else {
			tab_count = 0
		}

		if typed_character == '\t' {
			// tab handling (\t or 9)
			prefix := calculateLastWord(current_buffer)
			results := searchPrefixTree(prefix, auto_completion_db)

			if len(results) == 0 {
				fmt.Print(TERMINAL_BELL)
				tab_count = 0
				continue
			}

			if len(results) == 1 {
				len_prev_buffer := len(current_buffer)
				// add a space at the end after completion
				current_buffer = replaceLastWord(current_buffer, []byte(results[0]+" "))
				redrawBuffer(current_buffer, len_prev_buffer)
				tab_count = 0
				continue
			}

			// check common denominator for results
			denominator := getCompletionDenominator(prefix, results)
			if len(denominator) > 0 {
				for _, byt := range denominator {
					echoLetterAndAppendToBuffer(byt, &current_buffer)
				}

				tab_count = 0
			} else if tab_count == 1 {
				fmt.Print(TERMINAL_BELL)
			} else if tab_count == 2 {
				fmt.Print("\r\n")
				for _, result := range results {
					fmt.Printf("%s  ", result)
				}
				fmt.Print("\r\n")
				fmt.Print("$ ")
				fmt.Print(string(current_buffer))

				tab_count = 0
			}

		} else if typed_character == '\b' || typed_character == 127 {
			// backspace handling (\b or 8) or terminal might give (del or 127)

			// TODO delete key handling, print(" \b")
			// space overrides key after cursor and \b moves back cursor to original position

			if len(current_buffer) > 0 {
				current_buffer = current_buffer[:len(current_buffer)-1]
				// \b moves cursor backwards, space overrides and move cursor back again
				fmt.Print("\b \b")
			}
		} else if typed_character == 3 {
			// ctrl+c or sigint handling (3)
			fmt.Print("\r\n")
			return "", fmt.Errorf("SIGINT")
		} else if typed_character == '\n' || typed_character == '\r' {
			// return on line feed (LF) (\n or 10) or carriage return (CR) (\r or 13)
			fmt.Print("\r\n")

			// reset history navigation
			cur_history = -1

			return string(current_buffer), nil
		} else if typed_character == 27 { // arrow keys
			// (Left, Right, Up, Down) are (27 91 68, 27 91 67, 27 91 65, 27 91 66).
			next_bytes := make([]byte, 2)
			n, err := os.Stdin.Read(next_bytes)
			if err != nil || n < 2 {
				continue
			}
			if next_bytes[0] == 91 {
				if next_bytes[1] == 65 { // Up
					if cur_history == -1 { // navigation started first time
						cur_history = len(history) - 1
						temp_history = current_buffer
					} else {
						cur_history = cur_history - 1
					}

					if cur_history < 0 {
						cur_history = 0
						continue
					}
					len_prev_buffer := len(current_buffer)
					current_buffer = []byte(history[cur_history].command)
					redrawBuffer(current_buffer, len_prev_buffer)
				} else if next_bytes[1] == 66 { // Down
					if cur_history == -1 {
						continue
					} else {
						cur_history = cur_history + 1
					}

					if cur_history >= len(history) {
						len_prev_buffer := len(current_buffer)
						current_buffer = temp_history
						redrawBuffer(current_buffer, len_prev_buffer)

						//reset if we go all the way back to current command
						cur_history = -1
						temp_history = nil
						continue
					}
					len_prev_buffer := len(current_buffer)
					current_buffer = []byte(history[cur_history].command)
					redrawBuffer(current_buffer, len_prev_buffer)
				} else {
					continue
				}
			} else {
				continue
			}

		} else {
			// echo the rest
			echoLetterAndAppendToBuffer(typed_character, &current_buffer)
		}
	}
}

func echoLetterAndAppendToBuffer(typed_character byte, buffer *[]byte) {
	fmt.Print(string(typed_character))
	*buffer = append(*buffer, typed_character)
}

func redrawBuffer(buffer []byte, len_prev_buffer int) {
	fmt.Printf(MOVE_CURSOR_X_LEFT, len_prev_buffer) // move cursor to the beginning

	// override previous input
	for range len_prev_buffer {
		fmt.Print(" ")
	}

	fmt.Printf(MOVE_CURSOR_X_LEFT, len_prev_buffer) // move cursor to the beginning again
	fmt.Print(string(buffer))
}

func calculateLastWord(buffer []byte) string {
	if len(buffer) == 0 {
		return ""
	}

	words := bytes.Split(buffer, []byte{byte(' ')})
	return string(words[len(words)-1])
}

func replaceLastWord(buffer, new_last_word []byte) []byte {
	last_word := calculateLastWord(buffer)

	initial_words := buffer[:len(buffer)-len(last_word)]

	return append(initial_words, new_last_word...)
}

// Give additional common bits of given completions on top of prefix.
func getCompletionDenominator(prefix string, completions []string) []byte {
	result := make([]byte, 0)
	for i := len(prefix); ; i++ {
		if len(completions[0]) <= i {
			break
		}

		var unequal = false
		next_char := completions[0][i]
		for _, completion := range completions {
			if len(completion) <= i || completion[i] != next_char {
				unequal = true
				break
			}
		}

		if unequal {
			break
		}

		result = append(result, next_char)
	}
	return result
}
