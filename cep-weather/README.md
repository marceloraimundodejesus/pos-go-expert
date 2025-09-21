# CEP Weather (Go + Cloud Run)

**Objetivo:** Recebe um CEP (8 dígitos), identifica a cidade via **ViaCEP** e retorna o clima atual nas unidades **Celsius**, **Fahrenheit** e **Kelvin**.  
**Produção:** Cloud Run — URL: https://cep-weather-699154500419.southamerica-east1.run.app

# Requisitos atendidos

- CEP válido (8 dígitos) → senão: **422** `invalid zipcode`
- CEP não encontrado no ViaCEP → **404** `can not find zipcode`
- Sucesso → **200**:
  ```json
  { "temp_C": 28.5, "temp_F": 83.3, "temp_K": 301.5 }
  ```

# Testes Automatizados - cobre: "CEP inválido (422)", "CEP não encontrado (404)" e "sucesso (200 com conversões C/F/K)"

go test ./... -v

# Endpoints de Teste

## 200

curl -i "https://cep-weather-699154500419.southamerica-east1.run.app/weather/74366104"

## 422 (formato inválido)

curl -i "https://cep-weather-699154500419.southamerica-east1.run.app/weather/74366-104"

## 404 (não encontrado)

curl -i "https://cep-weather-699154500419.southamerica-east1.run.app/weather/00000000"
