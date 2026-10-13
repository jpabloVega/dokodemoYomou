package vocabtrie

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/gocarina/gocsv"
)

type Word struct {
	Level      string `csv:"level"`
	Kana       string `csv:"kana"`
	Kanji      string `csv:"kanji"`
	Definition string `csv:"definition"`
	Type       string `csv:"type"`
}

type StartEnd struct {
	Word       Word
	KanjiStart string
	KanjiEnd   string
	KanaStart  string
	KanaEnd    string
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

func conjugator(word Word) []Word {
	if word.Kanji == "" {
		return []Word{}
	}

	fullKanji := []rune(word.Kanji)
	lastCharPos := len(fullKanji) - 1
	kanjiStart := fullKanji[0:lastCharPos]
	kanjiEnd := fullKanji[lastCharPos]

	fullKana := []rune(word.Kana)
	lastCharPos = len(fullKana) - 1
	kanaStart := fullKana[0:lastCharPos]
	kanaEnd := fullKana[lastCharPos]

	allVerbConj := []Word{}
	resultConj := []Word{}
	allVerbConj = append(allVerbConj, word)

	conjugateData := StartEnd{
		Word:       word,
		KanjiStart: string(kanjiStart),
		KanjiEnd:   string(kanjiEnd),
		KanaStart:  string(kanaStart),
		KanaEnd:    string(kanaEnd),
	}

	switch word.Type {
	case "Verb Iku":
		resultConj = ConjugateIku(conjugateData)
	case "Verb Kuru":
		resultConj = ConjugateKuru(conjugateData)
	case "Verb Suru":
		return allVerbConj
	case "Verb":
		return allVerbConj
	case "Ichidan Verb":
		return allVerbConj
	case "I Adjective":
		resultConj = IAdjConjugator(conjugateData)
	case "Yoi Adjective":
		return allVerbConj
	default:
		return allVerbConj
	}
	return append(allVerbConj, resultConj...)

}

func ConjugateIku(conjData StartEnd) []Word {
	verb := conjData.Word
	var allVerbConj []Word
	verbEnds := []string{"きます", "かない", "きません", "こう", "きましょう",
		"け", "ってください", "くな", "かないでください", "った", "きました", "かなかった",
		"きませんでした", "って", "けば", "かなければ", "ったら", "かなかったら", "ける",
		"けます", "けない", "けません", "か"}
	verbEndDesc := []string{"\nPresent Polite",
		"\nPlain Negative",
		"\nPolite Negative",
		"\nVolitional Plain",
		"\nVolitional Polite",
		"\nImperative Plain",
		"\nImperative Polite",
		"\nImperative Plain Negative",
		"\nImperative Polite Negative",
		"\nPast Plain",
		"\nPast Polite",
		"\nPast Plain Negative",
		"\nPast Polite Negative",
		"\nTe-Form",
		"\nContional Plain",
		"\nConditional Negative",
		"\nConditional Tara Form",
		"\nConditional Tara Form Negative",
		"\nPotential Plain",
		"\nPotential Polite",
		"\nPotential Plain Negative",
		"\nPotential Polite Negative",
		"\nCausative Pasive"}

	for i, end := range verbEnds {
		allVerbConj = append(allVerbConj, Word{
			Level:      verb.Level,
			Kana:       conjData.KanaStart + end,
			Kanji:      conjData.KanjiStart + end,
			Definition: verb.Definition + verbEndDesc[i],
		})
	}
	return allVerbConj
}

func ConjugateKuru(conjData StartEnd) []Word {
	verb := conjData.Word
	var allVerbConj []Word
	verbEndsKanji := []string{"ます", "ない", "ません", "よう", "ましょう",
		"い", "てください", "るな", "ないでください", "た", "ました", "なかった",
		"ませんでした", "て", "れば", "なければ", "たら", "なかったら", "られる",
		"られます", "られない", "られません", "ら"}
	verbEndsKana := []string{"きます", "こない", "きません", "こよう", "きましょう",
		"こい", "きてください", "くるな", "こないでください", "きた", "きました", "こなかった",
		"きませんでした", "きて", "くれば", "こなければ", "きたら", "こなかったら", "こられる",
		"こられます", "こられない", "こられません", "こら"}
	verbEndDesc := []string{"\nPresent Polite",
		"\nPlain Negative",
		"\nPolite Negative",
		"\nVolitional Plain",
		"\nVolitional Polite",
		"\nImperative Plain",
		"\nImperative Polite",
		"\nImperative Plain Negative",
		"\nImperative Polite Negative",
		"\nPast Plain",
		"\nPast Polite",
		"\nPast Plain Negative",
		"\nPast Polite Negative",
		"\nTe-Form",
		"\nContional Plain",
		"\nConditional Negative",
		"\nConditional Tara Form",
		"\nConditional Tara Form Negative",
		"\nPotential Plain",
		"\nPotential Polite",
		"\nPotential Plain Negative",
		"\nPotential Polite Negative",
		"\nCausative Pasive"}

	kanjiMinusOne := []rune(conjData.KanjiStart)
	lastCharPos := len(kanjiMinusOne) - 1
	kanjiStart := kanjiMinusOne[0:lastCharPos]
	kanjiEnd := kanjiMinusOne[lastCharPos]

	kanaMinusOne := []rune(conjData.KanaStart)
	lastCharPos = len(kanaMinusOne) - 1
	kanaStart := kanaMinusOne[0:lastCharPos]

	if string(kanjiEnd) == "来" {
		for i, end := range verbEndsKanji {
			allVerbConj = append(allVerbConj, Word{
				Level:      verb.Level,
				Kana:       string(kanaStart) + verbEndsKana[i],
				Kanji:      conjData.KanjiStart + end,
				Definition: verb.Definition + verbEndDesc[i],
			})
		}
		return allVerbConj
	}

	for i, end := range verbEndsKana {
		allVerbConj = append(allVerbConj, Word{
			Level:      verb.Level,
			Kana:       string(kanaStart) + end,
			Kanji:      string(kanjiStart) + end,
			Definition: verb.Definition + verbEndDesc[i],
		})
	}
	return allVerbConj
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

func IAdjConjugator(conjData StartEnd) []Word {
	adj := conjData.Word
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

	for i, end := range adjEndings {
		allAdjConj = append(allAdjConj, Word{
			Level:      adj.Level,
			Kana:       conjData.KanaStart + end,
			Kanji:      conjData.KanjiStart + end,
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
