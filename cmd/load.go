/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
    "encoding/json"
    "context"
    "github.com/redis/go-redis/v9"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "github.com/spf13/cobra"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
    Args: cobra.ExactArgs(1),
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
        mode := args[0]

        inputReader := cmd.InOrStdin()
        decoder := json.NewDecoder(inputReader)

        if mode == "entity" {
            var target EntityBackup
            err := decoder.Decode(&target)
            if err != nil {
                fmt.Println("error:", err)
            }
            loadEntity(target)
        } else if mode == "stream" {
            var target StreamBackup
            err := decoder.Decode(&target)
            if err != nil {
                fmt.Println("error:", err)
            }
            loadStream(target)
        } else {
            fmt.Println("unknown mode")
        }
	},
}

var (
    load_dbname string
    load_dbhost string
    load_dbuser string
    load_dbpass string
    load_dbport string
    load_rdbaddr string
)

func init() {
	rootCmd.AddCommand(loadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
    loadCmd.Flags().StringVarP(&load_dbname, "dbname", "d", "concurrent", "Database name")
    loadCmd.Flags().StringVarP(&load_dbhost, "dbhost", "H", "localhost", "Database host")
    loadCmd.Flags().StringVarP(&load_dbuser, "dbuser", "u", "postgres", "Database user")
    loadCmd.Flags().StringVarP(&load_dbpass, "dbpassword", "p", "postgres", "Database password")
    loadCmd.Flags().StringVarP(&load_dbport, "dbport", "P", "5432", "Database port")
    loadCmd.Flags().StringVarP(&load_rdbaddr, "redisAddr", "r", "localhost:6379", "Redis address")
}

func loadEntity(backup EntityBackup) {

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        load_dbhost, load_dbuser, load_dbpass, load_dbname, load_dbport)
    redisAddr := load_rdbaddr

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }

    rdb := redis.NewClient(&redis.Options{
        Addr:     redisAddr,
        Password: "", // no password set
        DB:       0,  // use default DB
    })

    ctx := context.Background()

    // load entity
    db.Save(&backup.Entity)

    // load messages
    for _, message := range backup.Messages {
        db.Create(&message)
    }

    // load characters
    for _, character := range backup.Characters {
        db.Create(&character)
    }

    // load userkv
    for _, userkv := range backup.UserKV {
        rdb.Set(ctx, userkv.ID, userkv.Value, 0)
    }
}


func loadStream(backup StreamBackup) {

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        load_dbhost, load_dbuser, load_dbpass, load_dbname, load_dbport)
    redisAddr := load_rdbaddr

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }

    rdb := redis.NewClient(&redis.Options{
        Addr:     redisAddr,
        Password: "", // no password set
        DB:       0,  // use default DB
    })

    ctx := context.Background()

    // load stream
    db.Create(&backup.Stream)

    // load Elements
    for _, element := range backup.Elements {
        _, err := rdb.XAdd(ctx, &redis.XAddArgs{
            Stream: backup.Stream.ID,
            ID: element.ID,
            Values: element.Values,
        }).Result()
        if err != nil {
            fmt.Printf("fail to xadd: %v", err)
        }
    }

}

