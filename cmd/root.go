/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbname  string
	dbhost  string
	dbuser  string
	dbpass  string
	dbport  string
	rdbaddr string
	mcaddr  string
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
	rootCmd.PersistentFlags().StringVarP(&mcaddr, "mcaddr", "m", "localhost:11211", "Memcached address")
}

func openDB() *gorm.DB {

	logger := logger.New(
		log.New(os.Stderr, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:        time.Second,
			LogLevel:             logger.Silent,
			Colorful:             true,
			ParameterizedQueries: true,
		},
	)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbhost, dbuser, dbpass, dbname, dbport)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		panic(err)
	}

	return db
}

func openRDB() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     rdbaddr,
		Password: "",
		DB:       0,
	})

	return rdb
}
