/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"context"
	"github.com/spf13/cobra"
	"github.com/totegamma/concurrent/x/core"
	"github.com/totegamma/concurrent/x/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

func init() {
	rootCmd.AddCommand(entityCmd)

}

func createEntity(ccid string) {
	// create entity
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbhost, dbuser, dbpass, dbname, dbport)

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
		dbhost, dbuser, dbpass, dbname, dbport)

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
