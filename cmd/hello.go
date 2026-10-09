package cmd

import (
	"fmt"

	vocabtrie "github.com/jpabloVega/dokodemoYomou/dictionary/vocab_trie"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(tmpCmd)
}

var tmpCmd = &cobra.Command{
	Use:   "create",
	Short: "Get the urls",
	Long:  "Implement me",
	Run:   helloWorld,
}

func helloWorld(_ *cobra.Command, _ []string) {
	err := vocabtrie.CreateTrie()
	if err != nil {
		fmt.Println(err)
	}
}
