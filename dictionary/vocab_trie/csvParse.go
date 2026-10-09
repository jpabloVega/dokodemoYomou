package vocabtrie

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/gocarina/gocsv"
)

type Word struct {
	SequenceNumber int    `csv:"jmdict_seq"`
	Kana           string `csv:"kana"`
	Kanji          string `csv:"kanji"`
	Definition     string `csv:"waller_definition"`
}

func CSVtoJson() ([]Word, error) {
	file, err := os.Open("dictionary/vocab_N5-N1/n1.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var words []Word

	err = gocsv.UnmarshalFile(file, &words)
	if err != nil {
		panic(err)
	}

	if len(words) < 10 {
		panic(errors.New("words is empty"))
	}
	return words, nil
}

func CreateTrie() error {
	words, err := CSVtoJson()
	if err != nil {
		return err
	}

	t := NewTrie()
	err = t.LoadFromFile("dictionary/createjsontest/n5.json")
	if err != nil {
		return err
	}

	for _, word := range words {
		t.Insert(word.Kana)
		t.Insert(word.Kanji)
	}

	err = t.SaveToFile()
	if err != nil {
		return err
	}
	fmt.Println("trie created and saved")
	return nil
}

// trie logic

type TrieNode struct {
	Children map[string]*TrieNode `json:"children"`
	IsEnd    bool                 `json:"is_end"`
	Level    string               `level:"level"`
}

type Trie struct {
	Root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{
		Root: &TrieNode{Children: make(map[string]*TrieNode)},
	}
}

func (t *Trie) Insert(word string) {
	current := t.Root
	for _, ch := range word {
		charStr := string(ch)
		if _, exists := current.Children[charStr]; !exists {
			current.Children[charStr] = &TrieNode{Children: make(map[string]*TrieNode)}
		}
		current = current.Children[charStr]
	}
	current.Level = "n1"
	current.IsEnd = true
}

func (t *Trie) Search(word string) (string, bool) {
	current := t.Root
	for _, ch := range word {
		charStr := string(ch)
		if next, exists := current.Children[charStr]; exists {
			current = next
		} else {
			return "", false
		}
	}
	return current.Level, current.IsEnd
}

func (t *Trie) SaveToFile() error {
	data, err := json.MarshalIndent(t.Root, "", " ")
	if err != nil {
		return fmt.Errorf("Failed to serialize json: %v", err)
	}
	return os.WriteFile("dictionary/createjsontest/n5.json", data, 0644)
}

func (t *Trie) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var root TrieNode
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("failed to deserialize: %w", err)
	}

	t.Root = &root
	return nil
}
