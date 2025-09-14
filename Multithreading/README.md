# Desafio: Multithreading

Neste desafio, serão feitas **duas requests em paralelo** para buscar o endereço a partir de um CEP, aceitando a **resposta mais rápida** e descartando a mais lenta.

## Requisitos (escopo)
- Duas APIs chamadas simultaneamente:
  - `https://brasilapi.com.br/api/cep/v1/{CEP}`
  - `http://viacep.com.br/ws/{CEP}/json/`
- Limite de **1 segundo** para a resposta (timeout).
- Exibir no **console** o endereço retornado e **qual API** respondeu mais rápido.

## Status
*Planejamento inicial*
