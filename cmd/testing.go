package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Get the urls",
	Long:  "Implement me",
	RunE:  helloTest,
}

func helloTest(_ *cobra.Command, args []string) error {
	fmt.Println("not implemented")
	return nil
}
