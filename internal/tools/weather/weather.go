package weather

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

const (
	forecast = "https://api.open-meteo.com/v1/forecast"
	search   = "https://geocoding-api.open-meteo.com/v1/search"
)

type Weather struct {
	*resty.Client
}

type Response struct {
	Results []City `json:"results"`
}

type City struct {
	Name      string  `json:"name"`
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

type Data struct{}

func NewClient() *Weather {
	httpClient := resty.New()
	return &Weather{
		httpClient,
	}

}

func (w *Weather) Search(name string) (City, error) {
	resp, err := w.R().SetQueryParams(map[string]string{"name": name, "count": "1"}).Get(search)
	if err != nil {
		log.Error().Str("request", resp.Request.URL).Str("status", resp.Status()).Err(err).Msg("weather: problem search city")
		return City{}, err
	}
	var respObj Response
	err = json.Unmarshal(resp.Body(), &respObj)
	if err != nil {
		log.Error().Err(err).Str("request", resp.Request.URL).Str("status", resp.Status()).
			Msg("weather: problem unmarshal city search")
		return City{}, err
	}
	log.Debug().Str("request", resp.Request.URL).Str("status", resp.Status()).Any("data", respObj).
		Msg("weather: unmarshal search data")
	if len(respObj.Results) > 0 {
		return respObj.Results[0], nil
	}
	return City{}, fmt.Errorf("city %s not found", name)
}

func (w *Weather) Forecast(latitude float32, longitude float32, days int) (ForecastResponse, error) {
	params := map[string]string{
		"latitude":  fmt.Sprintf("%f", latitude),
		"longitude": fmt.Sprintf("%f", longitude),
		"current":   "temperature_2m,wind_speed_10m,rain,is_day",
	}
	if days == 1 {
		params["hourly"] = "temperature_2m,rain,wind_speed_10m"
	}
	if days > 1 {
		params["forecast_days"] = fmt.Sprintf("%d", days)
		params["daily"] = "temperature_2m_max,temperature_2m_min,rain_sum,wind_speed_10m_max"
	}
	resp, err := w.R().SetQueryParams(params).Get(forecast)
	if err != nil {
		log.Error().Err(err).Str("request", resp.Request.URL).Str("status", resp.Status()).
			Any("data", resp.Body()).Msg("weather: problem search city")
		return ForecastResponse{}, err
	}
	log.Debug().Str("request", resp.Request.URL).Str("status", resp.Status()).
		Any("data", resp.Body()).Msg("weather: forecast data")

	var forecast ForecastResponse
	err = json.Unmarshal(resp.Body(), &forecast)
	if err != nil {
		log.Error().Err(err).Str("request", resp.Request.URL).Str("status", resp.Status()).
			Msg("weather: problem unmarshal forecast")
		return ForecastResponse{}, err
	}

	return forecast, nil
}
