package core

import (
	"context"
	"errors"
	"math"
)

var (
	ErrInvalidCEP   = errors.New("invalid zipcode")
	ErrNotFoundCEP  = errors.New("can not find zipcode")
	ErrWeatherFetch = errors.New("weather fetch error")
)

type ViaCEPClient interface {
	Lookup(ctx context.Context, cep string) (city, uf string, found bool, err error)
}

type WeatherClient interface {
	CurrentTempC(ctx context.Context, city, uf string) (float64, error)
}

type Service struct {
	ViaCEP  ViaCEPClient
	Weather WeatherClient
}

func (s Service) GetWeatherByCEP(ctx context.Context, cep string) (WeatherResult, error) {
	if !IsValidCEP(cep) {
		return WeatherResult{}, ErrInvalidCEP
	}
	city, uf, found, err := s.ViaCEP.Lookup(ctx, cep)
	if err != nil {
		return WeatherResult{}, err
	}
	if !found {
		return WeatherResult{}, ErrNotFoundCEP
	}

	c, err := s.Weather.CurrentTempC(ctx, city, uf)
	if err != nil {
		return WeatherResult{}, ErrWeatherFetch
	}
	f := c*1.8 + 32
	k := c + 273

	c = round1(c)
	f = round1(f)
	k = round1(k)

	return WeatherResult{TempC: c, TempF: f, TempK: k}, nil
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}
