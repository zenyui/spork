package cmd

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "spork",
}

// Execute adds all child commands
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// load the env vars
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error loading .env file, %s", err.Error())
	}
}
