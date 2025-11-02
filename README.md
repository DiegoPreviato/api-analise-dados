# API de Análise de Dados de Comércios

Esta é uma API em Go que realiza a análise de dados de comércios. A API expõe endpoints para obter o top 10 de comércios por faturamento, o top 10 de cidades por faturamento, e o top 10 de categorias por faturamento. A API também possui um endpoint para gerar dados de teste.

## Como executar

1. Navegue até o diretório `api-go`: `cd api-go`
2. Execute o comando: `go run main.go`
3. O servidor estará rodando em `http://localhost:8080`

## Endpoints da API

* `GET /top10-faturamento`: Retorna o top 10 comércios com maior faturamento.
* `GET /top10-cidades`: Retorna o top 10 cidades com maior faturamento.
* `GET /top10-categorias`: Retorna o top 10 categorias com maior faturamento.
* `POST /gerar-dados`: Gera 10.000 novos registros de comércios.
