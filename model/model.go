package model

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocolly/colly"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type EventType int8

const (
	EventUnknown EventType = -1
	EventNormal = iota
	EventBirth
	EventDeath
)

var (
	lastEventYear  = math.MaxInt64
	eventType 	   = EventUnknown
)

type Event struct {
	Class	EventType
	Date 	string
	Detail	string	`gorm:"type:LONGTEXT"`
	Links 	string 	`gorm:"type:LONGTEXT"`
}

// 将抓取到的内容转为对象(历史事件|出生|逝世,这三者通过年份升序区分,升序->降序)
func ProcessEvent(e *colly.HTMLElement, eventDetail string) Event {
	detail := eventDetail
	texts := e.ChildTexts("a")
	links := e.ChildAttrs("a", "href")
	// 年份
	var year string
	// 去除不必要的年份前缀以及链接
	if len(texts) > 0 {
		eventRegexp := regexp.MustCompile(`^[\d]{1,4}年`)
		params := eventRegexp.FindStringSubmatch(eventDetail)
		for _, param := range params {
			year = param
			detail = strings.Trim(detail, year+"：")
			// 去除年份链接
			if texts[0] == year {
				texts = texts[1:len(texts)]
				if len(links) > 0 {
					links = links[1:len(links)]
				}
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
	// 事件发生日期
	components := strings.Split(e.Request.URL.String(), "/")
	var eventDate string
	if len(components) > 0 {
		result, _ := url.QueryUnescape(components[len(components) - 1])
		// 格式化日期(time库太难用....)
		eventDate = parseData(year + result)
	}
	result, _ := json.Marshal(linksMap)
	return Event{eventType, eventDate, detail, string(result)}
}

func Clear() {
	lastEventYear = math.MaxInt64
	eventType = EventUnknown
}

func parseData(date string) string {
	date = strings.ReplaceAll(date, "年", "-")
	date = strings.ReplaceAll(date, "月", "-")
	date = strings.ReplaceAll(date, "日", "")
	return date
}