package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	serverURL   = "http://localhost:8080/cotacao"
	clientLimit = 300 * time.Millisecond
	outFile     = "cotacao.txt"
)

type serverResp struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), clientLimit)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
	if err != nil {
		log.Fatalf("[client] erro criando request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Fatalf("[client] timeout de %v (context deadline exceeded)", clientLimit)
		}
		log.Fatalf("[client] erro na chamada: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("[client] status inesperado: %s", resp.Status)
	}

	var sr serverResp
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		log.Fatalf("[client] erro decodificando JSON: %v", err)
	}
	if sr.Bid == "" {
		log.Fatalf("[client] resposta sem bid")
	}

	content := fmt.Sprintf("Dólar: %s\n", sr.Bid)
	if err := os.WriteFile(outFile, []byte(content), 0644); err != nil {
		log.Fatalf("[client] erro salvando arquivo: %v", err)
	}

	log.Printf("[client] salvo em %s -> %s", outFile, content)
}
