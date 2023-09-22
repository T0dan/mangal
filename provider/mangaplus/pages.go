package mangaplus

import (
	"path/filepath"
	"strings"

	"github.com/metafates/mangal/log"
	"github.com/metafates/mangal/source"
)

func (m *Mangaplus) PagesOf(chapter *source.Chapter) ([]*source.Page, error) {
	var (
		pages []*source.Page
	)

	if m.use_app_api {

	} else {
		viewer, err := m.GetWebViewer(chapter.ID)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		for i, page := range viewer.Pages {
			if page.MangaPage != nil && page.MangaPage.ImageUrl != "" {
				ext := filepath.Ext(page.MangaPage.ImageUrl)
				// remove some query params from the extension
				ext = strings.Split(ext, "?")[0]

				pages = append(pages, &source.Page{
					URL:          page.MangaPage.ImageUrl,
					Index:        uint16(i),
					Chapter:      chapter,
					Extension:    ext,
					MangaPlusKey: page.MangaPage.EncryptionKey,
				})
			}
		}
	}

	chapter.Pages = pages
	return pages, nil
}
