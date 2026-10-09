package cmd

import (
	"fmt"
	"time"

	"github.com/jpabloVega/dokodemoYomou/api"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the urls",
	Long:  "Implement me",
	Run:   GetContents,
}

func GetContents(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("No arguments given")
		return
	}
	c := api.NewClient(5 * time.Second)
	for _, address := range args {
		res, err := c.WebToPDF(address)
		if err != nil {
			fmt.Printf("Error getting addresses: %v", err)
			return
		}
		fmt.Println(res)
	}

}
