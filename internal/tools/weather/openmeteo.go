package weather

type ForecastResponse struct {
	Latitude                 float32      `json:"latitude"`
	Longitude                float32      `json:"longitude"`
	GenerationStrategytimeMs float32      `json:"generationtime_ms"`
	UTCOffsetSeconds         int32        `json:"utc_offset_seconds"`
	Timezone                 string       `json:"timezone"`
	TimezoneAbbrev           string       `json:"timezone_abbreviation"`
	Elevation                float32      `json:"elevation"`
	CurrentUnits             CurrentUnits `json:"current_units"`
	Current                  CurrentData  `json:"current"`
}

type CurrentUnits struct {
	Time          string `json:"time"`
	Interval      string `json:"interval"`
	Temperature2m string `json:"temperature_2m"`
	WindSpeed10m  string `json:"wind_speed_10m"`
	Rain          string `json:"rain"`
	IsDay         string `json:"is_day"`
}

type CurrentData struct {
	Time          string  `json:"time"`
	Interval      int32   `json:"interval"`
	Temperature2m float32 `json:"temperature_2m"`
	WindSpeed10m  float32 `json:"wind_speed_10m"`
	Rain          float32 `json:"rain"`
	IsDay         int32   `json:"is_day"`
}
