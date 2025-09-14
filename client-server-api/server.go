package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

const (
	apiURL             = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	httpListenAddr     = ":8080"
	apiTimeout         = 200 * time.Millisecond
	dbTimeout          = 10 * time.Millisecond
	dbFile             = "quotes.db"
	createTableDDL     = `CREATE TABLE IF NOT EXISTS quotes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT NOT NULL,
		fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
)

type awesomeResp struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

type outResp struct {
	Bid string `json:"bid"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatalf("erro ao abrir sqlite: %v", err)
	}
	if _, err := db.Exec(createTableDDL); err != nil {
		log.Fatalf("erro ao criar tabela: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", handleCotacao)

	s := &http.Server{
		Addr:         httpListenAddr,
		Handler:      mux,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("server ouvindo em %s", httpListenAddr)
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server erro: %v", err)
	}
}

func handleCotacao(w http.ResponseWriter, r *http.Request) {
	// 1) Chamada externa com timeout de 200ms
	ctxAPI, cancelAPI := context.WithTimeout(r.Context(), apiTimeout)
	defer cancelAPI()

	req, err := http.NewRequestWithContext(ctxAPI, http.MethodGet, apiURL, nil)
	if err != nil {
		log.Printf("[external] erro criando request: %v", err)
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(ctxAPI.Err(), context.DeadlineExceeded) {
			log.Printf("[external] timeout de %v (context deadline exceeded)", apiTimeout)
		} else {
			log.Printf("[external] erro na chamada: %v", err)
		}
		http.Error(w, "timeout ao consultar cotação", http.StatusGatewayTimeout)
		return
	}
	defer resp.Body.Close()

	var ar awesomeResp
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		log.Printf("[external] erro decodificando JSON: %v", err)
		http.Error(w, "resposta inválida da API", http.StatusBadGateway)
		return
	}

	bid := ar.USDBRL.Bid
	if bid == "" {
		log.Printf("[external] JSON sem campo bid")
		http.Error(w, "dados indisponíveis", http.StatusBadGateway)
		return
	}

	// 2) Persistência com timeout de 10ms (logar erro em caso de deadline)
	ctxDB, cancelDB := context.WithTimeout(r.Context(), dbTimeout)
	defer cancelDB()
	if _, err := db.ExecContext(ctxDB, "INSERT INTO quotes(bid) VALUES (?)", bid); err != nil {
		if errors.Is(ctxDB.Err(), context.DeadlineExceeded) {
			log.Printf("[db] timeout de %v ao inserir (context deadline exceeded)", dbTimeout)
		} else {
			log.Printf("[db] erro ao inserir: %v", err)
		}
		// Mesmo que a gravação falhe/timeout, seguimos respondendo o bid.
	}

	// 3) Responder apenas { "bid": "..." }
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(outResp{Bid: bid})
}
