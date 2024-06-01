package mangaplus

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/metafates/mangal/log"
	"github.com/metafates/mangal/provider/mangaplus/mangaplus_resp_app"
	"github.com/metafates/mangal/source"
)

func (m *Mangaplus) ChaptersOf(manga *source.Manga) ([]*source.Chapter, error) {
	var (
		chapters []*source.Chapter
	)

	if m.use_app_api {
		if cached, ok := m.cache.chapters_app.Get(manga.ID).Get(); ok {
			for _, chapter := range cached {
				chapter.Manga = manga
			}

			return cached, nil
		}

		title_detail_view, err := m.GetAppTitleDetails(manga.ID)

		_ = title_detail_view
		if err != nil {
			log.Error(err)
			return nil, err
		}

		var (
			api_chapters []*mangaplus_resp_app.Chapter
		)

		for _, chapter_collection := range title_detail_view.Chapters {
			for _, chapter_desc := range chapter_collection.FirstChapterList {
				api_chapters = append(api_chapters, chapter_desc)
			}
			for _, chapter_desc := range chapter_collection.ChapterList {
				api_chapters = append(api_chapters, chapter_desc)
			}
			for _, chapter_desc := range chapter_collection.LastChapterList {
				api_chapters = append(api_chapters, chapter_desc)
			}
		}

		i := uint16(0)

		lastNumber := "0"
		lastSubNumber := int(1)

		for _, chapter := range api_chapters {

			number := ""

			if Is_Oneshot(chapter.TitleName, chapter.ChapterSubTitle) {
				number = "1"
			} else if Is_Extra(chapter.TitleName) {
				number = fmt.Sprintf("%s.%d", lastNumber, lastSubNumber)
				lastSubNumber++
			} else {
				number_int, err := Chapter_Name_To_Int(chapter.TitleName)
				if err == nil {
					number = strconv.FormatInt(int64(number_int), 10)
					lastNumber = number
					lastSubNumber = 1
				} else {
					number_parts := strings.Split(chapter.TitleName, "-")
					if len(number_parts) == 2 {
						number_main, err := Chapter_Name_To_Int(number_parts[0])
						if err == nil {
							number_sub, err := Chapter_Name_To_Int(number_parts[1])
							if err == nil {
								number = fmt.Sprintf("%d.%d", number_main, number_sub)
								lastNumber = strconv.FormatInt(int64(number_main), 10)
								lastSubNumber = number_sub + 1
							} else {
								number = fmt.Sprintf("%s.%d", lastNumber, lastSubNumber)
								lastSubNumber++
							}
						} else {
							number = fmt.Sprintf("%s.%d", lastNumber, lastSubNumber)
							lastSubNumber++
						}
					} else {
						number = fmt.Sprintf("%s.%d", lastNumber, lastSubNumber)
						lastSubNumber++
					}
				}
			}

			chapters = append(chapters, &source.Chapter{
				Name:   chapter.ChapterSubTitle,
				Index:  i,
				Number: number,
				ID:     strconv.FormatInt(int64(chapter.ChapterId), 10),
				URL:    "https://mangaplus.shueisha.co.jp/viewer/" + strconv.FormatInt(int64(chapter.ChapterId), 10),
				Manga:  manga,
				Volume: "",
			})
			i++
		}

		_ = m.cache.chapters_app.Set(manga.ID, chapters)
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

		for _, chapter := range append(title_detail_view.ChapterListGroup[0].FirstChapterList, title_detail_view.ChapterListGroup[1].LastChapterList...) {

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
					number_parts := strings.Split(chapter.Name, "-")
					if len(number_parts) == 2 {
						number_main, err := Chapter_Name_To_Int(number_parts[0])
						if err == nil {
							number_sub, err := Chapter_Name_To_Int(number_parts[1])
							if err == nil {
								number = fmt.Sprintf("%d.%d", number_main, number_sub)
								lastNumber = strconv.FormatInt(int64(number_main), 10)
								lastSubNumber = number_sub + 1
							} else {
								number = fmt.Sprintf("%s.%d", lastNumber, lastSubNumber)
								lastSubNumber++
							}
						} else {
							number = fmt.Sprintf("%s.%d", lastNumber, lastSubNumber)
							lastSubNumber++
						}
					} else {
						number = fmt.Sprintf("%s.%d", lastNumber, lastSubNumber)
						lastSubNumber++
					}
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
			i++
		}

		_ = m.cache.chapters_web.Set(manga.ID, chapters)
	}

	return chapters, nil
}
