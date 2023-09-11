/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"context"
    "gorm.io/driver/postgres"
	"github.com/spf13/cobra"
    "gorm.io/gorm"
    "github.com/totegamma/concurrent/x/core"
    "github.com/totegamma/concurrent/x/entity"
)

// entityCmd represents the entity command
var entityCmd = &cobra.Command{
	Use:   "entity",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {

		mode := args[0]
		target := args[1]

		if mode == "add" {
			createEntity(target)
		} else if mode == "role" {
			role := args[2]
			setEntityRole(target, role)
		} else {
			fmt.Println("unknown mode")
		}
	},
}

var (
    entity_dbname string
    entity_dbhost string
    entity_dbuser string
    entity_dbpass string
    entity_dbport string
    entity_rdbaddr string
)

func init() {
	rootCmd.AddCommand(entityCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// entityCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// entityCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

    entityCmd.Flags().StringVarP(&entity_dbname, "dbname", "d", "concurrent", "Database name")
    entityCmd.Flags().StringVarP(&entity_dbhost, "dbhost", "H", "localhost", "Database host")
    entityCmd.Flags().StringVarP(&entity_dbuser, "dbuser", "u", "postgres", "Database user")
    entityCmd.Flags().StringVarP(&entity_dbpass, "dbpassword", "p", "postgres", "Database password")
    entityCmd.Flags().StringVarP(&entity_dbport, "dbport", "P", "5432", "Database port")
    entityCmd.Flags().StringVarP(&entity_rdbaddr, "redisAddr", "r", "localhost:6379", "Redis address")
}

func createEntity(ccid string) {
	// create entity
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        dump_dbhost, dump_dbuser, dump_dbpass, dump_dbname, dump_dbport)

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }

	repo := entity.NewRepository(db)

	ctx := context.Background()
	repo.Create(ctx, &core.Entity{
		ID:   ccid,
		Tag:  "",
		Meta: "{}",
	})
}

func setEntityRole(ccid string, role string) {
	// set entity role
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        dump_dbhost, dump_dbuser, dump_dbpass, dump_dbname, dump_dbport)

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }

	repo := entity.NewRepository(db)

	ctx := context.Background()
	entity, err := repo.Get(ctx, ccid)
	if err != nil {
		panic(err)
	}

	entity.Tag = role
	repo.Update(ctx, &entity)
}

