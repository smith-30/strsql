package cmd

import (
	"log"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/spf13/cobra"
)

var (
	db string
)

type User struct {
	ID   string `gorm:"primary_key"`
	Name string
}

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "generate sql from struct using gorm.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		db, _, err := getDBMock(db)
		if err != nil {
			log.Fatalf("%v", err)
		}
		defer db.Close()
		db.LogMode(true)

		db.AutoMigrate(&User{})
	},
}

func init() {
	rootCmd.AddCommand(genCmd)

	genCmd.Flags().StringVarP(&db, "db", "", "mysql", "database (default: mysql)")
}

func getDBMock(dbType string) (*gorm.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	gdb, err := gorm.Open(dbType, db)
	if err != nil {
		return nil, nil, err
	}
	return gdb, mock, nil
}
