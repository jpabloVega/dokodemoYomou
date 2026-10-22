package cmd

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/jpabloVega/dokodemoYomou/api"
	vocabtrie "github.com/jpabloVega/dokodemoYomou/dictionary/vocab_trie"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	verbose bool
)

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show details")
	scanCmd.Flags().IntVarP(&chapterAmount, "num", "", 1, "Amount of chapters to scan")

}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Get the urls",
	Long:  "Implement me",
	RunE:  ScanBook,
}

func ScanBook(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("No argument given")
	}
	c := api.NewClient(5 * time.Second)
	for _, arg := range args {
		err := ScanBookCommand(arg, &c)
		if err != nil {
			return err
		}
	}
	return nil
}

func ScanBookCommand(address string, c *api.Client) error {
	urlData, err := api.GetSiteInfo(address)
	if err != nil {
		return err
	}
	chaptersChan := make(chan api.Chapter, chapterAmount)
	var wg sync.WaitGroup

	c.GetFromRange(urlData, chapterAmount, chaptersChan, &wg)

	go func() {
		wg.Wait()
		close(chaptersChan)
	}()

	chaptersData := make([]api.Chapter, chapterAmount)
	for res := range chaptersChan {
		chaptersData[res.Index] = res
		if res.Err != nil {
			fmt.Println(res.Err)
		}
	}

	t := vocabtrie.NewTrie()
	err = t.LoadFromFile("dictionary/vocab_trie/dictionary.json")
	if err != nil {
		return err
	}
	nCount := make(map[string][]string)

	for _, chapter := range chaptersData {
		chapterLen := len(chapter.Contents)
		chapterProgress := progressbar.Default(int64(chapterLen), "Scanning chapter...")
		for _, line := range chapter.Contents {
			chapterProgress.Add(1)
			if line == "" {
				continue
			}
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
	}

	if verbose {
		fmt.Printf("N1: %v\n", nCount["n1"])
		fmt.Println("")
		fmt.Printf("N2: %v\n", nCount["n2"])
		fmt.Println("")
		fmt.Printf("N3: %v\n", nCount["n3"])
		fmt.Println("")
		fmt.Printf("N4: %v\n", nCount["n4"])
		fmt.Println("")
		fmt.Printf("N5: %v\n", nCount["n5"])
		fmt.Println("")
	}
	fmt.Println("Words from every level")
	fmt.Printf("N1: %d\n", len(nCount["n1"]))
	fmt.Printf("N2: %d\n", len(nCount["n2"]))
	fmt.Printf("N3: %d\n", len(nCount["n3"]))
	fmt.Printf("N4: %d\n", len(nCount["n4"]))
	fmt.Printf("N5: %d\n", len(nCount["n5"]))
	return nil
}
