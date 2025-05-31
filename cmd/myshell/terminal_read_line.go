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
const MOVE_CURSOR_TO_BEG = "\033[0G"
const MOVE_CURSOR_1_LEFT = "\033[1D"

func terminalReadLine(auto_completion_db *PrefixTreeNode) (string, error) {
	// change terminal to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	// change terminal back to cooked mode
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	var current_buffer []byte
	input_char := make([]byte, 1)

	for {
		n, err := os.Stdin.Read(input_char)
		if err != nil || n == 0 {
			return "", err
		}

		typed_character := input_char[0]

		if typed_character == '\t' {
			// tab handling (\t or 9)
			prefix := calculateLastWord(current_buffer)
			results := searchPrefixTree(prefix, auto_completion_db)

			if len(results) == 0 {
				continue
			}

			if len(results) == 1 {
				len_prev_buffer := len(current_buffer)
				current_buffer = replaceLastWord(current_buffer, []byte(prefix))
				redrawBuffer(current_buffer, len_prev_buffer)
				continue
			}

			// TODO print all suggestions in new line and move cursor back to where we are typing

		} else if typed_character == '\b' || typed_character == 127 {
			// backspace handling (\b or 8) or terminal might give (del or 127)

			// TODO delete key handling, print(" \b")
			// space overrides key after cursor and \b moves back cursor to original position

			if len(current_buffer) > 0 {
				current_buffer = current_buffer[:len(current_buffer)-1]
			}
			// \b moves cursor backwards, space overrides and move cursor back again
			fmt.Print("\b \b")
		} else if typed_character == 3 {
			// ctrl+c or sigint handling (3)
			fmt.Print("\r\n")
			return "", fmt.Errorf("SIGINT")
		} else if typed_character == '\n' || typed_character == '\r' {
			// return on line feed (LF) (\n or 10) or carriage return (CR) (\r or 13)
			fmt.Print("\r\n")
			return string(current_buffer), nil
		} else {
			// echo the rest
			echoLetterAndAppenddToBuffer(typed_character, &current_buffer)
		}
	}
}

func echoLetterAndAppenddToBuffer(typed_character byte, buffer *[]byte) {
	fmt.Print(string(typed_character))
	*buffer = append(*buffer, typed_character)
}

func redrawBuffer(buffer []byte, len_prev_buffer int) {
	fmt.Print("\r") // move cursor to the beginning

	// override previous input
	for range len_prev_buffer {
		fmt.Print(" ")
	}

	fmt.Print("\r") // move cursor to the beginning again
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
