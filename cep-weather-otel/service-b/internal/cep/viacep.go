package cep

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"service-b/internal/core"
)

type ViaCEP struct {
	Client *http.Client
}

func NewViaCEP() ViaCEP {
	return ViaCEP{Client: &http.Client{Timeout: time.Second}}
}

func (v ViaCEP) Lookup(ctx context.Context, cep string) (string, string, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep), nil)
	if err != nil {
		return "", "", false, err
	}
	res, err := v.Client.Do(req)
	if err != nil {
		return "", "", false, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		io.Copy(io.Discard, res.Body)
		return "", "", false, nil
	}

	var out core.ViaCEPInfo
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", "", false, nil
	}

	if out.Erro || out.Localidade == "" || out.UF == "" {
		return "", "", false, nil
	}
	return out.Localidade, out.UF, true, nil
}
