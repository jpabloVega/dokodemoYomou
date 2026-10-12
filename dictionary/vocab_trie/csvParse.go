package vocabtrie

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/gocarina/gocsv"
)

type Word struct {
	Level      string `csv:"level"`
	Kana       string `csv:"kana"`
	Kanji      string `csv:"kanji"`
	Definition string `csv:"definition"`
	Type       string `csv:"type"`
}

func CSVtoJson() ([]Word, error) {
	file, err := os.Open("dictionary/vocab_N5-N1/verbs_n4.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var words []Word

	err = gocsv.UnmarshalFile(file, &words)
	if err != nil {
		panic(err)
	}

	if len(words) < 1 {
		panic(errors.New("words is empty"))
	}
	return words, nil
}

func TestVerbConjucation() error {
	file, err := os.Open("dictionary/vocab_N5-N1/verbs_n4.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var words []Word

	err = gocsv.UnmarshalFile(file, &words)
	if err != nil {
		panic(err)
	}

	if len(words) < 1 {
		panic(errors.New("words is empty"))
	}
	for _, verb := range words {
		if verb.Type == "verb" {
			conjVerb := conjugator(verb.Kanji)
			fmt.Println(conjVerb)
			continue
		}
		cleanAdj := AdjConjugator(verb)
		fmt.Println(cleanAdj)
	}
	return nil
}

func conjugator(verb string) string {
	if verb == "" {
		return ""
	}
	fullWord := []rune(verb)
	if len(fullWord) < 2 {
		return verb
	}
	terminations := []string{"う", "く", "ぐ", "す", "つ", "ぬ", "ぶ", "む", "る"}
	lastCharPos := len(fullWord) - 1
	start := fullWord[0:lastCharPos]
	last := fullWord[lastCharPos]
	if slices.Contains(terminations, string(last)) {
		return string(start)
	}
	return verb
}

func TestAdjConjugator() {
	testWord := Word{
		Level:      "n1",
		Kana:       "かがやかしい",
		Kanji:      "輝かしい",
		Definition: "brilliant, glorious, glittering, bright, splendid (e.g. achievement, success), promising (e.g. future)",
		Type:       "I Adjective",
	}
	results := AdjConjugator(testWord)
	for _, result := range results {
		fmt.Println(result.Level)
		fmt.Println(result.Kanji)
		fmt.Println(result.Kana)
		fmt.Println(result.Definition)
		fmt.Println()
	}
}

func AdjConjugator(adj Word) []Word {
	var allAdjConj []Word
	if adj.Kanji == "" {
		return []Word{}
	}

	adjEndings := []string{"くない", "かった", "くなかった", "くて", "く", "ければ"}
	adjEndDesc := []string{"\nPresent Negative",
		"\nPast Affirmative",
		"\nPast Negative",
		"\nTe-Form",
		"\nKu-Form",
		"\nConditional"}

	kanji := []rune(adj.Kanji)
	lastKanjiPos := len(kanji) - 1
	startKanji := kanji[0:lastKanjiPos]

	if string(kanji[len(kanji)-1]) != "い" {
		return allAdjConj
	}

	kana := []rune(adj.Kana)
	lastKanaPos := len(kana) - 1
	startKana := kana[0:lastKanaPos]

	allAdjConj = append(allAdjConj, adj)

	for i, end := range adjEndings {
		allAdjConj = append(allAdjConj, Word{
			Level:      adj.Level,
			Kana:       string(startKana) + end,
			Kanji:      string(startKanji) + end,
			Definition: adj.Definition + adjEndDesc[i],
		})
	}
	return allAdjConj
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
		if word.Type == "verb" {
			t.Insert(word)
		} else {
			t.Insert(word)
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
	Reading    string               `json:"reading"`
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

func (t *Trie) Insert(word Word) {
	current := t.Root
	for _, ch := range word.Kanji {
		charStr := string(ch)
		if _, exists := current.Children[charStr]; !exists {
			current.Children[charStr] = &TrieNode{Children: make(map[string]*TrieNode)}
		}
		current = current.Children[charStr]
	}
	current.Level = word.Level
	current.Reading = word.Kana
	current.Definition = word.Definition
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
