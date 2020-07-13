package main

import (
	"fmt"
	"sync"
	"wiki-crawler/crawl"
	"wiki-crawler/model"
)

var (
	wg        = sync.WaitGroup{}
	allEvents = make([]model.Event, 0)
	progress  = 0
)

func main() {
	crawl.HomeLinks(func(links []string) {
		taskNum := len(links)
		if taskNum == 0 {
			return
		}

		for _, link := range links {
			wg.Add(1)
			// FIXME: 每次执行一个任务，内部会并发去抓取图片，如果这里不限制会导致并发量过大
			// 抓取每日详细内容
			crawl.DailyEvent(link, func(events []model.Event) {
				allEvents = append(allEvents, events...)
				progress += 1
				wg.Done()
				// 更新进度
				fmt.Printf("任务进度:%.2f\r", (float32(progress)/float32(taskNum))*100)
			})
			// 阻塞
			wg.Wait()
		}
	})

	defer model.CloseDB()

	fmt.Println("任务完成, 写入数据库, 总数据量:", len(allEvents))

	model.CreateDataTableIfNotExists()
	model.InsertIntoDataTable(allEvents)
}