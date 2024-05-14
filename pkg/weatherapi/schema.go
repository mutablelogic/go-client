package weatherapi

import (
	"encoding/json"
	"strconv"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Location struct {
	Name           string  `json:"name"`
	Region         string  `json:"region"`
	Country        string  `json:"country"`
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
	Timezone       string  `json:"timezone"`
	LocaltimeEpoch int64   `json:"localtime_epoch"`
	Localtime      Time    `json:"localtime,omitempty"`
}

type Current struct {
	LastUpdatedEpoch int64   `json:"last_updated_epoch"`
	LastUpdated      Time    `json:"last_updated,omitempty"`
	TempC            float64 `json:"temp_c"`
	TempF            float64 `json:"temp_f"`
	IsDay            int     `json:"is_day"` // Whether to show day condition icon (1) or night icon (0)
	Condition        struct {
		Text string `json:"text"`
		Icon string `json:"icon"`
		Code int    `json:"code"`
	} `json:"condition"`
	WindMph    float64 `json:"wind_mph"`
	WindKph    float64 `json:"wind_kph"`
	WindDegree int     `json:"wind_degree"`
	WindDir    string  `json:"wind_dir"`
	PressureMb float64 `json:"pressure_mb"`
	PressureIn float64 `json:"pressure_in"`
	PrecipMm   float64 `json:"precip_mm"`
	PrecipIn   float64 `json:"precip_in"`
	Humidity   int     `json:"humidity"`
	Cloud      int     `json:"cloud"`
	FeelslikeC float64 `json:"feelslike_c"`
	FeelslikeF float64 `json:"feelslike_f"`
	VisKm      float64 `json:"vis_km"`
	VisMiles   float64 `json:"vis_miles"`
	Uv         float64 `json:"uv"`
	GustMph    float64 `json:"gust_mph"`
	GustKph    float64 `json:"gust_kph"`
}

type Weather struct {
	Id       int       `json:"custom_id,omitempty"`
	Query    string    `json:"q,omitempty"`
	Location *Location `json:"location,omitempty"`
	Current  *Current  `json:"current,omitempty"`
}

type Time struct {
	time.Time
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (w Weather) String() string {
	data, _ := json.MarshalIndent(w, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// MARSHAL TIME

func (t *Time) UnmarshalJSON(data []byte) error {
	if unquoted, err := strconv.Unquote(string(data)); err != nil {
		return err
	} else if v, err := time.ParseInLocation("2006-01-02 15:04", unquoted, time.Local); err != nil {
		return err
	} else {
		t.Time = v
	}
	return nil
}
