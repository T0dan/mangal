package generic

import (
	"net/http"

	"github.com/gocolly/colly/v2"
	"github.com/metafates/mangal/source"
	"github.com/metafates/mangal/util"
)

// PagesOf given source.Chapter
func (s *Scraper) PagesOf(chapter *source.Chapter) ([]*source.Page, error) {
	address_path := util.UrlGetPath(chapter.URL)
	if pages, ok := s.pages[address_path]; ok {
		return pages, nil
	}

	ctx := colly.NewContext()
	ctx.Put("chapter", chapter)
	err := s.pagesCollector.Request(http.MethodGet, chapter.URL, nil, ctx, nil)

	if err != nil {
		return nil, err
	}

	s.pagesCollector.Wait()

	return s.pages[address_path], nil
}
