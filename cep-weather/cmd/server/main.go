package main

import (
	"log"
	"net/http"
	"os"

	"cep-weather/internal/cep"
	"cep-weather/internal/core"
	httpx "cep-weather/internal/http"
	"cep-weather/internal/weather"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	via := cep.NewViaCEP()
	wapi := weather.NewWeatherAPI()

	h := httpx.Handler{
		Svc: core.Service{
			ViaCEP:  via,
			Weather: wapi,
		},
	}

	log.Printf("listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, h); err != nil {
		log.Fatal(err)
	}
}
