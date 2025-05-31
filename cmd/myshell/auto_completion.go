package main

import (
	"os"
	"strings"
)

type PrefixTreeNode struct {
	is_end_of_word bool
	children       map[rune]*PrefixTreeNode
}

// Build prefix tree with list of commands. Returns root node.
func buildAutocompletionDB(command_list []string) *PrefixTreeNode {
	root_node := PrefixTreeNode{
		false,
		make(map[rune]*PrefixTreeNode, 0),
	}

	for _, command := range command_list {
		addToPrefixTree(command, &root_node)
	}

	// add path commands
	path_commands := getCommandsFromPath()
	for _, command := range path_commands {
		addToPrefixTree(command, &root_node)
	}

	return &root_node
}

// Add command to given prefix tree.
func addToPrefixTree(command string, root *PrefixTreeNode) {
	var trav = root

	for _, cur_rune := range command {
		val, ok := trav.children[cur_rune]
		if ok {
			trav = val
		} else {
			trav.children[cur_rune] = &PrefixTreeNode{
				false,
				make(map[rune]*PrefixTreeNode, 0),
			}
			trav = trav.children[cur_rune]
		}
	}
	trav.is_end_of_word = true
}

// Search all possible words in given prefix tree, given a prefix.
func searchPrefixTree(prefix string, root *PrefixTreeNode) []string {
	trav := root
	for _, char := range prefix {
		val, ok := trav.children[char]
		if ok {
			trav = val
		} else {
			return []string{}
		}
	}

	result := make([]string, 0)
	getAllChildrenAsList(trav, prefix, &result)
	return result
}

// Recursive call that traverses node and prints all children.
func getAllChildrenAsList(node *PrefixTreeNode, current_string string, result *[]string) {
	if node.is_end_of_word {
		*result = append(*result, current_string)
	}

	if len(node.children) == 0 {
		return
	}

	for key, value := range node.children {
		getAllChildrenAsList(value, current_string+string(key), result)
	}
}

func getCommandsFromPath() []string {
	results := make([]string, 0)

	path := os.Getenv("PATH")
	dirs := strings.Split(path, string(os.PathListSeparator))

	for _, dir := range dirs {
		getCommandsFromDir(dir, &results)
	}
	return results
}

func getCommandsFromDir(dir string, results *[]string) {
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		entry.Type()
		if entry.Type()&0111 != 0 {
			*results = append(*results, entry.Name())
		}
	}
}
