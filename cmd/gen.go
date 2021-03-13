package cmd

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/DATA-DOG/go-sqlmock"
	dynamicstruct "github.com/ompluscator/dynamic-struct"
	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db      string
	srcFile string
)

type User struct {
	ID   string `gorm:"primary_key"`
	Name string
}

type Another struct {
	ID   string `gorm:"primary_key"`
	Name string
}

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "generate sql from struct using gorm.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		r := regexp.MustCompile(`'CREATE TABLE .*?'`)

		rec := bytes.NewBuffer([]byte{})
		db, err := getDBMock(rec, db)
		if err != nil {
			log.Fatalf("%v", err)
		}

		ss, err := getStruct(srcFile)
		if err != nil {
			log.Fatalf("%v", err)
		}

		for _, sitem := range ss {
			instance := dynamicstruct.NewStruct()
			for _, smitem := range sitem.StructMetaSlice {
				instance.AddField(smitem.Field, smitem.TypeExampleVal, smitem.Tag)
			}
			v := instance.Build().New()
			db.AutoMigrate(v)
		}

		out := rec.String()
		sep := strings.Split(out, "\r\n")
		var sqlIdx int
		for _, item := range sep {
			result := r.FindAllStringSubmatch(item, -1)
			if len(result) > 0 {
				sql := strings.Replace(result[0][0], "'", "", -1)
				fmt.Printf("%v\n", strings.Replace(sql, "``", "`"+ss[sqlIdx].TableName()+"`", 1))
				sqlIdx++
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(genCmd)

	genCmd.Flags().StringVarP(&db, "db", "", "mysql", "database (default: mysql)")
	genCmd.Flags().StringVarP(&srcFile, "srcFile", "f", "", "source file (required)")
	genCmd.MarkFlagRequired("srcFile")
}

func getDBMock(recorder io.Writer, dbType string) (*gorm.DB, error) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		return nil, err
	}

	newLogger := logger.New(
		log.New(recorder, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			LogLevel: logger.Warn, // Log level
			Colorful: false,       // Disable color
		},
	)

	var gormDB *gorm.DB
	switch dbType {
	case "mysql":
		gormDB, err = gorm.Open(mysql.New(mysql.Config{
			DriverName:                "my_mysql_driver",
			Conn:                      sqlDB,
			SkipInitializeWithVersion: true, // auto configure based on currently MySQL version
		}), &gorm.Config{
			Logger: newLogger,
		})
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unexpected db: %v", dbType)
	}

	return gormDB, nil
}

type Struct struct {
	Name            string
	StructMetaSlice []StructMeta
}

func (a Struct) TableName() string {
	return strings.ToLower(a.Name) + "s"
}

type StructMeta struct {
	Field          string
	Type           string
	TypeExampleVal interface{}
	Tag            string
}

func getStruct(srcFile string) ([]Struct, error) {
	bs, err := ioutil.ReadFile(srcFile)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(fset, "", string(bs), parser.ParseComments)
	if err != nil {
		return nil, err
	}

	sms := []Struct{}

	for _, item := range f.Scope.Objects {
		if item.Kind == ast.Typ {
			s := Struct{
				Name: item.Decl.(*ast.TypeSpec).Name.Name,
			}
			v := reflect.ValueOf(&item.Decl.(*ast.TypeSpec).Type).Elem()
			for _, fitem := range v.Interface().(*ast.StructType).Fields.List {
				n := fitem.Names[0]
				if n.Obj.Kind == ast.Var {
					f := n.Obj.Decl.(*ast.Field)
					rv := reflect.ValueOf(&f.Type).Elem()
					sm := StructMeta{
						Field: n.Name,
					}
					switch tv := rv.Interface().(type) {
					case *ast.Ident:
						sm.Type = tv.Name
					case *ast.ArrayType:
						arv := reflect.ValueOf(&tv.Elt).Elem()
						sm.Type = arv.Interface().(*ast.Ident).Name
					}
					if f.Tag != nil {
						sm.Tag = f.Tag.Value
					}
					switch sm.Type {
					case "string":
						sm.TypeExampleVal = ""
					case "int", "int32", "int64":
						sm.TypeExampleVal = 0
					case "float32", "float64":
						sm.TypeExampleVal = 0.1
					case "byte":
						sm.TypeExampleVal = []byte{}
					case "bool":
						sm.TypeExampleVal = false
					}
					if sm.TypeExampleVal != nil {
						s.StructMetaSlice = append(s.StructMetaSlice, sm)
					}

				}
			}
			sms = append(sms, s)
		}
	}

	return sms, nil
}
