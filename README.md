# Pós Go Expert - Full Cycle

Este repositório contém os desafios desenvolvidos durante a pós-graduação **Go Expert** da [Full Cycle](https://fullcycle.com.br/).

---

## Estrutura client-server-api

- [`client-server-api/`](./client-server-api)  
  **Desafio - Client_Server-API**: comunicação Client ↔ Server em Go.

  - `server.go`: servidor HTTP na porta 8080, endpoint `/cotacao`.
  - `client.go`: cliente HTTP que consulta o servidor e salva a cotação em arquivo.
  - Uso de `context` para controle de timeouts.
  - Persistência das cotações em SQLite.

- [`Multithreading/`](./Multithreading)  
  **Desafio - Multithreading**: uso de **multithreading** e concorrência em Go.

  - Consultar duas APIs de CEP em paralelo.
  - Responder apenas a mais rápida.
  - Timeout máximo de **1s**.

- [`Clean Architecture/`](./Clean%20Architecture)  
  **Desafio - Clean Architecture**: implementação seguindo **Clean Architecture**.
  - Listagem de orders via **REST (GET /order)**, **gRPC** e **GraphQL**.
  - Banco de dados com migrações.
  - Dockerfile e docker-compose para subir a stack.
  - Arquivo `api.http` com requests de exemplo.

---

## Como navegar

Cada pasta de desafio contém:

- Código fonte em Go (`.go`)
- Arquivo `README.md` com instruções específicas
- `.gitignore` para ignorar artefatos gerados em runtime

---

## CEP Weather (Go + Cloud Run)

- **Código:** [/cep-weather](./cep-weather)
- **Produção:** https://cep-weather-699154500419.southamerica-east1.run.app

**Status esperados (produção)**

```bash
# 200
curl -i "https://cep-weather-699154500419.southamerica-east1.run.app/weather/74366104"

# 422 (formato inválido)
curl -i "https://cep-weather-699154500419.southamerica-east1.run.app/weather/74366-104"

# 404 (não encontrado)
curl -i "https://cep-weather-699154500419.southamerica-east1.run.app/weather/00000000"
```

## Testes automatizados - Cobre:

- **CEP inválido (422)**
- **CEP não encontrado (404)**
- **sucesso (200 com conversões C/F/K)**

go test ./... -v
