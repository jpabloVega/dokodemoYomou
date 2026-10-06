package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func helloWorld(_ *cobra.Command, _ []string) {
	fmt.Println("Hello world")
}
