package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dodeyo",
	Short: "The easiest way to import you favorite novel from syosetsu.",
	Long: `Scan a novel to know the vocabulary level
	Download it and save it to your e-Reader
	Or email it to your kindle`,
	RunE: GetContents,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
