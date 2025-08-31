package tools

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mirwide/miaou/internal/tools/weather"
	"github.com/rs/zerolog/log"
)

func GetTime() string {
	now := time.Now()
	return now.Format(time.RFC3339)
}

func GetWeather(city string, days int) string {

	type Response struct {
		City     weather.City             `json:"city"`
		Forecast weather.ForecastResponse `json:"forecast"`
	}

	cl := weather.NewClient()
	ct, err := cl.Search(city)
	if err != nil {
		log.Error().Err(err).Msg("tools: problem city search")
		return fmt.Sprintf("Город %s не найден", city)
	}
	resp, err := cl.Forecast(ct.Latitude, ct.Longitude, days)
	if err != nil {
		log.Error().Err(err).Msg("tools: problem forecast weather")
		return fmt.Sprintf("Данных о прогнозе погоды для %s нет", city)
	}

	r := Response{City: ct, Forecast: resp}
	rs, err := json.Marshal(r)
	if err != nil {
		log.Error().Err(err).Msg("tools: error forecast serialization")
		return fmt.Sprintf("Данных о прогнозе погоды для %s нет", city)
	}

	return string(rs)
}
