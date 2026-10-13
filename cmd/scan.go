package cmd

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/jpabloVega/dokodemoYomou/api"
	vocabtrie "github.com/jpabloVega/dokodemoYomou/dictionary/vocab_trie"
	"github.com/spf13/cobra"
)

var verbose bool

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show details")

}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Get the urls",
	Long:  "Implement me",
	Run:   ScanBook,
}

func ScanBook(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("No arguments given")
		return
	}
	c := api.NewClient(5 * time.Second)
	for _, arg := range args {
		err := ScanBookCommand(arg, &c)
		if err != nil {
			fmt.Printf("Error scanning: %v", err)
		}
	}
}

func ScanBookCommand(address string, c *api.Client) error {
	urlData, err := api.GetSiteInfo(address)
	if err != nil {
		return err
	}
	chapterData, err := c.GetSingleAddress(urlData)
	if err != nil {
		return err
	}
	t := vocabtrie.NewTrie()
	err = t.LoadFromFile("dictionary/vocab_trie/dictionary.json")
	if err != nil {
		return err
	}
	nCount := make(map[string][]string)
	for _, line := range chapterData.Contents {
		splitLine := []rune(line)
		start := 0
		end := len(splitLine)
		for true {
			currWord := string(splitLine[start:end])
			foundWord, found := t.Search(currWord)
			if found {
				nCount[foundWord.Level] = append(nCount[foundWord.Level], foundWord.Kanji)
				start += utf8.RuneCountInString(currWord)
				end = len(splitLine)
				continue
			}
			end -= 1
			if end == start {
				start += 1
				end = len(splitLine)
			}
			if start >= len(splitLine)-1 {
				break
			}
		}
	}

	fmt.Println("Words from every level")
	for key, value := range nCount {
		if key == "" {
			continue
		}
		fmt.Printf("%s : %d\n", key, len(value))
		if verbose {
			for i, word := range value {
				fmt.Print(word + "  ")
				if i != 0 && i%11 == 0 || i == len(value)-1 {
					fmt.Println()
				}
			}
		}
	}
	return nil
}
