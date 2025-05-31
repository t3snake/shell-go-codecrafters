package main

type PrefixTreeNode struct {
	is_end_of_word bool
	children       map[rune]PrefixTreeNode
}

func addToPrefixTree(command string, root *PrefixTreeNode) {
	var trav = root

	runes := []rune(command)

	for idx, cur_rune := range runes {
		val, ok := trav.children[cur_rune]
		if ok {
			trav = &val
		} else {
			(*trav).children[cur_rune] = PrefixTreeNode{
				false,
				make(map[rune]PrefixTreeNode, 0),
			}
			child := trav.children[cur_rune]
			trav = &child
		}

		if idx == len(runes)-1 {
			trav.is_end_of_word = true
		}
	}
}

func getAllChildrenAsList(node PrefixTreeNode, current_string string, result *[]string) {
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

func searchPrefixTree(prefix string, root PrefixTreeNode) []string {
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

func buildAutocompletionDB(allowed_prompts []string) PrefixTreeNode {
	root := PrefixTreeNode{
		false,
		make(map[rune]PrefixTreeNode, 0),
	}

	for _, prompt := range allowed_prompts {
		addToPrefixTree(prompt, &root)
	}

	return root
}
