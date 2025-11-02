package main

import (
	"encoding/json"
	"fmt"

	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

// Estruturas de dados para o JSON
type RamoAtividade struct {
	ID           string `json:"id"`
	Categoria    string `json:"categoria"`
	Subcategoria string `json:"subcategoria"`
}

type DadosFinanceiros struct {
	FaturamentoAnualBruto   float64 `json:"faturamento_anual_bruto"`
	CustosOperacionaisAnual float64 `json:"custos_operacionais_anual"`
	ImpostoTotalPagoAnual   float64 `json:"imposto_total_pago_anual"`
	MargemLucroLiquida      float64 `json:"margem_lucro_liquida"`
	AnoFiscal               int     `json:"ano_fiscal"`
}

type Coordenadas struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Localizacao struct {
	Endereco         string      `json:"endereco"`
	Cidade           string      `json:"cidade"`
	Estado           string      `json:"estado"`
	CEP              string      `json:"cep"`
	Coordenadas      Coordenadas `json:"coordenadas"`
	RegiaoGeografica string      `json:"regiao_geografica"`
}

type Comercio struct {
	IDComercio        int              `json:"id_comercio"`
	NomeFantasia      string           `json:"nome_fantasia"`
	RamoAtividade     RamoAtividade    `json:"ramo_atividade"`
	DadosFinanceiros  DadosFinanceiros `json:"dados_financeiros"`
	Localizacao       Localizacao      `json:"localizacao"`
	PorteEmpresa      string           `json:"porte_empresa"`
	DataAbertura      string           `json:"data_abertura"`
	StatusOperacional string           `json:"status_operacional"`
}

// Estrutura da resposta da API
type TopComerciosResponse struct {
	Top10Comercios     []Comercio `json:"top_10_comercios"`
	TempoProcessamento string     `json:"tempo_processamento"`
	FonteDados         string     `json:"fonte_dados"`
}

// Função para carregar e processar os dados
func getTop10Comercios() ([]Comercio, error) {
	filePath := "dados_comercios.json"
	log.Printf("Tentando ler o arquivo: %s", filePath)
	// Carregar os dados do JSON
	dados, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler o arquivo: %v", err)
	}

	log.Println("Arquivo lido com sucesso.")

	var comercios []Comercio
	if err := json.Unmarshal(dados, &comercios); err != nil {
		return nil, fmt.Errorf("erro ao decodificar o JSON: %v", err)
	}

	log.Println("JSON decodificado com sucesso.")

	// Ordenar por faturamento (do maior para o menor)
	sort.Slice(comercios, func(i, j int) bool {
		return comercios[i].DadosFinanceiros.FaturamentoAnualBruto > comercios[j].DadosFinanceiros.FaturamentoAnualBruto
	})

	// Retornar os 10 primeiros
	top10 := comercios[:10]

	return top10, nil
}

// Estrutura do Cache
type Cache struct {
	data       interface{}
	expiration time.Time
	mu         sync.Mutex
}

var top10FaturamentoCache Cache
var top10CidadesCache Cache

// Handler para o endpoint
func top10FaturamentoHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Endpoint /top10-faturamento chamado")
	inicio := time.Now()

	refresh := r.Header.Get("X-Cache-Refresh") == "true"

	if !refresh {
		top10FaturamentoCache.mu.Lock()
		// Verifica se o cache é válido
		if time.Now().Before(top10FaturamentoCache.expiration) {
			log.Println("Servindo do cache: /top10-faturamento")
			duracao := time.Since(inicio)
			resposta := top10FaturamentoCache.data.(TopComerciosResponse)
			resposta.TempoProcessamento = duracao.String()
			resposta.FonteDados = "Cache"
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resposta)
			top10FaturamentoCache.mu.Unlock()
			return
		}
		top10FaturamentoCache.mu.Unlock()
	}

	log.Println("Cache de /top10-faturamento inválido ou refresh solicitado, buscando dados...")
	top10, err := getTop10Comercios()
	if err != nil {
		log.Printf("Erro ao obter o top 10: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	duracao := time.Since(inicio)

	resposta := TopComerciosResponse{
		Top10Comercios:     top10,
		TempoProcessamento: duracao.String(),
		FonteDados:         "Processamento ao Vivo",
	}

	top10FaturamentoCache.mu.Lock()
	top10FaturamentoCache.data = resposta
	top10FaturamentoCache.expiration = time.Now().Add(1 * time.Minute)
	top10FaturamentoCache.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resposta); err != nil {
		log.Printf("Erro ao codificar a resposta em JSON: %v", err)
		http.Error(w, "Erro ao codificar a resposta em JSON", http.StatusInternalServerError)
	}
	log.Println("Resposta enviada com sucesso.")
}

type CidadeFaturamentoJSON struct {
	Cidade      string `json:"cidade"`
	Faturamento string `json:"faturamento"`
}

type CidadeFaturamento struct {
	Cidade      string
	Faturamento float64
}

type TopCidadesResponse struct {
	Top10Cidades       []CidadeFaturamentoJSON `json:"top_10_cidades"`
	TempoProcessamento string                  `json:"tempo_processamento"`
	FonteDados         string                  `json:"fonte_dados"`
}

func formatToBRL(value float64) string {
	p := message.NewPrinter(language.BrazilianPortuguese)
	return p.Sprintf("R$ %.2f", number.Decimal(value))
}

// Função para carregar e processar os dados
func getTop10Cidades() ([]CidadeFaturamento, error) {
	filePath := "dados_comercios.json"
	log.Printf("Tentando ler o arquivo: %s", filePath)
	// Carregar os dados do JSON
	dados, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler o arquivo: %v", err)
	}

	log.Println("Arquivo lido com sucesso.")

	var comercios []Comercio
	if err := json.Unmarshal(dados, &comercios); err != nil {
		return nil, fmt.Errorf("erro ao decodificar o JSON: %v", err)
	}

	log.Println("JSON decodificado com sucesso.")

	// Calcular faturamento por cidade
	faturamentoPorCidade := make(map[string]float64)
	for _, c := range comercios {
		faturamentoPorCidade[c.Localizacao.Cidade] += c.DadosFinanceiros.FaturamentoAnualBruto
	}

	// Converter mapa para slice
	var cidades []CidadeFaturamento
	for cidade, faturamento := range faturamentoPorCidade {
		cidades = append(cidades, CidadeFaturamento{Cidade: cidade, Faturamento: faturamento})
	}

	// Ordenar por faturamento (do maior para o menor)
	sort.Slice(cidades, func(i, j int) bool {
		return cidades[i].Faturamento > cidades[j].Faturamento
	})

	// Retornar as 10 primeiras
	top10 := cidades[:10]

	return top10, nil
}

// Handler para o endpoint
func top10CidadesHandler(w http.ResponseWriter, r *http.Request) {
	inicio := time.Now()
	refresh := r.Header.Get("X-Cache-Refresh") == "true"

	if !refresh {
		top10CidadesCache.mu.Lock()
		// Verifica se o cache é válido
		if time.Now().Before(top10CidadesCache.expiration) {
			log.Println("Servindo do cache: /top10-cidades")
			duracao := time.Since(inicio)
			resposta := top10CidadesCache.data.(TopCidadesResponse)
			resposta.TempoProcessamento = duracao.String()
			resposta.FonteDados = "Cache"
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resposta)
			top10CidadesCache.mu.Unlock()
			return
		}
		top10CidadesCache.mu.Unlock()
	}

	log.Println("Cache de /top10-cidades inválido ou refresh solicitado, buscando dados...")
	top10, err := getTop10Cidades()
	if err != nil {
		log.Printf("Erro ao obter o top 10 cidades: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Formatar faturamento para BRL
	var top10JSON []CidadeFaturamentoJSON
	for _, c := range top10 {
		top10JSON = append(top10JSON, CidadeFaturamentoJSON{
			Cidade:      c.Cidade,
			Faturamento: formatToBRL(c.Faturamento),
		})
	}

	duracao := time.Since(inicio)

	resposta := TopCidadesResponse{
		Top10Cidades:       top10JSON,
		TempoProcessamento: duracao.String(),
		FonteDados:         "Processamento ao Vivo",
	}

	top10CidadesCache.mu.Lock()
	top10CidadesCache.data = resposta
	top10CidadesCache.expiration = time.Now().Add(1 * time.Minute)
	top10CidadesCache.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resposta); err != nil {
		log.Printf("Erro ao codificar a resposta em JSON: %v", err)
		http.Error(w, "Erro ao codificar a resposta em JSON", http.StatusInternalServerError)
	}
	log.Println("Resposta enviada com sucesso.")
}

var top10CategoriasCache Cache

type CategoriaFaturamento struct {
	Categoria   string
	Faturamento float64
}

type CategoriaFaturamentoJSON struct {
	Categoria   string `json:"categoria"`
	Faturamento string `json:"faturamento"`
}

type TopCategoriasResponse struct {
	Top10Categorias    []CategoriaFaturamentoJSON `json:"top_10_categorias"`
	TempoProcessamento string                     `json:"tempo_processamento"`
	FonteDados         string                     `json:"fonte_dados"`
}

func getTop10Categorias() ([]CategoriaFaturamento, error) {
	filePath := "dados_comercios.json"
	log.Printf("Tentando ler o arquivo: %s", filePath)
	// Carregar os dados do JSON
	dados, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler o arquivo: %v", err)
	}

	log.Println("Arquivo lido com sucesso.")

	var comercios []Comercio
	if err := json.Unmarshal(dados, &comercios); err != nil {
		return nil, fmt.Errorf("erro ao decodificar o JSON: %v", err)
	}

	log.Println("JSON decodificado com sucesso.")

	// Calcular faturamento por categoria
	faturamentoPorCategoria := make(map[string]float64)
	for _, c := range comercios {
		faturamentoPorCategoria[c.RamoAtividade.Categoria] += c.DadosFinanceiros.FaturamentoAnualBruto
	}

	// Converter mapa para slice
	var categorias []CategoriaFaturamento
	for categoria, faturamento := range faturamentoPorCategoria {
		categorias = append(categorias, CategoriaFaturamento{Categoria: categoria, Faturamento: faturamento})
	}

	// Ordenar por faturamento (do maior para o menor)
	sort.Slice(categorias, func(i, j int) bool {
		return categorias[i].Faturamento > categorias[j].Faturamento
	})

	// Retornar as 10 primeiras, se houver
	if len(categorias) < 10 {
		return categorias, nil
	}

	top10 := categorias[:10]

	return top10, nil
}

func top10CategoriasHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Endpoint /top10-categorias chamado")
	inicio := time.Now()

	refresh := r.Header.Get("X-Cache-Refresh") == "true"

	if !refresh {
		top10CategoriasCache.mu.Lock()
		// Verifica se o cache é válido
		if time.Now().Before(top10CategoriasCache.expiration) {
			log.Println("Servindo do cache: /top10-categorias")
			duracao := time.Since(inicio)
			resposta := top10CategoriasCache.data.(TopCategoriasResponse)
			resposta.TempoProcessamento = duracao.String()
			resposta.FonteDados = "Cache"
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resposta)
			top10CategoriasCache.mu.Unlock()
			return
		}
		top10CategoriasCache.mu.Unlock()
	}

	log.Println("Cache de /top10-categorias inválido ou refresh solicitado, buscando dados...")
	top10, err := getTop10Categorias()
	if err != nil {
		log.Printf("Erro ao obter o top 10 categorias: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Formatar faturamento para BRL
	var top10JSON []CategoriaFaturamentoJSON
	for _, c := range top10 {
		top10JSON = append(top10JSON, CategoriaFaturamentoJSON{
			Categoria:   c.Categoria,
			Faturamento: formatToBRL(c.Faturamento),
		})
	}

	duracao := time.Since(inicio)

	resposta := TopCategoriasResponse{
		Top10Categorias:    top10JSON,
		TempoProcessamento: duracao.String(),
		FonteDados:         "Processamento ao Vivo",
	}

	top10CategoriasCache.mu.Lock()
	top10CategoriasCache.data = resposta
	top10CategoriasCache.expiration = time.Now().Add(1 * time.Minute)
	top10CategoriasCache.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resposta); err != nil {
		log.Printf("Erro ao codificar a resposta em JSON: %v", err)
		http.Error(w, "Erro ao codificar a resposta em JSON", http.StatusInternalServerError)
	}
	log.Println("Resposta enviada com sucesso.")
}

// Endpoint para gerar dados
type GerarDadosResponse struct {
	Mensagem           string `json:"mensagem"`
	TempoProcessamento string `json:"tempo_processamento"`
	RegistrosGerados   int    `json:"registros_gerados"`
}

func gerarComercioGo(idComercio int) Comercio {
	ramosAtividade := []RamoAtividade{
		{ID: "FOOD001", Categoria: "Alimentação", Subcategoria: "Padaria/Confeitaria"},
		{ID: "FOOD002", Categoria: "Alimentação", Subcategoria: "Restaurante/Lanchonete"},
		{ID: "RET001", Categoria: "Varejo", Subcategoria: "Loja de Roupas"},
		{ID: "RET002", Categoria: "Varejo", Subcategoria: "Eletrônicos"},
		{ID: "SERV001", Categoria: "Serviços", Subcategoria: "Consultoria"},
		{ID: "SERV002", Categoria: "Serviços", Subcategoria: "Beleza/Estética"},
		{ID: "AUTO001", Categoria: "Automotivo", Subcategoria: "Oficina Mecânica"},
		{ID: "HEAL001", Categoria: "Saúde", Subcategoria: "Farmácia"},
	}
	portesEmpresa := []string{"MEI", "Pequena", "Média", "Grande"}
	statusOperacional := []string{"Ativo", "Ativo", "Ativo", "Ativo", "Ativo", "Ativo", "Ativo", "Ativo", "Ativo", "Fechado"}

	ramo := ramosAtividade[rand.Intn(len(ramosAtividade))]
	porte := portesEmpresa[rand.Intn(len(portesEmpresa))]

	var faturamentoBase float64
	switch porte {
	case "MEI":
		faturamentoBase = 60000 + rand.Float64()*(81000-60000)
	case "Pequena":
		faturamentoBase = 200000 + rand.Float64()*(4800000-200000)
	case "Média":
		faturamentoBase = 5000000 + rand.Float64()*(35000000-5000000)
	default: // Grande
		faturamentoBase = 10000000 + rand.Float64()*(35000000-10000000)
	}

	faturamento := faturamentoBase
	custos := faturamento * (0.4 + rand.Float64()*(0.6-0.4))
	imposto := faturamento * (0.05 + rand.Float64()*(0.15-0.05))
	margemLucro := faturamento - custos - imposto

	latitude := -23.8 + rand.Float64()*(-22.5-(-23.8))
	longitude := -47.0 + rand.Float64()*(-43.0-(-47.0))

	return Comercio{
		IDComercio:    idComercio,
		NomeFantasia:  fmt.Sprintf("Comercio %d", idComercio),
		RamoAtividade: ramo,
		DadosFinanceiros: DadosFinanceiros{
			FaturamentoAnualBruto:   faturamento,
			CustosOperacionaisAnual: custos,
			ImpostoTotalPagoAnual:   imposto,
			MargemLucroLiquida:      margemLucro,
			AnoFiscal:               2024,
		},
		Localizacao: Localizacao{
			Endereco: fmt.Sprintf("Rua %d, %d", idComercio, rand.Intn(1000)),
			Cidade:   fmt.Sprintf("Cidade %d", rand.Intn(100)),
			Estado:   "SP",
			CEP:      fmt.Sprintf("%05d-000", rand.Intn(99999)),
			Coordenadas: Coordenadas{
				Latitude:  latitude,
				Longitude: longitude,
			},
			RegiaoGeografica: "Sudeste",
		},
		PorteEmpresa:      porte,
		DataAbertura:      time.Now().AddDate(-rand.Intn(10), -rand.Intn(12), -rand.Intn(28)).Format("2006-01-02"),
		StatusOperacional: statusOperacional[rand.Intn(len(statusOperacional))],
	}
}

func gerarDadosHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Endpoint /gerar-dados chamado")
	inicio := time.Now()

	// Carregar dados existentes
	filePath := "dados_comercios.json"
	var comercios []Comercio
	dadosAntigos, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		http.Error(w, "Erro ao ler arquivo de dados", http.StatusInternalServerError)
		return
	}

	if err == nil {
		if err := json.Unmarshal(dadosAntigos, &comercios); err != nil {
			http.Error(w, "Erro ao decodificar JSON existente", http.StatusInternalServerError)
			return
		}
	}

	// Gerar novos dados
	numRegistros := 10000 // Gerar 10k novos registros
	novosDados := make([]Comercio, numRegistros)
	for i := 0; i < numRegistros; i++ {
		// O ID deve continuar a partir do último existente
		novoID := len(comercios) + i + 1
		novosDados[i] = gerarComercioGo(novoID)
	}

	// Adicionar novos dados aos existentes
	comercios = append(comercios, novosDados...)

	// Salvar em JSON
	fileData, err := json.MarshalIndent(comercios, "", "    ")
	if err != nil {
		http.Error(w, "Erro ao gerar JSON", http.StatusInternalServerError)
		return
	}

	err = os.WriteFile(filePath, fileData, 0644)
	if err != nil {
		http.Error(w, "Erro ao salvar arquivo", http.StatusInternalServerError)
		return
	}

	duracao := time.Since(inicio)

	resposta := GerarDadosResponse{
		Mensagem:           fmt.Sprintf("%d novos registros adicionados com sucesso!", numRegistros),
		TempoProcessamento: duracao.String(),
		RegistrosGerados:   len(comercios),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resposta)
}

func gerarArquivoInicial() {
	filePath := "dados_comercios.json"
	log.Printf("Verificando se o arquivo de dados '%s' existe...", filePath)

	// Verifica se o arquivo já existe
	if _, err := os.Stat(filePath); err == nil {
		log.Println("Arquivo de dados já existe. Nenhuma ação necessária.")
		return
	} else if !os.IsNotExist(err) {
		log.Fatalf("Erro ao verificar o arquivo de dados: %v", err)
	}

	// Se não existe, gera um novo com 10.000 registros
	log.Println("Arquivo de dados não encontrado. Gerando um novo com 10.000 registros...")
	inicio := time.Now()

	numRegistros := 10000
	dados := make([]Comercio, numRegistros)
	for i := 0; i < numRegistros; i++ {
		dados[i] = gerarComercioGo(i + 1)
	}

	fileData, err := json.MarshalIndent(dados, "", "    ")
	if err != nil {
		log.Fatalf("Erro ao gerar JSON inicial: %v", err)
	}

	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		log.Fatalf("Erro ao salvar arquivo de dados inicial: %v", err)
	}

	duracao := time.Since(inicio)
	log.Printf("Arquivo de dados gerado com sucesso em %s.", duracao)
}

func main() {
	// Garante que o arquivo de dados existe
	gerarArquivoInicial()

	// Servir arquivos estáticos
	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Rota para a API
	http.HandleFunc("/top10-faturamento", top10FaturamentoHandler)
	http.HandleFunc("/top10-cidades", top10CidadesHandler)
	http.HandleFunc("/top10-categorias", top10CategoriasHandler)
	http.HandleFunc("/gerar-dados", gerarDadosHandler)

	// Rota para a página inicial
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/index.html")
	})

	fmt.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
