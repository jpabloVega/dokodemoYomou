package cmd

import (
	vocabtrie "github.com/jpabloVega/dokodemoYomou/dictionary/vocab_trie"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Get the urls",
	Long:  "Implement me",
	Run:   helloTest,
}

func helloTest(_ *cobra.Command, _ []string) {
	vocabtrie.TestAdjConjugator()
}
