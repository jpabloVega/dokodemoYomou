package cmd

import (
	"fmt"
	"os"

	"github.com/jpabloVega/dokodemoYomou/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	userLicense string

	startFrom     int
	chapterAmount int
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

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().StringP("author", "a", "YOUR NAME", "author name for copyright attribution")
	rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "name of license for the project")
	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
	viper.SetDefault("license", "apache")

	//rootCmd.AddCommand(addCmd)
	//rootCmd.AddCommand(initCmd)
}

func initConfig() {
	if cfgFile != "" {

	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Config load successfully")
		cobra.CheckErr(err)
	}

	err = database.StartDB()
	if err != nil {
		fmt.Println(err)
	}
}
