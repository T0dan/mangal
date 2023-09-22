package mangaplus

import (
	"strconv"

	"github.com/metafates/mangal/log"
	"github.com/metafates/mangal/source"
)

func (m *Mangaplus) Search(query string) ([]*source.Manga, error) {
	var (
		mangas []*source.Manga
	)

	if m.use_app_api {
		if cached, ok := m.cache.mangas_app.Get(query).Get(); ok {
			for _, manga := range cached {
				manga.Source = m
			}

			return cached, nil
		}

		title_detail_view, err := m.GetAppTitleDetails(query)

		if err != nil {
			log.Error(err)
			return nil, err
		}

		Escape_Path(title_detail_view.Title.TitleName)

		mangas = append(mangas, &source.Manga{
			Name:   Escape_Path(title_detail_view.Title.TitleName),
			URL:    "https://mangaplus.shueisha.co.jp/titles/" + strconv.FormatInt(int64(title_detail_view.Title.TitleId), 10),
			Index:  0,
			ID:     strconv.FormatInt(int64(title_detail_view.Title.TitleId), 10),
			Source: m,
		})

		_ = m.cache.mangas_app.Set(query, mangas)
	} else {
		if cached, ok := m.cache.mangas_web.Get(query).Get(); ok {
			for _, manga := range cached {
				manga.Source = m
			}

			return cached, nil
		}

		title_detail_view, err := m.GetWebTitleDetails(query)

		if err != nil {
			log.Error(err)
			return nil, err
		}

		Escape_Path(title_detail_view.Title.Name)

		mangas = append(mangas, &source.Manga{
			Name:   Escape_Path(title_detail_view.Title.Name),
			URL:    "https://mangaplus.shueisha.co.jp/titles/" + strconv.FormatInt(int64(title_detail_view.Title.TitleId), 10),
			Index:  0,
			ID:     strconv.FormatInt(int64(title_detail_view.Title.TitleId), 10),
			Source: m,
		})

		_ = m.cache.mangas_web.Set(query, mangas)
	}

	return mangas, nil
}
