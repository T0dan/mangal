package onepiecetube

import (
	"encoding/json"
	"errors"
	"html"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/metafates/mangal/constant"
	"github.com/metafates/mangal/log"
	"github.com/metafates/mangal/network"
	"github.com/metafates/mangal/source"
	"github.com/metafates/mangal/util"
)

type OnePieceTubePageJson struct {
	Chapter struct {
		Name  string `json:"name"`
		Pages []struct {
			URL    string `json:"url"`
			Height int    `json:"height"`
			Width  int    `json:"width"`
			Type   string `json:"type"`
		} `json:"pages"`
	} `json:"chapter"`
	CurrentPage          string `json:"currentPage"`
	CurrentChapter       string `json:"currentChapter"`
	CurrentChapterID     int    `json:"currentChapterId"`
	CurrentChapterFormat string `json:"currentChapterFormat"`
	CategoryEntries      []struct {
		ID     int    `json:"id"`
		Number int    `json:"number"`
		Name   string `json:"name"`
	} `json:"categoryEntries"`
	NextChapter  string `json:"nextChapter"`
	PrevChapter  any    `json:"prevChapter"`
	CategoryHome struct {
		Href  string `json:"href"`
		Title string `json:"title"`
	} `json:"categoryHome"`
	CategoryLink string `json:"categoryLink"`
	Breadcrumbs  []struct {
		Text string `json:"text"`
		Href string `json:"href"`
	} `json:"breadcrumbs"`
}

func (o *Onepiecetube) PagesOf(chapter *source.Chapter) ([]*source.Page, error) {
	var (
		pages []*source.Page
	)

	req, err := http.NewRequest(http.MethodGet, chapter.URL, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	req.Header.Set("Referer", "https://onepiece-tube.com")
	req.Header.Set("User-Agent", constant.UserAgent)

	resp, err := network.Client.Do(req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	defer util.Ignore(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		err = errors.New("http error: " + resp.Status)
		log.Error(err)
		return nil, err
	}

	buf, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	text := html.UnescapeString(string(buf))

	re := regexp.MustCompile(`<script>window\.__data = (.*);<\/script>`)
	re_match := re.FindStringSubmatch(text)
	if re_match != nil {
		optPages := OnePieceTubePageJson{}

		err = json.Unmarshal([]byte(re_match[1]), &optPages)
		if err != nil {
			return nil, err
		}

		for i, v := range optPages.Chapter.Pages {
			ext := filepath.Ext(v.URL)
			// remove some query params from the extension
			ext = strings.Split(ext, "?")[0]

			pages = append(pages, &source.Page{
				URL:       v.URL,
				Index:     uint16(i),
				Chapter:   chapter,
				Extension: ext,
			})
		}
	}

	chapter.Pages = pages
	return pages, nil
}
