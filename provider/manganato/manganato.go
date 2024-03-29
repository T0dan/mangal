package manganato

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/metafates/mangal/provider/generic"
)

var Config = &generic.Configuration{
	Name:            "Manganato",
	StdLang:         "en",
	Delay:           50 * time.Millisecond,
	Parallelism:     50,
	ReverseChapters: true,
	BaseURL:         "https://manganato.com/",
	GenerateSearchURL: func(query string) string {
		query = strings.ReplaceAll(query, " ", "_")
		query = strings.TrimSpace(query)
		query = strings.ToLower(query)
		query = url.QueryEscape(query)
		template := "https://chapmanganato.com/https://manganato.com/search/story/%s"
		return fmt.Sprintf(template, query)
	},
	MangaExtractor: &generic.Extractor{
		Selector: "div.search-story-item",
		Name: func(selection *goquery.Selection) string {
			return strings.TrimSpace(selection.Find("a.item-title").Text())
		},
		URL: func(selection *goquery.Selection) string {
			return selection.Find("a.item-title").AttrOr("href", "")
		},
		Cover: func(selection *goquery.Selection) string {
			return selection.Find("img").AttrOr("src", "")
		},
		Language: func(selection *goquery.Selection) string {
			return "en"
		},
	},
	ChapterExtractor: &generic.Extractor{
		Selector: "li.a-h",
		Name: func(selection *goquery.Selection) string {
			name := selection.Find("a").Text()
			if strings.HasPrefix(name, "Vol.") {
				splitted := strings.Split(name, " ")
				name = strings.Join(splitted[1:], " ")
			}
			return name
		},
		Number: func(selection *goquery.Selection) string {
			number := selection.Find("a").Text()
			re := regexp.MustCompile(`Chapter (\d+\.?\d*)`)
			re_match := re.FindStringSubmatch(number)
			if re_match != nil {
				return re_match[1]
			} else {
				return ""
			}
		},
		URL: func(selection *goquery.Selection) string {
			return selection.Find("a").AttrOr("href", "")
		},
		Volume: func(selection *goquery.Selection) string {
			name := selection.Find("a").Text()
			if strings.HasPrefix(name, "Vol.") {
				splitted := strings.Split(name, " ")
				return splitted[0]
			}
			return ""
		},
	},
	PageExtractor: &generic.Extractor{
		Selector: ".container-chapter-reader img",
		URL: func(selection *goquery.Selection) string {
			return selection.AttrOr("src", "")
		},
	},
}
