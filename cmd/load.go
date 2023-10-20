/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
	Args:  cobra.ExactArgs(1),
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

func init() {
	rootCmd.AddCommand(loadCmd)
}

func loadEntity(backup EntityBackup) {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbhost, dbuser, dbpass, dbname, dbport)
	redisAddr := rdbaddr

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
		dbhost, dbuser, dbpass, dbname, dbport)
	redisAddr := rdbaddr

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
			ID:     element.ID,
			Values: element.Values,
		}).Result()
		if err != nil {
			fmt.Printf("fail to xadd: %v", err)
		}
	}

}
