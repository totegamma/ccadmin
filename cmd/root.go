/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var (
	dbname  string
	dbhost  string
	dbuser  string
	dbpass  string
	dbport  string
	rdbaddr string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ccadmin",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&dbname, "dbname", "d", "concurrent", "Database name")
	rootCmd.PersistentFlags().StringVarP(&dbhost, "dbhost", "H", "localhost", "Database host")
	rootCmd.PersistentFlags().StringVarP(&dbuser, "dbuser", "u", "postgres", "Database user")
	rootCmd.PersistentFlags().StringVarP(&dbpass, "dbpassword", "p", "postgres", "Database password")
	rootCmd.PersistentFlags().StringVarP(&dbport, "dbport", "P", "5432", "Database port")
	rootCmd.PersistentFlags().StringVarP(&rdbaddr, "rdbaddr", "r", "localhost:6379", "Redis address")
}
