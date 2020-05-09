package crawl

import (
	"fmt"
	"github.com/gocolly/colly"
	"net/http"
	"regexp"
	"strings"
	"wiki-crawler/model"
)

// 抓取Wiki历史上的今天每个月份的链接
func HomeLinks(completion func([]string)) {
	links := make([]string, 0)

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnHTML(".wikitable a", func(e *colly.HTMLElement) {
		// Get link
		link := e.Request.AbsoluteURL(e.Attr("href"))
		// Store link
		links = append(links, link)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Error Occurred when crawl home links: ", err)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finish crawling home links")
		//回调首页所有链接
		completion(links)
	})

	c.Visit("https://zh.wikipedia.org/wiki/%E5%8E%86%E5%8F%B2%E4%B8%8A%E7%9A%84%E4%BB%8A%E5%A4%A9")
}

var (
	events []model.Event
	eventRegexp = regexp.MustCompile(`^前?\d{1,4}年.*`)
)

// 抓取Wiki历史上的今天
func DailyEvent(links []string, completion func([]model.Event)) {
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
		events = make([]model.Event, 0)
	})

	// 大事记
	c.OnHTML("h3+ul>li", func(e *colly.HTMLElement) {
		for _, param := range formatAndRegularText(e.Text) {
			events = append(events, model.ProcessEvent(e, param, model.EventNormal))
		}
	})

	// 出生
	c.OnHTML("h2:has(span#出生)+ul>li", func(e *colly.HTMLElement) {
		for _, param := range formatAndRegularText(e.Text) {
			events = append(events, model.ProcessEvent(e, param, model.EventBirth))
		}
	})

	// 逝世
	c.OnHTML("h2:has(span#逝世)+ul>li", func(e *colly.HTMLElement) {
		for _, param := range formatAndRegularText(e.Text) {
			events = append(events, model.ProcessEvent(e, param, model.EventDeath))
		}
	})

	// 回调当前链接所有事件
	c.OnScraped(func(r *colly.Response) {
		completion(events)
		fmt.Println("Finish crawling daily event, return events")
	})

	for _, v := range links {
		// 指定wiki网页为简体中文
		c.Request("GET", v, nil, nil, http.Header{"accept-language":[]string{"zh-CN"}})
	}
}

// 去除换行以及首个空格
func formatAndRegularText(target string) []string {
	target = strings.ReplaceAll(target, "\n", " ")
	target = strings.Replace(target, " ", "", 1)
	return eventRegexp.FindStringSubmatch(target)
}