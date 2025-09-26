package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type WeatherAPI struct {
	Client *http.Client
	Key    string
}

func NewWeatherAPI() WeatherAPI {
	return WeatherAPI{
		Client: &http.Client{Timeout: time.Second},
		Key:    os.Getenv("WEATHERAPI_KEY"),
	}
}

type currentResp struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

func (w WeatherAPI) CurrentTempC(ctx context.Context, city, uf string) (float64, error) {
	q := url.QueryEscape(fmt.Sprintf("%s,%s,BR", city, uf))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no",
			w.Key, q), nil)
	if err != nil {
		return 0, err
	}
	res, err := w.Client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	var cr currentResp
	if err := json.NewDecoder(res.Body).Decode(&cr); err != nil {
		return 0, err
	}
	return cr.Current.TempC, nil
}
