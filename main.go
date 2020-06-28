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

	eventsSum := model.EventsCount()
	curIndex := 0

	// 抓取图片
	rows, err := model.FindAllEventsLinks()
	if err != nil {
		return
	}
	for rows.Next() {
		event := model.Event{}
		err := rows.Scan(&event.ID, &event.Class, &event.IsBC, &event.Date, &event.Detail, &event.Links, &event.ImgLinks)
		if err != nil {
			fmt.Println("Scan err", err)
		}
		crawl.EventPictures(&event, func(err2 error) {
			model.UpdateEvent(event)
			curIndex += 1
			// 打印进度
			if err2 != nil {
				fmt.Println("Occurred error, skip")
				return
			}
			fmt.Printf("图片抓取进度:%d/%d\n", curIndex, eventsSum)
		})
	}

	defer func(){
		rows.Close()
		model.CloseDB()
	}()
}