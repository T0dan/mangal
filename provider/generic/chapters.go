package generic

import (
	"net/http"

	"github.com/gocolly/colly/v2"
	"github.com/metafates/mangal/source"
	"github.com/metafates/mangal/util"
)

// ChaptersOf given source.Manga
func (s *Scraper) ChaptersOf(manga *source.Manga) ([]*source.Chapter, error) {
	address_path := util.UrlGetPath(manga.URL)
	if chapters, ok := s.chapters[address_path]; ok {
		return chapters, nil
	}

	ctx := colly.NewContext()
	ctx.Put("manga", manga)
	err := s.chaptersCollector.Request(http.MethodGet, manga.URL, nil, ctx, nil)

	if err != nil {
		return nil, err
	}

	s.chaptersCollector.Wait()

	if s.config.ReverseChapters {
		// reverse chapters
		chapters := s.chapters[address_path]
		reversed := make([]*source.Chapter, len(chapters))
		for i, chapter := range chapters {
			reversed[len(chapters)-i-1] = chapter
			chapter.Index = uint16(len(chapters) - i - 1)
			chapter.Index++
		}

		s.chapters[address_path] = reversed
	}

	return s.chapters[address_path], nil
}
