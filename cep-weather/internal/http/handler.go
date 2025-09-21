package httpx

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"cep-weather/internal/core"
)

type Handler struct {
	Svc core.Service
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := splitPath(r.URL.Path)
	if len(parts) != 2 || parts[0] != "weather" {
		http.NotFound(w, r)
		return
	}
	cep := parts[1]

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	res, err := h.Svc.GetWeatherByCEP(ctx, cep)
	switch err {
	case nil:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	case core.ErrInvalidCEP:
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
	case core.ErrNotFoundCEP:
		http.Error(w, "can not find zipcode", http.StatusNotFound)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func splitPath(p string) []string {
	for len(p) > 0 && p[0] == '/' {
		p = p[1:]
	}
	if p == "" {
		return nil
	}
	var out []string
	cur := ""
	for i := 0; i < len(p); i++ {
		if p[i] == '/' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(p[i])
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
