package tools

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mirwide/miaou/internal/tools/weather"
	"github.com/ollama/ollama/api"
	"github.com/rs/zerolog/log"
)

type Parameters struct {
	Type       string   `json:"type"`
	Defs       any      `json:"$defs,omitempty"`
	Items      any      `json:"items,omitempty"`
	Required   []string `json:"required"`
	Properties map[string]struct {
		Type        api.PropertyType `json:"type"`
		Items       any              `json:"items,omitempty"`
		Description string           `json:"description"`
		Enum        []any            `json:"enum,omitempty"`
	} `json:"properties"`
}

type Properties struct {
	Type        api.PropertyType `json:"type"`
	Items       any              `json:"items,omitempty"`
	Description string           `json:"description"`
	Enum        []any            `json:"enum,omitempty"`
}

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

func NewProperties(props map[string]Properties) map[string]struct {
	Type        api.PropertyType `json:"type"`
	Items       any              `json:"items,omitempty"`
	Description string           `json:"description"`
	Enum        []any            `json:"enum,omitempty"`
} {
	var result = make(map[string]struct {
		Type        api.PropertyType `json:"type"`
		Items       any              `json:"items,omitempty"`
		Description string           `json:"description"`
		Enum        []any            `json:"enum,omitempty"`
	})
	for name, prop := range props {

		result[name] = struct {
			Type        api.PropertyType `json:"type"`
			Items       any              `json:"items,omitempty"`
			Description string           `json:"description"`
			Enum        []any            `json:"enum,omitempty"`
		}{
			Type:        prop.Type,
			Items:       prop.Items,
			Description: prop.Description,
			Enum:        prop.Enum,
		}
	}
	return result
}
