package main

import (
	"fmt"
	"sync"
	"wiki-crawler/crawl"
	"wiki-crawler/model"
)

var (
	wg        = sync.WaitGroup{}
	wl 	      = sync.Mutex{}
	allEvents = make([]model.Event, 0)
	progress  = 0
)

func main() {
	model.CreateDataTableIfNotExists()

	crawl.HomeLinks(func(links []string) {
		taskNum := len(links)
		if taskNum == 0 {
			return
		}

		for _, link := range links {
			wg.Add(1)
			// FIXME: 每次执行一个任务，内部会并发去抓取图片，如果这里不限制会导致并发量过大
			// 抓取每日详细内容
			go func(val string) {
				crawl.DailyEvent(val, func(events []model.Event) {
					allEvents = append(allEvents, events...)
					progress += 1
					fmt.Printf("任务进度:%.2f\r", (float32(progress)/float32(taskNum))*100)
					wl.Unlock()
					wg.Done()
				})
			}(link)
			wl.Lock()
		}
	})

	defer model.CloseDB()

	// 阻塞
	wg.Wait()

	fmt.Println("任务完成, 写入数据库, 总数据量:", len(allEvents))

	model.InsertIntoDataTable(allEvents)
}