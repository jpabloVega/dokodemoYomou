package cmd

import (
	"fmt"
	"time"

	"github.com/jpabloVega/dokodemoYomou/api"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(scanCmd)
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
	err := c.ScanBook(args[0])
	if err != nil {
		fmt.Printf("Error scanning: %v", err)
	}
}
