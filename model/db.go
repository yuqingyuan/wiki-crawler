package model

import (
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
)

var (
	database *gorm.DB

	username = "root"
	password = "yourpassword"
	dbName 	 = "wiki_crawler"
)

func init() {
	database, _ = gorm.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, dbName))
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
		database.Create(&event)
	}
}

func FindAllEventsLinks() (*sql.Rows, error) {
	return database.Raw("select *from events").Rows()
}

func EventsCount() (count int64) {
	database.Model(&Event{}).Count(&count)
	return count
}

func UpdateEvent(event Event) {
	database.Model(&event).Update("img_links", event.ImgLinks)
}