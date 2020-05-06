package main

import (
	"fmt"
	"wiki-crawler/crawl"
	"wiki-crawler/model"
)

func main() {
	crawl.HomeLinks(func(links []string) {
		if len(links) == 0 {
			fmt.Println("获取到的首页链接为空")
			return
		}
		fmt.Println("链接数量: ", len(links))
		// 抓取每日详细内容,并写入数据库
		crawl.DailyEvent(links, func(events []model.Event) {
			//TODO: 写入数据库
			model.CreateDataTableIfNotExists()
			model.InsertIntoDataTable(events)
		})
	})

	defer model.CloseDB()
}