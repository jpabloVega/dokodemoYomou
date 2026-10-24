package api

import (
	"strings"

	vocabtrie "github.com/jpabloVega/dokodemoYomou/dictionary/vocab_trie"
)

func replaceKanji(chapter Chapter) (Chapter, error) {
	for _, line := range chapter.Contents {
		if line == "" {
			continue
		}

		t := vocabtrie.NewTrie()
		err = t.LoadFromFile("dictionary/vocab_trie/dictionary.json")
		if err != nil {
			return err
		}

		runes := []rune(line)
		var resultStr strings.Builder
		start := 0
		for start < len(runes) {
			found := false
			var kanaWord []rune
			for j := len(runes); j >= 0; j-- {
				currWord := string(runes[start:j])
				foundWord, found := t.Search(currWord)
				if found {
					kanaWord = []rune(foundWord.Kana)
					break
				}
			}
			if found {
				resultStr.WriteRune(kanaWord[])
			}
		}
	}
}
