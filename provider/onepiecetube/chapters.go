package onepiecetube

import (
	"encoding/json"
	"errors"
	"fmt"
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

type OnePieceTubeChapterJson struct {
	Options struct {
		Livesearch bool   `json:"livesearch"`
		IsChapter  bool   `json:"isChapter"`
		IsEpisode  bool   `json:"isEpisode"`
		Segmented  bool   `json:"segmented"`
		SegmentCat string `json:"segment_cat"`
		SegmentID  string `json:"segment_id"`
	} `json:"options"`
	Category struct {
		ID            int    `json:"id"`
		Type          string `json:"type"`
		Status        int    `json:"status"`
		ShowFrontpage int    `json:"show_frontpage"`
		Order         int    `json:"order"`
		Name          string `json:"name"`
		NamePlural    string `json:"name_plural"`
		EntityFormat  string `json:"entity_format"`
		StreamFormat  string `json:"stream_format"`
		Slug          string `json:"slug"`
		EntrySlug     string `json:"entry_slug"`
		Description   string `json:"description"`
		Published     string `json:"published"`
		Affiliates    []any  `json:"affiliates"`
	} `json:"category"`
	Specials []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"specials"`
	Arcs []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Min  int    `json:"min"`
		Max  int    `json:"max"`
	} `json:"arcs"`
	Entries []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Number      int    `json:"number"`
		CategoryID  int    `json:"category_id"`
		ArcID       int    `json:"arc_id"`
		SpecialsID  int    `json:"specials_id"`
		Lang        string `json:"lang"`
		Pages       int    `json:"pages"`
		IsAvailable bool   `json:"is_available"`
		Date        string `json:"date"`
		Href        string `json:"href"`
	} `json:"entries"`
}

func (o *Onepiecetube) ChaptersOf(manga *source.Manga) ([]*source.Chapter, error) {
	if cached, ok := o.cache.chapters.Get(manga.URL).Get(); ok {
		for _, chapter := range cached {
			chapter.Manga = manga
		}

		return cached, nil
	}

	var (
		chapters []*source.Chapter
	)

	req, err := http.NewRequest(http.MethodGet, manga.URL, nil)
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
		optChapters := OnePieceTubeChapterJson{}

		err = json.Unmarshal([]byte(re_match[1]), &optChapters)
		if err != nil {
			return nil, err
		}

		for _, v := range optChapters.Entries {
			if v.IsAvailable {
				number := fmt.Sprintf("%d", v.Number)
				name := strings.TrimSpace(v.Name)
				chapters = append(chapters, &source.Chapter{
					Name:   name,
					Index:  uint16(v.Number),
					Number: number,
					ID:     filepath.Base(v.Href),
					URL:    v.Href,
					Manga:  manga,
					Volume: "",
				})
			}
		}
	}

	reversed := make([]*source.Chapter, len(chapters))
	for i, chapter := range chapters {
		reversed[len(chapters)-i-1] = chapter
		chapter.Index = uint16(len(chapters) - i - 1)
		chapter.Index++
	}

	manga.Chapters = reversed
	_ = o.cache.chapters.Set(manga.URL, reversed)
	return reversed, nil
}
