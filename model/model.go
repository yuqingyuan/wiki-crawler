package model

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocolly/colly"
	"math"
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
	ID 			int64    `gorm:"MEDIUMINT;PRIMARY_KEY;AUTO_INCREMENT"`
	Class	 	EventType
	// 是否为公元前
	IsBC	 	bool
	Date 	 	string
	Detail	 	string	 `gorm:"type:LONGTEXT"`
	Links 	 	string   `gorm:"type:LONGTEXT"`
	ImgLinks 	string 	 `gorm:"type:LONGTEXT"`
}

var (
	dateRegexp   = regexp.MustCompile(`^前?\d{1,4}年$`)
	linkRegexp 	 = regexp.MustCompile(`\[\d+]`)
	sourceRegexp = regexp.MustCompile(`\[来源请求]`)
	whichRegexp  = regexp.MustCompile(`\[哪个／哪些？]`)
	whoRegexp    = regexp.MustCompile(`\[谁？]`)
	doubtRegexp  = regexp.MustCompile(`\[可疑 –讨论]`)
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
	// 去除[哪个／哪些？]
	params = whichRegexp.FindAllString(detail, math.MaxInt8)
	for _, param := range params {
		detail = strings.ReplaceAll(detail, param, "")
	}
	// 去除[谁？]
	params = whoRegexp.FindAllString(detail, math.MaxInt8)
	for _, param := range params {
		detail = strings.ReplaceAll(detail, param, "")
	}
	// 去除[可疑 –讨论]
	params = doubtRegexp.FindAllString(detail, math.MaxInt8)
	for _, param := range params {
		detail = strings.ReplaceAll(detail, param, "")
	}

	// Event实例,构建关键字链接
	linksMap := make(map[string]string)
	minLen := math.Min(float64(len(texts)), float64(len(links)))
	for i := 0; i < int(minLen); i++ {
		if strings.Contains(detail, texts[i]) {
			linksMap[texts[i]] = links[i]
		}
	}
	// 事件发生日期
	components := strings.Split(e.Request.URL.String(), "/")
	var eventDate string
	var isBC bool
	if len(components) > 0 {
		result, _ := url.QueryUnescape(components[len(components) - 1])
		// 格式化日期(time库太难用....)
		eventDate, isBC = parseData(year + result)
	}
	result, _ := json.Marshal(linksMap)
	if len(linksMap) == 0 {
		result = make([]byte, 0)
	}
	return Event{0, eventType, isBC, eventDate, detail, string(result), ""}
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