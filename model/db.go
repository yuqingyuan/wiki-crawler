package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

var (
	database *gorm.DB

	username = "root"
	password = "你的数据库密码"
	dbName = "wiki_crawler"
)

func init() {
	database, _ = gorm.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", username, password, dbName))
}

func CloseDB() {
	if database != nil {
		err := database.Close()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func CreateDataTableIfNotExists() {
	if !database.HasTable(&Event{}) {
		database.CreateTable(&Event{})
		fmt.Println(database)
	}
}

func DeleteDataTable() {
	database.DropTableIfExists(&Event{})
}

func InsertIntoDataTable(events []Event) {
	for _, event := range events {
		database.Create(event)
	}
}