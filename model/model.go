package model

import (
	"fmt"
	"github.com/jinzhu/gorm"

	_ "github.com/go-sql-driver/mysql"
)

var (
	DB *gorm.DB

	username = "root"
	password = "bsb@1993BSB"
	dbName = "wiki_crawler"
)

func init() {
	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", username, password, dbName))
	if err != nil {
		fmt.Println(err)
	}
	DB = db
}