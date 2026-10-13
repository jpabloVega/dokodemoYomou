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

type StartEnd struct {
	Word       Word
	KanjiStart string
	KanjiEnd   string
	KanaStart  string
	KanaEnd    string
}

func CSVtoJson(directory string) ([]Word, error) {
	file, err := os.Open(directory)
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

func TestConjucator() {
	ikuTest := Word{
		Level:      "n5",
		Kanji:      "行く",
		Kana:       "行[い]く",
		Definition: "to go, to move (towards), to head (towards), to leave (for)",
		Type:       "Verb Iku",
	}
	kuruTest := Word{
		Level:      "n5",
		Kanji:      "来る",
		Kana:       "来[く]る",
		Definition: "to come (spatially or temporally), to approach, to arrive",
		Type:       "Verb Kuru",
	}
	iAdjTest := Word{
		Level:      "extra",
		Kanji:      "暇がない",
		Kana:       "暇[ひま]がない",
		Definition: "busy, having no (free) time",
		Type:       "I Adjective",
	}
	ichidanTest := Word{
		Level:      "extra",
		Kanji:      "報いる",
		Kana:       "報[むく]いる",
		Definition: "to reward, to recompense, to repay",
		Type:       "Ichidan Verb",
	}
	godanTest := Word{
		Level:      "extra",
		Kanji:      "雨が降る",
		Kana:       "雨[あめ]が 降[ふ]る",
		Definition: "to rain",
		Type:       "Verb",
	}
	yoiTest := Word{
		Level:      "extra",
		Kanji:      "心地良い",
		Kana:       "心地良[ここちよ]い",
		Definition: "comfortable, pleasant",
		Type:       "Yoi Adjective",
	}
	suruTest := Word{
		Level:      "extra",
		Kanji:      "擁する",
		Kana:       "擁[よう]する",
		Definition: "to embrace, to hug",
		Type:       "Verb Suru",
	}
	normalTest := Word{
		Level:      "n4",
		Kanji:      "工場",
		Kana:       "工場[こうじょう]",
		Definition: "factory, plant, mill, workshop",
		Type:       "Noun",
	}
	var words []Word
	words = append(words, ikuTest, kuruTest, iAdjTest, ichidanTest, godanTest, yoiTest, suruTest, normalTest)

	for _, word := range words {
		conjugations := conjugator(word)
		for _, conj := range conjugations {
			fmt.Println(conj.Level)
			fmt.Println(conj.Kanji)
			fmt.Println(conj.Kana)
			fmt.Println(conj.Definition)
		}
	}
}

func conjugator(word Word) []Word {
	if word.Kanji == "" {
		return []Word{}
	}

	conjugatedTypes := []string{"Verb Iku", "Verb Kuru", "Verb Suru", "Verb", "Ichidan Verb", "I Adjective", "Yoi Adjective"}

	if !slices.Contains(conjugatedTypes, word.Type) {
		return []Word{word}
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
		resultConj = ConjugateSuru(conjugateData)
	case "Verb":
		resultConj = ConjugateGodan(conjugateData)
	case "Ichidan Verb":
		resultConj = ConjugateIchidan(conjugateData)
	case "I Adjective":
		resultConj = IAdjConjugator(conjugateData)
	case "Yoi Adjective":
		resultConj = YoiConjugator(conjugateData)
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

func ConjugateSuru(conjData StartEnd) []Word {
	verb := conjData.Word
	var allVerbConj []Word
	verbEnds := []string{"します", "しない", "しません", "しよう", "しましょう",
		"しろ", "してくだい", "するな", "しないでください", "した", "しました", "しなかった",
		"しませんでした", "して", "すれば", "しなければ", "したら", "しなかったら", "できる",
		"できます", "できない", "できません", "され"}
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

	if len([]rune(conjData.Word.Kanji)) <= 2 {
		for i, end := range verbEnds {
			allVerbConj = append(allVerbConj, Word{
				Level:      verb.Level,
				Kana:       end,
				Kanji:      end,
				Definition: verb.Definition + verbEndDesc[i],
			})
		}
		return allVerbConj
	}

	kanjiMinusTwo := []rune(conjData.KanjiStart)
	lastCharPos := len(kanjiMinusTwo) - 1
	kanjiStart := kanjiMinusTwo[0:lastCharPos]

	kanaMinusTwo := []rune(conjData.KanaStart)
	lastCharPos = len(kanaMinusTwo) - 1
	kanaStart := kanaMinusTwo[0:lastCharPos]

	for i, end := range verbEnds {
		allVerbConj = append(allVerbConj, Word{
			Level:      verb.Level,
			Kana:       string(kanaStart) + end,
			Kanji:      string(kanjiStart) + end,
			Definition: verb.Definition + verbEndDesc[i],
		})
	}
	return allVerbConj
}

func ConjugateGodan(conjData StartEnd) []Word {
	verb := conjData.Word
	v := getEnds(conjData.KanjiEnd)
	var allVerbConj []Word
	verbEnds := []string{v.i + "ます", v.a + "ない", v.i + "ません", v.o, v.i + "ましょう",
		v.e, v.te + "ください", conjData.KanjiEnd + "な", v.a + "ないでください",
		v.ta, v.i + "ました", v.a + "なかった", v.i + "ませんでした", v.te, v.e + "ば",
		v.a + "なければ", v.ta + "ら", v.a + "なかったら", v.e + "る",
		v.e + "ます", v.e + "ない", v.e + "ません", v.a}
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

func ConjugateIchidan(conjData StartEnd) []Word {
	verb := conjData.Word
	var allVerbConj []Word
	verbEnds := []string{"ます", "ない", "ません", "よう", "ましょう",
		"よ", "てください", "るな", "ないでください",
		"た", "ました", "なかった", "ませんでした", "て", "れば",
		"なければ", "たら", "なかったら", "られる",
		"られます", "られない", "られません", "ら"}
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

type Godan struct {
	a  string
	e  string
	o  string
	i  string
	te string
	ta string
}

func getEnds(char string) Godan {
	switch char {
	case "う":
		return Godan{a: "わ", e: "え", o: "おう", i: "い", te: "って", ta: "った"}
	case "く":
		return Godan{a: "か", e: "け", o: "こう", i: "き", te: "いて", ta: "いた"}
	case "す":
		return Godan{a: "さ", e: "せ", o: "そう", i: "し", te: "して", ta: "した"}
	case "つ":
		return Godan{a: "た", e: "て", o: "とう", i: "ち", te: "って", ta: "った"}
	case "ぬ":
		return Godan{a: "な", e: "ね", o: "のう", i: "に", te: "んで", ta: "んだ"}
	case "ぶ":
		return Godan{a: "ば", e: "べ", o: "ぼう", i: "び", te: "んで", ta: "んだ"}
	case "む":
		return Godan{a: "ま", e: "め", o: "もう", i: "み", te: "んで", ta: "んだ"}
	case "る":
		return Godan{a: "ら", e: "れ", o: "ろう", i: "り", te: "って", ta: "った"}
	case "ぐ":
		return Godan{a: "が", e: "げ", o: "ごう", i: "ぎ", te: "いで", ta: "いた"}
	default:
		return Godan{}
	}

}

func IAdjConjugator(conjData StartEnd) []Word {
	adj := conjData.Word
	var allAdjConj []Word

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

func YoiConjugator(conjData StartEnd) []Word {
	adj := conjData.Word
	var allAdjConj []Word

	adjEndings := []string{"よくない", "よかった", "よくなかった", "よくて", "よく", "よければ"}
	adjEndDesc := []string{"\nPresent Negative",
		"\nPast Affirmative",
		"\nPast Negative",
		"\nTe-Form",
		"\nKu-Form",
		"\nConditional"}

	kanjiMinusTwo := []rune(conjData.KanjiStart)
	lastCharPos := len(kanjiMinusTwo) - 1
	kanjiStart := kanjiMinusTwo[0:lastCharPos]

	kanaMinusTwo := []rune(conjData.KanaStart)
	lastCharPos = len(kanaMinusTwo) - 1
	kanaStart := kanaMinusTwo[0:lastCharPos]

	for i, end := range adjEndings {
		allAdjConj = append(allAdjConj, Word{
			Level:      adj.Level,
			Kana:       string(kanaStart) + end,
			Kanji:      string(kanjiStart) + end,
			Definition: adj.Definition + adjEndDesc[i],
		})
	}
	return allAdjConj
}

func CreateTrie() error {
	words, err := CSVtoJson("dictionary/vocab_N5-N1/complete.csv")
	if err != nil {
		return err
	}

	t := NewTrie()

	//err = t.LoadFromFile("dictionary/vocab_trie/dictionary.json")
	//if err != nil {
	//return err
	//}

	for _, word := range words {
		conjugations := conjugator(word)
		for _, conj := range conjugations {
			t.Insert(conj)
		}
	}

	err = t.SaveToFile("dictionary/vocab_trie/dictionary.json")
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

func (t *Trie) Search(word string) (Word, bool) {
	current := t.Root
	for _, ch := range word {
		charStr := string(ch)
		if next, exists := current.Children[charStr]; exists {
			current = next
		} else {
			return Word{}, false
		}
	}
	return Word{Level: current.Level, Kana: current.Reading, Kanji: word, Definition: current.Definition}, true
}

func (t *Trie) SaveToFile(directory string) error {
	data, err := json.MarshalIndent(t.Root, "", " ")
	if err != nil {
		return fmt.Errorf("Failed to serialize json: %v", err)
	}
	return os.WriteFile(directory, data, 0644)
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
