package mangaplus

import (
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/metafates/mangal/constant"
	"github.com/metafates/mangal/log"
	"github.com/metafates/mangal/network"
	"github.com/metafates/mangal/provider/mangaplus/mangaplus_resp_web"
	"github.com/metafates/mangal/source"
	"github.com/metafates/mangal/util"
	"google.golang.org/protobuf/proto"
)

const (
	Name = "MangaPlus"
	ID   = Name + " built-in"
)

type Mangaplus struct {
	webapi_url  string
	appapi_url  string
	use_app_api bool
	cache       struct {
		mangas_web   *cacher[[]*source.Manga]
		mangas_app   *cacher[[]*source.Manga]
		chapters_web *cacher[[]*source.Chapter]
		chapters_app *cacher[[]*source.Chapter]
	}
}

func (m *Mangaplus) GetWebViewer(chapter_id string) (*mangaplus_resp_web.MangaViewer, error) {
	req, err := http.NewRequest(http.MethodGet, m.webapi_url+"/manga_viewer", nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	//req.Header.Set("Referer", "https://mangaplus.shueisha.co.jp/manga_list/all")
	req.Header.Set("User-Agent", constant.UserAgent)

	params := req.URL.Query()

	params.Add("chapter_id", strings.TrimSpace(chapter_id))
	params.Add("split", "no")
	params.Add("img_quality", "super_high")
	//params.Add("format", "json")

	req.URL.RawQuery = params.Encode()

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

	resp_data := &mangaplus_resp_web.Response{}

	if err := proto.Unmarshal(buf, resp_data); err != nil {
		err = errors.New("couldn't unmarschal response")
		log.Error(err)
		return nil, err
	}

	if resp_data.Success == nil {
		err = errors.New("unsuccessfull request")
		log.Error(err)
		return nil, err
	}

	return resp_data.Success.MangaViewer, nil
}

func (m *Mangaplus) GetWebTitleDetails(title_id string) (*mangaplus_resp_web.TitleDetailView, error) {
	req, err := http.NewRequest(http.MethodGet, m.webapi_url+"/title_detail", nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	//req.Header.Set("Referer", "https://mangaplus.shueisha.co.jp/manga_list/all")
	req.Header.Set("User-Agent", constant.UserAgent)

	params := req.URL.Query()

	params.Add("title_id", strings.TrimSpace(title_id))
	//params.Add("format", "json")

	req.URL.RawQuery = params.Encode()

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

	resp_data := &mangaplus_resp_web.Response{}

	if err := proto.Unmarshal(buf, resp_data); err != nil {
		err = errors.New("couldn't unmarschal response")
		log.Error(err)
		return nil, err
	}

	if resp_data.Success == nil {
		err = errors.New("unsuccessfull request")
		log.Error(err)
		return nil, err
	}

	return resp_data.Success.TitleDetailView, nil
}

func Chapter_Name_To_Int(chapter_name string) (int, error) {
	return strconv.Atoi(strings.TrimLeft(chapter_name, "#"))
}

func Is_Oneshot(chapter_name string, chapter_subtitle string) bool {
	chapter_name = strings.ToLower(chapter_name)
	chapter_subtitle = strings.ToLower(chapter_subtitle)
	if (strings.Contains(chapter_name, "one") && strings.Contains(chapter_name, "shot")) ||
		(strings.Contains(chapter_subtitle, "one") && strings.Contains(chapter_subtitle, "shot")) {
		return true
	} else {
		return false
	}
}

func Is_Extra(chapter_name string) bool {
	return strings.Trim(chapter_name, "#") == "ex"
}

func Escape_Path(path string) string {
	var re = regexp.MustCompile(`[^\w]+`)
	path = re.ReplaceAllString(path, " ")
	path = strings.Trim(path, "!\"#$%&'()*+, -./:;<=>?@[\\]^_`{|}~ ")
	return path
}

func (*Mangaplus) Name() string {
	return Name
}

func (*Mangaplus) ID() string {
	return ID
}

func New() *Mangaplus {
	mp := &Mangaplus{}

	mp.webapi_url = "https://jumpg-webapi.tokyo-cdn.com/api"
	mp.use_app_api = false
	mp.cache.mangas_web = newCacher[[]*source.Manga](ID + "_mangas_web")
	mp.cache.mangas_app = newCacher[[]*source.Manga](ID + "_mangas_app")
	mp.cache.chapters_web = newCacher[[]*source.Chapter](ID + "_chapters_web")
	mp.cache.chapters_app = newCacher[[]*source.Chapter](ID + "_chapters_app")

	return mp
}
