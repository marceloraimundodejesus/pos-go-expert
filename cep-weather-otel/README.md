````markdown
# CEP → Weather com OTEL + Zipkin

Dois serviços em Go que implementam tracing distribuído:

- **service-a**: recebe `POST /` com `{ "cep": "########" }`, valida CEP e orquestra chamada ao **service-b**.
- **service-b**: recebe `GET /weather/{cep}`, resolve a cidade (ViaCEP) e consulta a temperatura (WeatherAPI), retornando C/F/K.

Observabilidade com **OpenTelemetry (OTLP gRPC → Collector)** e visualização no **Zipkin**.

---

## Requisitos

- Docker e Docker Compose
- Internet com acesso a `viacep.com.br` e `api.weatherapi.com`
- Go 1.23+ (apenas se for rodar/testar localmente fora do Docker)

> As imagens são **distroless** e compiladas com **Go 1.23** (multi-stage build).

---

## Subir o ambiente

```bash
docker compose up -d --build
docker compose ps
```
````

Serviços:

- **service-a**: [http://localhost:8080](http://localhost:8080)
- **service-b**: [http://localhost:8081](http://localhost:8081)
- **zipkin**: [http://localhost:9411](http://localhost:9411)

---

## Endpoints e exemplos

### Sucesso (CEP válido)

**Request**

```bash
# PowerShell
$body = '{"cep":"74366104"}'
iwr http://localhost:8080 -Method POST -ContentType 'application/json' -Body $body
```

**Response 200**

```json
{ "city": "Goiânia", "temp_C": 26, "temp_F": 78.8, "temp_K": 299 }
```

### CEP inválido (formato incorreto)

**Request**

```bash
$body = '{"cep":"123"}'
iwr http://localhost:8080 -Method POST -ContentType 'application/json' -Body $body -SkipHttpErrorCheck
```

**Response 422**

```
invalid zipcode
```

### CEP inexistente

**Request**

```bash
$body = '{"cep":"99999999"}'
iwr http://localhost:8080 -Method POST -ContentType 'application/json' -Body $body -SkipHttpErrorCheck
```

**Response 404**

```
not find zipcode
```

### Service B direto (debug)

**Request**

```bash
iwr http://localhost:8081/weather/74366104 -Method GET
```

---

## Observabilidade (Zipkin)

### Ver traces

1. Acesse **[http://localhost:9411](http://localhost:9411)**.
2. Aba **Find a trace** → adicione filtro `serviceName=service-a` → **RUN QUERY**.
3. Abra o trace mais recente (**SHOW**) e confira:

   - span **`service-a.root`** (server)
   - span filha **`service-a: http get`** chamando `service-b:8080/weather/{cep}`
   - no **service-b**: span **`service-b.server`** (server) e **duas spans cliente**:

     - chamada ao **ViaCEP** (URL `viacep.com.br`)
     - chamada à **WeatherAPI** (URL `api.weatherapi.com`)

### Ver grafo de dependências

1. Aba **Dependencies** → ajuste o período para incluir as últimas chamadas → **RUN QUERY**.
2. Deve aparecer a aresta **`service-a → service-b`**.

> Dica: se não aparecer, gere um novo request de sucesso no A e rode novamente, garantindo que o período inclui o horário atual.

---

## Decisões técnicas

- **Validação de CEP** no A: regex `^\d{8}$` (apenas números).

- **Mapeamento de respostas do B no A**:

  - `404` do B → retorna `404` com corpo **`not find zipcode`** (conforme enunciado).
  - Qualquer outro status ≠200 do B → `500` com **`internal error`**.
  - Falha de comunicação com B → `502` com **`service b unavailable`**.
  - CEP inválido → `422` com **`invalid zipcode`**.

- **Propagação OTEL** (A e B):
  Ambos os serviços configuram o propagador W3C:

  ```go
  otel.SetTextMapPropagator(
      propagation.NewCompositeTextMapPropagator(
          propagation.TraceContext{},
          propagation.Baggage{},
      ),
  )
  ```

  Isso garante **mesmo traceId** de ponta a ponta (A → B → provedores externos).

- **HTTP client** com `Timeout: 8s` e `otelhttp.NewTransport(...)` para instrumentação automática.

- **Build** multi-stage (Go 1.23) e **runtime distroless** com `ca-certificates` copiados para TLS.

---

## Variáveis de ambiente

### service-a

- `SERVICE_B_URL` (default: `http://service-b:8080`)
- `OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317`
- `OTEL_TRACES_SAMPLER=always_on` (via compose)
- `OTEL_RESOURCE_ATTRIBUTES=service.name=service-a` (via compose)

### service-b

- `WEATHER_API_KEY` (definida no `docker-compose.yml`)
- `OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317`
- `OTEL_TRACES_SAMPLER=always_on` (via compose)
- `OTEL_RESOURCE_ATTRIBUTES=service.name=service-b` (via compose)

> Dentro da rede do Compose, os serviços acessam o Collector como `otel-collector:4317`.

---

## Testes unitários (service-b)

```bash
cd service-b
go test ./internal/core -v
```

---

## Problemas comuns

- **500 no A** com erro OTEL no B apontando `api.weatherapi.com`: falha intermitente de rede/fornecedor. Tente novamente e confira a span do B no Zipkin (tags `http.url`, `otel.status_code`, `otel.status_description`).
- **Dependencies vazio**: ajuste o período e confirme a propagação em **ambos** os serviços.
- **`docker exec` falha no A**: imagem distroless (não tem `sh`). Use `docker inspect` para ver envs.

---

## Subir / Parar

```bash
# subir / rebuild
docker compose up -d --build

# rebuild apenas de um serviço
docker compose up -d --build --no-deps service-a
docker compose up -d --build --no-deps service-b

# desligar
docker compose down
```

---

## Estrutura (resumo)

```
.
├── service-a
│   ├── cmd/server/main.go        # valida CEP, chama service-b, OTEL + propagator
│   └── Dockerfile                # Go 1.23 build → distroless runtime
├── service-b
│   ├── cmd/server/main.go        # orquestra ViaCEP + WeatherAPI, OTEL + propagator
│   ├── internal/                 # core, cep, weather, http handler + testes
│   └── Dockerfile                # Go 1.23 build → distroless runtime
├── otel-collector/collector.yaml # pipeline OTLP → Zipkin
├── docker-compose.yml            # services, envs e rede
└── README.md
```

---

## Observações finais

- Todos os requisitos do desafio foram atendidos:

  - Validação e mensagens padronizadas (`invalid zipcode`, `not find zipcode`);
  - OTEL + Zipkin com tracing distribuído A ↔ B;
  - Spans de cliente para ViaCEP e WeatherAPI;
  - Docker/Compose prontos para testes.

- Prints de referência: Find a trace (A e B), spans detalhadas e Dependencies (aresta A → B).

```

```
