package tools

import (
	"fmt"
	"time"

	"github.com/mirwide/miaou/internal/tools/weather"
	"github.com/rs/zerolog/log"
)

type Parameters struct {
	Type       string   `json:"type"`
	Required   []string `json:"required"`
	Properties map[string]struct {
		Type        string   `json:"type"`
		Description string   `json:"description"`
		Enum        []string `json:"enum,omitempty"`
	} `json:"properties"`
}

type Properties struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

func GetTime() string {
	now := time.Now()
	return now.Format(time.RFC3339)
}

func GetWeather(city string, days int) string {

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
	return string(resp)
}

func NewProperties(props map[string]Properties) map[string]struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
} {
	var result = make(map[string]struct {
		Type        string   `json:"type"`
		Description string   `json:"description"`
		Enum        []string `json:"enum,omitempty"`
	})
	for name, prop := range props {

		result[name] = struct {
			Type        string   `json:"type"`
			Description string   `json:"description"`
			Enum        []string `json:"enum,omitempty"`
		}{
			Type:        prop.Type,
			Description: prop.Description,
			Enum:        prop.Enum,
		}
	}
	return result
}
