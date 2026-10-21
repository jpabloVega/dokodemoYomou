package cmd

import (
	"fmt"

	"github.com/jpabloVega/dokodemoYomou/secure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	setEmail           bool
	showConfig         bool
	newDestinationPath string
	kindleEmail        string
)

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().BoolVar(&setEmail, "email", false, "Create or overwrite email")
	configCmd.Flags().BoolVar(&showConfig, "show", false, "Show current config")
	configCmd.Flags().StringVar(&newDestinationPath, "defaultPath", "", "Default path to save content")
	configCmd.Flags().StringVar(&kindleEmail, "kindleEmail", "", "Set your kindle address")

}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set Default config values",
	Long:  "Set Default behaviour for the app",
	Args:  cobra.NoArgs,
	Run:   setConfig,
}

func setConfig(_ *cobra.Command, _ []string) {
	if newDestinationPath != "" {
		viper.Set("defaultPath", newDestinationPath)
		err := viper.WriteConfig()
		if err != nil {
			fmt.Println(err)
		}
	}
	if kindleEmail != "" {
		viper.Set("kindlemail", kindleEmail)
		err := viper.WriteConfig()
		if err != nil {
			fmt.Println(err)
		}
	}
	if setEmail {
		_, _, err := secure.NewEmailPassword()
		if err != nil {
			fmt.Println(err)
		}
	}

	if showConfig {
		fmt.Println(viper.Get("author"))
		fmt.Println(viper.Get("userMail"))
		fmt.Println(viper.Get("defaultPath"))
	}
}
