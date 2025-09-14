# Desafio: Client-Server-API (Pós Go Expert - Full Cycle)

Este projeto implementa o desafio de comunicação **Client ↔ Server** em Go,
com uso de `context` e timeouts.

---
## Estrutura
pos-go-expert/
client-server-api/
client.go
server.go
go.mod
go.sum
quotes.db # gerado pelo servidor
cotacao.txt # gerado pelo cliente
---

## Executando

### 1. Rodar o servidor
go run server.go

- Sobe na porta 8080
- Endpoint: http://localhost:8080/cotacao
- Persiste cotações no SQLite (quotes.db)
- Timeouts:
    - 200ms para chamada da API externa
    - 10ms para persistência no banco
- Logs exibem erro se os limites forem estourados.

### 2. Rodar o cliente
Em outro terminal: go run client.go

- Timeout de 300ms para resposta do servidor.
- Salva o arquivo cotacao.txt no formato:
    - Dólar: 5.3414

## Observações

- Arquivo cotacao.txt é gravado em UTF-8 sem BOM.
    - No Windows, para visualizar com acento correto:
        Get-Content .\cotacao.txt -Encoding utf8
    - Ou abra no VS Code.

- Banco de dados: quotes.db com tabela quotes:
    CREATE TABLE IF NOT EXISTS quotes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bid TEXT NOT NULL,
    fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
