/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
    "encoding/json"
    "context"
    "github.com/totegamma/concurrent/x/core"
    "github.com/redis/go-redis/v9"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
	"github.com/spf13/cobra"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
    Args: cobra.ExactArgs(2),
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

type EntityBackup struct {
    CCID string `json:"ccid"`
    Entity core.Entity `json:"entity"`
    Messages []core.Message `json:"messages"`
    Characters []core.Character `json:"characters"`
    UserKV []UserKV `json:"userkv"`
}

type UserKV struct {
    ID string `json:"key"`
    Value string `json:"value"`
}

var (
    dump_dbname string
    dump_dbhost string
    dump_dbuser string
    dump_dbpass string
    dump_dbport string
    dump_rdbaddr string
)

func init() {
	rootCmd.AddCommand(dumpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dumpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dumpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
    dumpCmd.Flags().StringVarP(&dump_dbname, "dbname", "d", "concurrent", "Database name")
    dumpCmd.Flags().StringVarP(&dump_dbhost, "dbhost", "H", "localhost", "Database host")
    dumpCmd.Flags().StringVarP(&dump_dbuser, "dbuser", "u", "postgres", "Database user")
    dumpCmd.Flags().StringVarP(&dump_dbpass, "dbpassword", "p", "postgres", "Database password")
    dumpCmd.Flags().StringVarP(&dump_dbport, "dbport", "P", "5432", "Database port")
    dumpCmd.Flags().StringVarP(&dump_rdbaddr, "redisAddr", "r", "localhost:6379", "Redis address")
}

func dumpEntity(targetID string) {

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        dump_dbhost, dump_dbuser, dump_dbpass, dump_dbname, dump_dbport)
    redisAddr := dump_rdbaddr

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }

    rdb := redis.NewClient(&redis.Options{
        Addr:     redisAddr,
        Password: "", // no password set
        DB:       0,  // use default DB
    })

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

    // print backup
    b, err := json.MarshalIndent(backup, "", "  ")
    if err != nil {
        panic(err)
    }
    fmt.Println(string(b))
}

type StreamBackup struct {
    Stream core.Stream `json:"stream"`
    Elements []redis.XMessage `json:"elements"`
}


func dumpStream(targetStream string) {

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        dump_dbhost, dump_dbuser, dump_dbpass, dump_dbname, dump_dbport)
    redisAddr := dump_rdbaddr


    ctx := context.Background()

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }

    rdb := redis.NewClient(&redis.Options{
        Addr:     redisAddr,
        Password: "", // no password set
        DB:       0,  // use default DB
    })

    backup := StreamBackup{}

    stream := core.Stream{}
    db.First(&stream, "id = ?", targetStream)
    backup.Stream = stream

    cmd := rdb.XRange(ctx, targetStream, "-", "+")
    backup.Elements, err = cmd.Result()
    if err != nil {
        panic(err)
    }

    b, err := json.MarshalIndent(backup, "", "  ")
    if err != nil {
        panic(err)
    }
    fmt.Println(string(b))
}
