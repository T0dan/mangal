package source

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/metafates/mangal/constant"
	"github.com/metafates/mangal/filesystem"
	"github.com/metafates/mangal/key"
	"github.com/metafates/mangal/style"
	"github.com/metafates/mangal/util"
	"github.com/samber/mo"
	"github.com/spf13/viper"
)

// Chapter is a struct that represents a chapter of a manga.
type Chapter struct {
	// Name of the chapter
	Name string `json:"name" jsonschema:"description=Name of the chapter"`
	// URL of the chapter
	URL string `json:"url" jsonschema:"description=URL of the chapter"`
	// Index of the chapter in the manga.
	Index uint16 `json:"index" jsonschema:"description=Index of the chapter in the manga"`
	// Number of the chapter in the manga.
	Number string `json:"number" jsonschema:"description=Number of the chapter in the manga"`
	// ID of the chapter in the source.
	ID string `json:"id" jsonschema:"description=ID of the chapter in the source"`
	// Volume which the chapter belongs to.
	Volume string `json:"volume" jsonschema:"description=Volume which the chapter belongs to"`
	// Manga that the chapter belongs to.
	Manga *Manga `json:"-"`
	// Pages of the chapter.
	Pages []*Page `json:"pages" jsonschema:"description=Pages of the chapter"`

	isDownloaded mo.Option[bool]
	size         uint64
}

func (c *Chapter) String() string {
	return c.Name
}

func (c *Chapter) CleanPages() {
	for _, page := range c.Pages {
		if page != nil {
			page.Contents = nil
		}
	}
}

// DownloadPages downloads the Pages contents of the Chapter.
// Pages needs to be set before calling this function.
func (c *Chapter) DownloadPages(temp bool, progress func(string)) (err error) {
	c.size = 0
	status := func() string {
		return fmt.Sprintf(
			"Downloading %s %s",
			util.Quantify(len(c.Pages), "page", "pages"),
			style.Faint(c.SizeHuman()),
		)
	}

	progress(status())
	wg := sync.WaitGroup{}
	wg.Add(len(c.Pages))

	for _, page := range c.Pages {
		if page == nil {
			return fmt.Errorf("page #%d is empty, aborting download", page.Index)
		}

		d := func(page *Page) {
			defer wg.Done()

			// if at any point, an error is encountered, stop downloading other pages
			if err != nil {
				return
			}

			err = page.Download()
			c.size += page.Size
			progress(status())
		}

		if viper.GetBool(key.DownloaderAsync) {
			go d(page)
		} else {
			d(page)
		}
	}

	wg.Wait()

	if err != nil {
		c.isDownloaded = mo.Some(false)
		return err
	}

	c.isDownloaded = mo.Some(!temp)
	return
}

// formattedName of the chapter according to the template in the config.
func (c *Chapter) formattedName() (name string) {
	name = viper.GetString(key.DownloaderChapterNameTemplate)

	var sourceName string
	if c.Source() != nil {
		sourceName = c.Source().Name()
	}

	number := int(c.Index)
	numberExt := ""

	re := regexp.MustCompile(`(\d+)(\.?\d*)`)
	re_match := re.FindStringSubmatch(c.Number)
	if re_match != nil {
		number_buf, err := strconv.Atoi(re_match[1])
		if err == nil {
			number = number_buf
			numberExt = re_match[2]
		}
	}

	for variable, value := range map[string]string{
		"manga":          c.Manga.Name,
		"chapter":        c.Name,
		"index":          fmt.Sprintf("%d", c.Index),
		"padded-index":   fmt.Sprintf("%04d", c.Index),
		"number":         fmt.Sprintf("%d%s", number, numberExt),
		"padded-number":  fmt.Sprintf("%04d%s", number, numberExt),
		"chapters-count": fmt.Sprintf("%d", len(c.Manga.Chapters)),
		"volume":         c.Volume,
		"source":         sourceName,
	} {
		name = strings.ReplaceAll(name, fmt.Sprintf("{%s}", variable), value)
	}

	return
}

// SizeHuman is the same as Size but returns a human-readable string.
func (c *Chapter) SizeHuman() string {
	return humanize.Bytes(c.size)
}

func (c *Chapter) Filename() (filename string) {
	filename = util.SanitizeFilenameWows(c.formattedName())

	if viper.GetBool(key.DownloaderVolSafeFilename) {
		filename = util.VolSafeFileName(filename)
	}

	// plain format assumes that chapter is a directory with images
	// rather than a single file. So no need to add extension to it
	if f := viper.GetString(key.FormatsUse); f != constant.FormatPlain {
		return filename + "." + f
	}

	return
}

func (c *Chapter) IsDownloaded() bool {
	if c.isDownloaded.IsPresent() {
		return c.isDownloaded.MustGet()
	}

	path, _ := c.path(c.Manga.peekPath(), false)
	exists, _ := filesystem.Api().Exists(path)
	c.isDownloaded = mo.Some(exists)
	return exists
}

func (c *Chapter) path(relativeTo string, createVolumeDir bool) (path string, err error) {
	if createVolumeDir {
		path = filepath.Join(path, util.SanitizeFilenameWows(c.Volume))
		err = filesystem.Api().MkdirAll(path, os.ModePerm)
		if err != nil {
			return
		}
	}

	path = filepath.Join(relativeTo, c.Filename())
	return
}

func (c *Chapter) Path(temp bool) (path string, err error) {
	var manga string
	manga, err = c.Manga.Path(temp)
	if err != nil {
		return
	}

	return c.path(manga, c.Volume != "" && viper.GetBool(key.DownloaderCreateVolumeDir))
}

func (c *Chapter) Source() Source {
	return c.Manga.Source
}

func (c *Chapter) ComicInfo() *ComicInfo {
	var (
		day, month, year int
	)

	if viper.GetBool(key.MetadataComicInfoXMLAddDate) {
		if viper.GetBool(key.MetadataComicInfoXMLAlternativeDate) {
			// get current date
			t := time.Now()
			day = t.Day()
			month = int(t.Month())
			year = t.Year()
		} else {
			day = c.Manga.Metadata.StartDate.Day
			month = c.Manga.Metadata.StartDate.Month
			year = c.Manga.Metadata.StartDate.Year
		}
	} // empty dates will be omitted

	chapter_num := strconv.FormatUint(uint64(c.Index), 10)
	if c.Number != "" {
		chapter_num = c.Number
	}

	return &ComicInfo{
		XmlnsXsd: "http://www.w3.org/2001/XMLSchema",
		XmlnsXsi: "http://www.w3.org/2001/XMLSchema-instance",

		Title:       c.Name,
		Series:      c.Manga.Name,
		Number:      chapter_num,
		Web:         c.URL,
		Genre:       strings.Join(c.Manga.Metadata.Genres, ","),
		PageCount:   len(c.Pages),
		Summary:     c.Manga.Metadata.Summary,
		Count:       c.Manga.Metadata.Chapters,
		Characters:  strings.Join(c.Manga.Metadata.Characters, ","),
		Year:        year,
		Month:       month,
		Day:         day,
		Writer:      strings.Join(c.Manga.Metadata.Staff.Story, ","),
		Penciller:   strings.Join(c.Manga.Metadata.Staff.Art, ","),
		Letterer:    strings.Join(c.Manga.Metadata.Staff.Lettering, ","),
		Translator:  strings.Join(c.Manga.Metadata.Staff.Translation, ","),
		Tags:        strings.Join(c.Manga.Metadata.Tags, ","),
		Notes:       "Downloaded with Mangal. https://github.com/metafates/mangal",
		LanguageISO: c.Manga.Metadata.LanguageISO,
		Manga:       "YesAndRightToLeft",
		OrigTitle:   c.Name,
		OrigIndex:   int(c.Index),
	}
}
