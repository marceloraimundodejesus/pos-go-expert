package core

import (
	"context"
	"testing"
	"time"
)

type fakeVia struct {
	city string
	uf   string
	ok   bool
	err  error
}

func (f fakeVia) Lookup(ctx context.Context, cep string) (string, string, bool, error) {
	return f.city, f.uf, f.ok, f.err
}

type fakeW struct {
	c   float64
	err error
}

func (f fakeW) CurrentTempC(ctx context.Context, city, uf string) (float64, error) {
	return f.c, f.err
}

func TestIsValidCEP(t *testing.T) {
	if IsValidCEP("1234567") || IsValidCEP("123456789") || IsValidCEP("abcd1234") {
		t.Fatal("should be invalid")
	}
	if !IsValidCEP("12345678") {
		t.Fatal("should be valid")
	}
}

func TestService_InvalidCEP(t *testing.T) {
	s := Service{}
	_, err := s.GetWeatherByCEP(context.Background(), "123")
	if err != ErrInvalidCEP {
		t.Fatalf("expected ErrInvalidCEP, got %v", err)
	}
}

func TestService_NotFound(t *testing.T) {
	s := Service{
		ViaCEP:  fakeVia{ok: false},
		Weather: fakeW{c: 25},
	}
	_, err := s.GetWeatherByCEP(context.Background(), "12345678")
	if err != ErrNotFoundCEP {
		t.Fatalf("expected ErrNotFoundCEP, got %v", err)
	}
}

func TestService_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	s := Service{
		ViaCEP:  fakeVia{city: "Goiania", uf: "GO", ok: true},
		Weather: fakeW{c: 25.0},
	}
	res, err := s.GetWeatherByCEP(ctx, "74000000")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.TempC != 25.0 || res.TempF != 77.0 || res.TempK != 298.0 {
		t.Fatalf("unexpected values: %+v", res)
	}
}
