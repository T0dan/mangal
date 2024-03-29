package mangaplus

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/metafates/mangal/constant"
	"github.com/metafates/mangal/key"
	"github.com/metafates/mangal/log"
	"github.com/metafates/mangal/network"
	"github.com/metafates/mangal/provider/mangaplus/mangaplus_resp_app"
	"github.com/metafates/mangal/provider/mangaplus/mangaplus_resp_web"
	"github.com/metafates/mangal/source"
	"github.com/metafates/mangal/util"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
)

const (
	Name    = "MangaPlus"
	ID      = Name + " built-in"
	StdLang = "en"
)

type Mangaplus struct {
	webapi_url     string
	appapi_url     string
	appapi_secret  string
	appapi_os      string
	appapi_os_ver  string
	appapi_app_ver string
	use_app_api    bool
	cache          struct {
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

func (m *Mangaplus) Generate_App_Secret() error {
	android_id_buf := make([]byte, 8)
	rand.Read(android_id_buf)
	android_id := hex.EncodeToString(android_id_buf)
	defice_token_buf := md5.Sum([]byte(android_id))
	device_token := hex.EncodeToString(defice_token_buf[:])
	security_key_buf := md5.Sum([]byte(device_token + "4Kin9vGg"))
	security_key := hex.EncodeToString(security_key_buf[:])

	req, err := http.NewRequest(http.MethodPut, m.appapi_url+"/register", nil)
	if err != nil {
		log.Error(err)
		return err
	}

	req.Header.Set("accept", "*/*")
	req.Header.Set("User-Agent", "okhttp/4.9.0")

	params := req.URL.Query()

	params.Add("os", m.appapi_os)
	params.Add("os_ver", m.appapi_os_ver)
	params.Add("app_ver", m.appapi_app_ver)

	params.Add("device_token", device_token)
	params.Add("security_key", security_key)

	req.URL.RawQuery = params.Encode()

	resp, err := network.Client.Do(req)
	if err != nil {
		log.Error(err)
		return err
	}

	defer util.Ignore(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		err = errors.New("http error: " + resp.Status)
		log.Error(err)
		return err
	}

	buf, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	resp_data := &mangaplus_resp_app.Response{}

	if err := proto.Unmarshal(buf, resp_data); err != nil {
		err = errors.New("couldn't unmarschal response")
		log.Error(err)
		return err
	}

	if resp_data.Success == nil {
		if resp_data.Error != nil {
			err = errors.New("Api Error: " + resp_data.Error.Default.Code)
			log.Error(err)
			return err
		} else {
			err = errors.New("unsuccessfull request")
			log.Error(err)
			return err
		}
	}

	m.appapi_secret = resp_data.Success.RegisterView.Secret

	viper.Set(key.MangaplusAppApiToken, resp_data.Success.RegisterView.Secret)
	switch err := viper.WriteConfig(); err.(type) {
	case viper.ConfigFileNotFoundError:
		return viper.SafeWriteConfig()
	default:
		return err
	}
}

func (m *Mangaplus) GetAppViewer(chapter_id string) (*mangaplus_resp_app.MangaViewer, error) {
	req, err := http.NewRequest(http.MethodGet, m.appapi_url+"/manga_viewer", nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	req.Header.Set("accept", "*/*")
	req.Header.Set("User-Agent", "okhttp/4.9.0")

	params := req.URL.Query()

	params.Add("os", m.appapi_os)
	params.Add("os_ver", m.appapi_os_ver)
	params.Add("app_ver", m.appapi_app_ver)
	params.Add("secret", m.appapi_secret)

	params.Add("chapter_id", strings.TrimSpace(chapter_id))
	params.Add("split", "no")
	params.Add("img_quality", "super_high")

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

	resp_data := &mangaplus_resp_app.Response{}

	if err := proto.Unmarshal(buf, resp_data); err != nil {
		err = errors.New("couldn't unmarschal response")
		log.Error(err)
		return nil, err
	}

	if resp_data.Success == nil {
		if resp_data.Error != nil {
			err = errors.New("Api Error: " + resp_data.Error.Default.Code)
			log.Error(err)
			return nil, err
		} else {
			err = errors.New("unsuccessfull request")
			log.Error(err)
			return nil, err
		}
	}

	return resp_data.Success.MangaViewer, nil
}

func (m *Mangaplus) GetAppTitleDetails(title_id string) (*mangaplus_resp_app.TitleDetailView, error) {
	req, err := http.NewRequest(http.MethodGet, m.appapi_url+"/title_detailV2", nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	req.Header.Set("accept", "*/*")
	req.Header.Set("User-Agent", "okhttp/4.9.0")

	params := req.URL.Query()

	params.Add("os", m.appapi_os)
	params.Add("os_ver", m.appapi_os_ver)
	params.Add("app_ver", m.appapi_app_ver)
	params.Add("secret", m.appapi_secret)

	params.Add("title_id", title_id)

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

	resp_data := &mangaplus_resp_app.Response{}

	if err := proto.Unmarshal(buf, resp_data); err != nil {
		err = errors.New("couldn't unmarschal response")
		log.Error(err)
		return nil, err
	}

	if resp_data.Success == nil {
		if resp_data.Error != nil {
			err = errors.New("Api Error: " + resp_data.Error.Default.Code)
			log.Error(err)
			return nil, err
		} else {
			err = errors.New("unsuccessfull request")
			log.Error(err)
			return nil, err
		}
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

func (*Mangaplus) StdLang() string {
	return StdLang
}

func New() *Mangaplus {
	mp := &Mangaplus{}

	mp.webapi_url = "https://jumpg-webapi.tokyo-cdn.com/api"
	mp.appapi_url = "https://jumpg-api.tokyo-cdn.com/api"
	mp.use_app_api = viper.GetBool(key.MangaplusUseAppApi)
	mp.appapi_secret = viper.GetString(key.MangaplusAppApiToken)
	mp.appapi_os = viper.GetString(key.MangaplusAppApiOs)
	mp.appapi_os_ver = viper.GetString(key.MangaplusAppApiOsVer)
	mp.appapi_app_ver = viper.GetString(key.MangaplusAppApiVer)
	mp.cache.mangas_web = newCacher[[]*source.Manga](ID + "_mangas_web")
	mp.cache.mangas_app = newCacher[[]*source.Manga](ID + "_mangas_app")
	mp.cache.chapters_web = newCacher[[]*source.Chapter](ID + "_chapters_web")
	mp.cache.chapters_app = newCacher[[]*source.Chapter](ID + "_chapters_app")

	return mp
}
