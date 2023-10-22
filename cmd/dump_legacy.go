/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/totegamma/concurrent/x/core"
	"gorm.io/gorm"
)

// dumpCmd represents the dump command
var legacyCmd = &cobra.Command{
	Use:   "dump_legacy",
	Args:  cobra.ExactArgs(2),
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		mode := args[0]
		target := args[1]

		if mode == "entity" {
			dumpEntity(target)
		} else if mode == "stream" {
			dumpStream(target)
		} else {
			fmt.Println("unknown mode")
		}
	},
}

type AllEntityBackup struct {
	Entities map[string]EntityBackup `json:"entities"`
}

type EntityBackup struct {
	CCID       string           `json:"ccid"`
	Entity     core.Entity      `json:"entity"`
	Messages   []core.Message   `json:"messages"`
	Characters []core.Character `json:"characters"`
	UserKV     []UserKV         `json:"userkv"`
}

type UserKV struct {
	ID    string `json:"key"`
	Value string `json:"value"`
}

func init() {
	rootCmd.AddCommand(legacyCmd)
}

func dumpSingleEntity(db *gorm.DB, rdb *redis.Client, targetID string) (EntityBackup, error) {
	backup := EntityBackup{}
	backup.CCID = targetID

	var entity core.Entity
	db.First(&entity, "id = ?", targetID)
	backup.Entity = entity

	// get all messages
	var messages []core.Message
	db.Preload("Associations").Find(&messages, "author = ?", targetID)
	backup.Messages = messages

	// get all characters
	var characters []core.Character
	db.Preload("Associations").Find(&characters, "author = ?", targetID)
	backup.Characters = characters

	// get all userkv
	ctx := context.Background()
	pattern := "userkv:" + targetID + ":*"
	var cursor uint64
	var keys []string
	var userkvs []UserKV
	for {
		var err error
		keys, cursor, err = rdb.Scan(ctx, cursor, pattern, 10).Result()
		if err != nil {
			panic(err)
		}

		for _, key := range keys {
			val, err := rdb.Get(ctx, key).Result()
			if err != nil {
				panic(err)
			}
			userkvs = append(userkvs, UserKV{key, val})
		}

		if cursor == 0 {
			break
		}
	}
	backup.UserKV = userkvs

	return backup, nil

}

func dumpEntity(targetID string) {

	db := openDB()
	rdb := openRDB()

	if targetID == "all" {
		// get all entities
		entities := []core.Entity{}
		err := db.Find(&entities).Error
		if err != nil {
			panic(err)
		}
		backup := AllEntityBackup{}
		backup.Entities = make(map[string]EntityBackup)
		for i, entity := range entities {
			fmt.Fprintf(os.Stderr, "dumping %s (%d/%d)\n", entity.ID, i, len(entities))
			backup.Entities[entity.ID], err = dumpSingleEntity(db, rdb, entity.ID)
			if err != nil {
				panic(err)
			}
		}

		fmt.Fprintf(os.Stderr, "Serializing...\n")
		b, err := json.MarshalIndent(backup, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))
	} else {
		backup, err := dumpSingleEntity(db, rdb, targetID)
		if err != nil {
			panic(err)
		}
		// print backup
		b, err := json.MarshalIndent(backup, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))
	}
}

type StreamBackup struct {
	Stream   core.Stream      `json:"stream"`
	Items[]  core.StreamItem   `json:"items"`
}

type AllStreamBackup struct {
	Streams map[string]StreamBackup `json:"streams"`
}

func dumpSingleStream(db *gorm.DB, rdb *redis.Client, targetStream string) (StreamBackup, error) {
	backup := StreamBackup{}

	stream := core.Stream{}
	db.First(&stream, "id = ?", targetStream)
	backup.Stream = stream

	ctx := context.Background()
	cmd := rdb.XRange(ctx, targetStream, "-", "+")
	results, err := cmd.Result()
	if err != nil {
		panic(err)
	}

	for _, item := range results {
		streamItem := core.StreamItem{}
		streamItem.StreamID = targetStream

		unixTime := item.ID // 00000-00 notation
		split := strings.Split(unixTime, "-")

		millis, err := strconv.ParseInt(split[0], 10, 64)
		if err != nil {
			continue
		}
		decimal := millis / 1000
		fraction := (millis % 1000) * 1000000
		streamItem.CDate = time.Unix(decimal, fraction)

		id, ok := item.Values["id"].(string)
		if ok {
			streamItem.ObjectID = id
		}

		typ, ok := item.Values["type"].(string)
		if ok {
			streamItem.Type = typ
		}

		author, ok := item.Values["author"].(string)
		if ok {
			streamItem.Author = author
		}

		owner, ok := item.Values["owner"].(string)
		if ok {
			streamItem.Owner = owner
		} else {
			streamItem.Owner = author
		}

		backup.Items = append(backup.Items, streamItem)
	}

	return backup, nil
}

func dumpStream(targetStream string) {

	db := openDB()
	rdb := openRDB()

	if targetStream == "all" {
		streams := []core.Stream{}
		err := db.Find(&streams).Error
		if err != nil {
			panic(err)
		}
		backup := AllStreamBackup{}
		backup.Streams = make(map[string]StreamBackup)
		for i, stream := range streams {
			// print progress
			fmt.Fprintf(os.Stderr, "Dumping stream %s (%d/%d)\n", stream.ID, i, len(streams))

			backup.Streams[stream.ID], err = dumpSingleStream(db, rdb, stream.ID)
			if err != nil {
				panic(err)
			}
		}

		fmt.Fprintf(os.Stderr, "Serializing...\n")

		b, err := json.MarshalIndent(backup, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))
	} else {
		backup, err := dumpSingleStream(db, rdb, targetStream)
		if err != nil {
			panic(err)
		}
		b, err := json.MarshalIndent(backup, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))
	}
}
