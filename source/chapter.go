package source

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/metafates/mangal/constant"
	"github.com/metafates/mangal/filesystem"
	"github.com/metafates/mangal/log"
	"github.com/metafates/mangal/util"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Chapter is a struct that represents a chapter of a manga.
type Chapter struct {
	// Name of the chapter
	Name string
	// URL of the chapter
	URL string
	// Index of the chapter in the manga.
	Index uint16
	// ID of the chapter in the source.
	ID string
	// Volume which the chapter belongs to.
	Volume string
	// Manga that the chapter belongs to.
	Manga *Manga `json:"-"`
	// Pages of the chapter.
	Pages []*Page
}

func (c *Chapter) String() string {
	return c.Name
}

// DownloadPages downloads the Pages contents of the Chapter.
// Pages needs to be set before calling this function.
func (c *Chapter) DownloadPages() error {
	log.Debugf("Downloading %d pages of chapter \"%s\"", len(c.Pages), c.Name)
	wg := sync.WaitGroup{}
	wg.Add(len(c.Pages))

	var err error
	for _, page := range c.Pages {
		d := func(page *Page) {
			defer wg.Done()

			// if at any point, an error is encountered, stop downloading other pages
			if err != nil {
				return
			}

			err = page.Download()
		}

		if viper.GetBool(constant.DownloaderAsync) {
			go d(page)
		} else {
			d(page)
		}
	}

	wg.Wait()
	return err
}

// formattedName of the chapter according to the template in the config.
func (c *Chapter) formattedName() (name string) {
	name = viper.GetString(constant.DownloaderChapterNameTemplate)

	for variable, value := range map[string]string{
		"manga":          c.Manga.Name,
		"chapter":        c.Name,
		"index":          fmt.Sprintf("%d", c.Index),
		"padded-index":   fmt.Sprintf("%04d", c.Index),
		"chapters-count": fmt.Sprintf("%d", len(c.Manga.Chapters)),
		"volume":         c.Volume,
		"source":         c.Source().Name(),
	} {
		name = strings.ReplaceAll(name, fmt.Sprintf("{%s}", variable), value)
	}

	return
}

// Size of the chapter in bytes.
func (c *Chapter) Size() uint64 {
	var n uint64

	for _, page := range c.Pages {
		n += page.Size
	}

	return n
}

// SizeHuman is the same as Size but returns a human-readable string.
func (c *Chapter) SizeHuman() string {
	if size := c.Size(); size == 0 {
		return "Unknown size"
	} else {
		return humanize.Bytes(size)
	}
}

func (c *Chapter) Filename() (filename string) {
	filename = util.SanitizeFilename(c.formattedName())

	// plain format assumes that chapter is a directory with images
	// rather than a single file. So no need to add extension to it
	if f := viper.GetString(constant.FormatsUse); f != constant.Plain {
		return filename + "." + f
	}

	return
}

func (c *Chapter) Path(temp bool) (path string, err error) {
	path, err = c.Manga.Path(temp)
	if err != nil {
		return
	}

	if c.Volume != "" && viper.GetBool(constant.DownloaderCreateVolumeDir) {
		path = filepath.Join(path, util.SanitizeFilename(c.Volume))
		err = filesystem.Get().MkdirAll(path, os.ModePerm)
		if err != nil {
			return
		}
	}

	path = filepath.Join(path, c.Filename())
	return
}

func (c *Chapter) Source() Source {
	return c.Manga.Source
}
