package crawl

import (
	"fmt"
	"github.com/gocolly/colly"
	"math"
	"regexp"
	"strconv"
	"strings"
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

var events = make([]Event, 0)

// 抓取Wiki历史上的今天
func DailyEvent(links []string, completion func([]Event)) {
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnHTML("ul>li", func(e *colly.HTMLElement) {
		if e.Name == "li" {
			// 指定抓取内容
			eventRegexp := regexp.MustCompile(`^[\d]{1,4}年\D.*`)
			params := eventRegexp.FindStringSubmatch(e.Text)
			for _, param := range params {
				events = append(events, processEvent(e, param))
			}
		}
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finish crawling daily event, return events")
		completion(events)
	})

	for _, v := range links {
		c.Visit(v)
	}
}

type EventType int8

const (
	EventUnknown EventType = -1
	EventNormal = iota
	EventBirth
	EventDeath
)

var lastEventYear  = math.MaxInt64
var eventType 	   = EventUnknown

type Event struct {
	class	EventType
	date 	string
	detail	string
	links 	map[string]string
}

// 将抓取到的内容转为对象(历史事件|出生|逝世,这三者通过年份升序区分,升序->降序)
func processEvent(e *colly.HTMLElement, event string) Event {
	detail := event
	texts := e.ChildTexts("a")
	links := e.ChildAttrs("a", "href")
	// 年份
	var year string
	// 去除不必要的年份前缀以及链接
	if len(texts) > 0 {
		eventRegexp := regexp.MustCompile(`^[\d]{1,4}年`)
		params := eventRegexp.FindStringSubmatch(event)
		for _, param := range params {
			year = param
			detail = strings.Trim(detail, year+"：")
			if texts[0] == year {
				texts = texts[1:len(texts)]
			}
			if len(links) > 0 && links[0] == year {
				links = links[1:len(links)]
			}
		}
	}
	// Event实例
	linksMap := make(map[string]string)
	minLen := math.Min(float64(len(texts)), float64(len(links)))
	for i := 0; i < int(minLen); i++ {
		linksMap[texts[i]] = links[i]
	}
	// 根据年份升序与否区分事件类型
	var curYear, _ = strconv.Atoi(strings.Trim(year, "年"))
	if curYear < lastEventYear {
		eventType += 1
	}
	lastEventYear = curYear
	return Event{eventType, year, detail, linksMap}
}