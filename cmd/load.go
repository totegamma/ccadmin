/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
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
		} else if mode == "entities" {
			var target AllEntityBackup
			err := decoder.Decode(&target)
			if err != nil {
				fmt.Println("error:", err)
			}
			loadEntities(target)
		} else if mode == "stream" {
			var target StreamBackup
			err := decoder.Decode(&target)
			if err != nil {
				fmt.Println("error:", err)
			}
			loadStream(target)
		} else if mode == "streams" {
			var target AllStreamBackup
			err := decoder.Decode(&target)
			if err != nil {
				fmt.Println("error:", err)
			}
			loadStreams(target)
		} else {
			fmt.Println("unknown mode")
		}
	},
}

func init() {
	rootCmd.AddCommand(loadCmd)
}

func loadEntity(backup EntityBackup) {

	db := openDB()
	rdb := openRDB()
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

func loadEntities(backup AllEntityBackup) {

	db := openDB()
	rdb := openRDB()
	ctx := context.Background()

	for _, entity := range backup.Entities {
		// load entity
		db.Save(&entity.Entity)

		// load messages
		for _, message := range entity.Messages {
			db.Create(&message)
		}

		// load characters
		for _, character := range entity.Characters {
			db.Create(&character)
		}

		// load userkv
		for _, userkv := range entity.UserKV {
			rdb.Set(ctx, userkv.ID, userkv.Value, 0)
		}
	}
}

func loadStream(backup StreamBackup) {

	db := openDB()

	// load stream
	db.Create(&backup.Stream)

	// load Items
	for _, item := range backup.Items {
		db.Create(&item)
	}
}

func loadStreams(backup AllStreamBackup) {

	db := openDB()

	for _, stream := range backup.Streams {
		// load stream
		db.Create(&stream.Stream)

		// load Items
		for _, item := range stream.Items {
			db.Create(&item)
		}
	}
}
