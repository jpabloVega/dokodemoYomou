package cmd

import (
	"time"

	"github.com/jpabloVega/dokodemoYomou/api"
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
	c := api.NewClient(5 * time.Second)
	_, err := api.TestGetHTML(args[0], &c)
	return err
}
