package vocabtrie

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/gocarina/gocsv"
)

const (
	CURR = "n1"
)

type Word struct {
	SequenceNumber int    `csv:"jmdict_seq"`
	Kana           string `csv:"kana"`
	Kanji          string `csv:"kanji"`
	Definition     string `csv:"waller_definition"`
}

type Verb struct {
	Kanji   string `csv:"kanji"`
	Kana    string `csv:"kana"`
	Meaning string `csv:"meaning"`
}

func CSVtoJson() ([]Verb, error) {
	file, err := os.Open("dictionary/vocab_N5-N1/verbs_n5.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var words []Verb

	err = gocsv.UnmarshalFile(file, &words)
	if err != nil {
		panic(err)
	}

	if len(words) < 10 {
		panic(errors.New("words is empty"))
	}
	return words, nil
}

func TestVerbConjucation(verbs []Verb) error {
	for _, verb := range verbs {

	}
}

func conjugator(verb string) ([]string, error) {
	irregulars := []string{"要る", "入る", "走る", "帰る", "切る", "知る", "知る", "喋る", "蹴る", "減る", "焦る", "握る", "散る", "弄る", "滑る"}
	if verb == "来る" || verb == "くる" {
		return []string{"来ます", "来ない", "来ません", "来た", "来ました", "来なかった", "来て", "来よう", "来ましょう", "来れば", "来い"}, nil
	}
	fullWord := []rune(verb)
	lastChar := len(fullWord) - 1
	start := fullWord[0:lastChar]
	last := fullWord[lastChar]
	present_end := []string{"ます", "ない", "ませんあ"}
	ru_conj := []string{""}
	if slices.Contains(irregulars, verb) {

	}
}

func CreateTrie() error {
	words, err := CSVtoJson()
	if err != nil {
		return err
	}

	t := NewTrie()

	err = t.LoadFromFile("dictionary/vocab_trie/dictionary.json")
	if err != nil {
		return err
	}

	for _, word := range words {
		if word.Kanji == "" {
			t.Insert(word.Kana)
		} else {
			t.Insert(word.Kanji)
		}
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
	Children   map[string]*TrieNode `json:"children"`
	IsEnd      bool                 `json:"is_end"`
	Level      string               `json:"level"`
	Definition string               `json:"definition"`
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
	current.Level = CURR
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
	return os.WriteFile("dictionary/vocab_trie/dictionary.json", data, 0644)
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
