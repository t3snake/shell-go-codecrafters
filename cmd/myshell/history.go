package main

type HistoryEntry struct {
	command string
	id      int
}

func initializeHistory() []HistoryEntry {
	return make([]HistoryEntry, 0)
}

func addToHistory(command string, history *[]HistoryEntry) {
	id := len(*history)

	entry := HistoryEntry{
		command,
		id,
	}

	*history = append(*history, entry)
}
