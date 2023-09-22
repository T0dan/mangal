package mangaplus

import (
	"fmt"
	"strconv"

	"github.com/metafates/mangal/log"
	"github.com/metafates/mangal/source"
)

func (m *Mangaplus) ChaptersOf(manga *source.Manga) ([]*source.Chapter, error) {
	var (
		chapters []*source.Chapter
	)

	if m.use_app_api {
		if cached, ok := m.cache.chapters_app.Get(manga.URL).Get(); ok {
			for _, chapter := range cached {
				chapter.Manga = manga
			}

			return cached, nil
		}

		_ = m.cache.chapters_app.Set(manga.URL, chapters)
	} else {
		if cached, ok := m.cache.chapters_web.Get(manga.ID).Get(); ok {
			for _, chapter := range cached {
				chapter.Manga = manga
			}

			return cached, nil
		}

		title_detail_view, err := m.GetWebTitleDetails(manga.ID)

		if err != nil {
			log.Error(err)
			return nil, err
		}

		i := uint16(0)

		lastNumber := "0"
		lastSubNumber := int(1)

		for _, chapter := range append(title_detail_view.FirstChapterList, title_detail_view.LastChapterList...) {

			number := ""

			if Is_Oneshot(chapter.Name, chapter.SubTitle) {
				number = "1"
			} else if Is_Extra(chapter.Name) {
				number = fmt.Sprintf("%s.%d", lastNumber, lastSubNumber)
				lastSubNumber++
			} else {
				number_int, err := Chapter_Name_To_Int(chapter.Name)
				if err == nil {
					number = strconv.FormatInt(int64(number_int), 10)
					lastNumber = number
					lastSubNumber = 1
				} else {
					number = fmt.Sprintf("%s.%d", lastNumber, lastSubNumber)
					lastSubNumber++
				}
			}

			chapters = append(chapters, &source.Chapter{
				Name:   chapter.SubTitle,
				Index:  i,
				Number: number,
				ID:     strconv.FormatInt(int64(chapter.ChapterId), 10),
				URL:    "https://mangaplus.shueisha.co.jp/viewer/" + strconv.FormatInt(int64(chapter.ChapterId), 10),
				Manga:  manga,
				Volume: "",
			})
		}

		_ = m.cache.chapters_web.Set(manga.ID, chapters)
	}

	return chapters, nil
}
