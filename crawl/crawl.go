package crawl

import (
	"fmt"
	"github.com/gocolly/colly"
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

var events = make([]model.Event, 0)

// 抓取Wiki历史上的今天
func DailyEvent(links []string, completion func([]model.Event)) {
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnHTML("ul>li", func(e *colly.HTMLElement) {
		if e.Name == "li" {
			// 指定抓取内容
			eventRegexp := regexp.MustCompile(`^[\d]{1,4}年\D.*`)
			// 去除换行以及首个空格
			e.Text = strings.ReplaceAll(e.Text, "\n", " ")
			e.Text = strings.Replace(e.Text, " ", "", 1)
			// 正则匹配
			params := eventRegexp.FindStringSubmatch(e.Text)
			for _, param := range params {
				events = append(events, model.ProcessEvent(e, param))
			}
		}
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finish crawling daily event, return events")
		model.Clear()
		completion(events)
	})

	for _, v := range links {
		c.Visit(v)
	}
}