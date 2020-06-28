package main

import (
	"fmt"
	"wiki-crawler/crawl"
	"wiki-crawler/model"
)

func main() {
	model.CreateDataTableIfNotExists()

	crawl.HomeLinks(func(links []string) {
		if len(links) == 0 {
			fmt.Println("获取到的首页链接为空")
			return
		}
		fmt.Println("链接数量: ", len(links))
		// 抓取每日详细内容,并写入数据库
		crawl.DailyEvent(links, func(events []model.Event) {
			// 写入数据库
			model.InsertIntoDataTable(events)
		})
	})

	// 抓取图片
	rows, err := model.FindAllEventsLinks()
	if err != nil {
		return
	}
	event := model.Event{}
	for rows.Next() {
		rows.Scan(&event.Class, &event.IsBC, &event.Date, &event.Detail, &event.Links, &event.ImgLinks)
		crawl.EventPictures(event, func() {
			model.UpdateEvent(event)
		})
	}

	defer model.CloseDB()
}