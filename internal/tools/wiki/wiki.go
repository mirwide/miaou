package wiki

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

type SearchResponse struct {
	Pages []Page
}
type Page struct {
	Key string
}

func GetWiki(key string, lang string) string {

	r := resty.New()
	search, err := r.R().Get(fmt.Sprintf("https://%s.wikipedia.org/w/rest.php/v1/search/page?q=%s&limit=1", lang, url.QueryEscape(key)))
	log.Debug().Str("request", search.Request.URL).Str("status", search.Status()).Any("data", search.Body()).Msg("tools: wiki search data")
	if err != nil {
		log.Error().Str("request", search.Request.URL).Str("status", search.Status()).Err(err).Msg("tools: problem search wiki")
		return "При запросе информации возникла ошибка"
	}

	var searchObj SearchResponse
	err = json.Unmarshal(search.Body(), &searchObj)
	if err != nil || len(searchObj.Pages) == 0 {
		log.Error().Err(err).Str("request", search.Request.URL).Str("status", search.Status()).
			Msg("weather: problem unmarshal city search")
		return "Информация не найдена"
	}

	resp, err := r.R().Get(fmt.Sprintf("https://%s.wikipedia.org/w/rest.php/v1/page/%s", lang, searchObj.Pages[0].Key))
	log.Debug().Str("request", resp.Request.URL).Str("status", resp.Status()).Any("data", resp.Body()).Msg("tools: wiki data")
	if err != nil {
		log.Error().Str("request", resp.Request.URL).Str("status", resp.Status()).Err(err).Msg("tools: problem get wiki")
		return "При запросе информации возникла ошибка"
	}
	return string(resp.Body())
}
