package model

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocolly/colly"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type EventType int8

const (
	EventNormal = iota
	EventBirth
	EventDeath
)

type Event struct {
	Class	EventType
	// 是否为公元前
	IsBC	bool
	Date 	string
	Detail	string	`gorm:"type:LONGTEXT"`
	Links 	string 	`gorm:"type:LONGTEXT"`
	ImgLink string	`gorm:"type:LONGTEXT"`
}
var (
	dateRegexp   = regexp.MustCompile(`^前?\d{1,4}年$`)
	linkRegexp 	 = regexp.MustCompile(`\[\d+]`)
	sourceRegexp = regexp.MustCompile(`\[来源请求]`)
)

// 将抓取到的内容转为对象(历史事件|出生|逝世)
func ProcessEvent(e *colly.HTMLElement, year string, detail string, eventType EventType) Event {
	texts := e.ChildTexts("a")
	links := e.ChildAttrs("a", "href")
	// 去除不必要的链接
	for i := 0; i < len(texts); {
		if dateRegexp.MatchString(texts[i]) {
			texts = append(texts[:i], texts[i+1:]...)
			links = append(links[:i], links[i+1:]...)
		} else {
			i++
		}
	}
	// 去除文献引用
	params := linkRegexp.FindAllString(detail, math.MaxInt8)
	for _, param := range params {
		detail = strings.ReplaceAll(detail, param, "")
	}
	// 去除来源请求
	params = sourceRegexp.FindAllString(detail, math.MaxInt8)
	for _, param := range params {
		detail = strings.ReplaceAll(detail, param, "")
	}
	// Event实例,构建关键字链接
	linksMap := make(map[string]string)
	var keyLink string
	minLen := math.Min(float64(len(texts)), float64(len(links)))
	for i := 0; i < int(minLen); i++ {
		if strings.Contains(detail, texts[i]) {
			linksMap[texts[i]] = links[i]
			if len(keyLink) == 0 {
				keyLink = links[i]
			}
		}
	}
	// 事件发生日期
	components := strings.Split(e.Request.URL.String(), "/")
	var eventDate string
	var isBC bool
	if len(components) > 0 {
		result, _ := url.QueryUnescape(components[len(components) - 1])
		// 格式化日期
		eventDate, isBC = parseData(year + result)
	}
	result, _ := json.Marshal(linksMap)

	event := Event{eventType, isBC, eventDate, detail, string(result), ""}

	if len(keyLink) != 0 && eventType == EventNormal {
		c := colly.NewCollector(colly.Async(true))
		c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: math.MaxInt8})

		c.OnHTML("meta[property=\"og:image\"]", func(e *colly.HTMLElement) {
			event.ImgLink = e.Attr("content")
		})

		c.Request("GET", "https://zh.wikipedia.org"+keyLink, nil, nil, http.Header{"accept-language":[]string{"zh-CN"}})

		c.Wait()
	}

	return event
}

func parseData(date string) (string, bool) {
	date = strings.ReplaceAll(date, "年", "-")
	date = strings.ReplaceAll(date, "月", "-")
	date = strings.ReplaceAll(date, "日", "")
	isBC := strings.Contains(date, "前")
	if isBC {
		date = strings.Replace(date, "前", "", 1)
	}
	return date, isBC
}