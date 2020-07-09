package crawl

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
	"wiki-crawler/model"
)

// 抓取Wiki历史上的今天每个月份的链接
func HomeLinks(completion func([]string)) {
	links := make([]string, 0)

	c := colly.NewCollector()

	c.OnHTML(".wikitable a", func(e *colly.HTMLElement) {
		links = append(links, e.Request.AbsoluteURL(e.Attr("href")))
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Error Occurred when crawl home links: ", err)
	})

	c.OnScraped(func(r *colly.Response) {
		completion(links)
	})

	c.Visit("https://zh.wikipedia.org/wiki/%E5%8E%86%E5%8F%B2%E4%B8%8A%E7%9A%84%E4%BB%8A%E5%A4%A9")
}

var (
	eventRegexp  = regexp.MustCompile(`^前?\d{1,4}年.*`)
	dateRegexp   = regexp.MustCompile(`^前?\d{1,4}年`)
	filterRegexp = regexp.MustCompile(`前?\d{1,4}年+[:：︰﹕，；－—]+[—]?`)
)

// 抓取Wiki历史上的今天
func DailyEvent(link string, completion func([]model.Event)) {
	events := make([]model.Event, 0)

	c := colly.NewCollector()

	// 大事记
	c.OnHTML("h3+ul>li", func(e *colly.HTMLElement) {
		filterEvents(e, &events, model.EventNormal)
	})

	// 出生
	c.OnHTML("h2:has(span#出生)+ul>li", func(e *colly.HTMLElement) {
		filterEvents(e, &events, model.EventBirth)
	})

	// 逝世
	c.OnHTML("h2:has(span#逝世)+ul>li", func(e *colly.HTMLElement) {
		filterEvents(e, &events, model.EventDeath)
	})

	// 当前链接所有事件
	wg := sync.WaitGroup{}
	c.OnScraped(func(r *colly.Response) {
		// 抓取二级页面图片
		wg.Add(len(events))
		for i, v := range events {
			// FIXME: 防止并发量过大，暂时先这么处理
			time.Sleep(40 * time.Millisecond)

			go func(index int, value model.Event) {
				analysisSecondaryPage(value, func(imgLinks string) {
					events[index].ImgLinks = imgLinks
					wg.Done()
				})
			}(i, v)
		}
		wg.Wait()
		// 回调数据
		completion(events)
	})

	c.OnError(func(r *colly.Response, err error) {
		//fmt.Println(err)
		completion(events)
	})

	// 指定wiki网页为简体中文
	c.Request("GET", link, nil, nil, http.Header{"accept-language":[]string{"zh-CN"}})
}

// 获取二级页面链接
func analysisSecondaryPage(event model.Event, completion func(imgLinks string)) {
	links := make(map[string]string)
	json.Unmarshal([]byte(event.Links), &links)

	pageLinks := make([]string, 0)
	for _, link := range links {
		if !strings.HasPrefix(link, "https://") {
			pageLinks = append(pageLinks, "https://zh.wikipedia.org"+link)
		} else {
			pageLinks = append(pageLinks, link)
		}
	}

	getPictureLink(pageLinks, func(link string) {
		completion(link)
	})
}

// 抓取二级页面图片链接
func getPictureLink(pageLinks []string, completion func(link string)) {
	links := make([]string, 0)

	c := colly.NewCollector(colly.Async(true))

	q, _ := queue.New(len(pageLinks), &queue.InMemoryQueueStorage{MaxSize: 100000000})

	c.OnHTML("meta[property=\"og:image\"]", func(e *colly.HTMLElement) {
		link := strings.ReplaceAll(e.Attr("content"), "https://upload.wikimedia.org/wikipedia", "")
		links = append(links, link)
	})

	c.OnError(func(r *colly.Response, err error) {
		//fmt.Println(err)
	})

	for _, link := range pageLinks {
		q.AddURL(link)
	}

	q.Run(c)

	c.Wait()

	completion(strings.Join(links, ","))
}

// 信息初步过滤
func filterEvents(e *colly.HTMLElement, event *[]model.Event, eventType model.EventType) {
	year := dateRegexp.FindString(e.Text)
	for _, param := range formatAndRegularText(e.Text) {
		for _, text := range removeDateAndSplitText(param, year) {
			if len(text) != 0 {
				*event = append(*event, model.ProcessEvent(e, year, text, eventType))
			}
		}
	}
}

// 去除换行以及首个空格,存在一行多个事件用&&分割
func formatAndRegularText(target string) []string {
	target = strings.ReplaceAll(target, "\n", "&&")
	target = strings.Replace(target, " ", "", 1)
	target = strings.ReplaceAll(target, " ", "")
	return eventRegexp.FindStringSubmatch(target)
}

// 去除年份前缀
func removeDateAndSplitText(target string, year string) []string  {
	target = filterRegexp.ReplaceAllString(target, "")
	if strings.Contains(target, year) {
		target = strings.Replace(target, year, "", 1)
	}
	return strings.Split(target, "&&")
}