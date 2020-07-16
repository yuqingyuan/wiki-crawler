package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"os"
	"os/exec"
)

var (
	database *gorm.DB

	username = "root"
	password = "root"
	dbName 	 = "wiki_crawler"
)

func init() {
	verify()
	database, _ = gorm.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, dbName))
}

func verify() {
	// 输入数据库相关信息
	fmt.Println("Please enter your database's name: ")
	fmt.Scanln(&dbName)
	fmt.Println("Please enter your database's username: ")
	fmt.Scanln(&username)
	fmt.Println("Please enter your database's password: ")
	fmt.Scanln(&password)
	// 清屏
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
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
	}
}

func InsertIntoDataTable(events []Event) {
	for _, event := range events {
		database.Create(&event)
	}
}