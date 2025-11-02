let currentView = 'top10'; // 'top10' or 'cidades'

document.addEventListener('DOMContentLoaded', () => {
    const top10Link = document.getElementById('top10-faturamento');
    const top10CidadesLink = document.getElementById('top10-cidades');
    const top10CategoriasLink = document.getElementById('top10-categorias');
    const gerarDadosLink = document.getElementById('gerar-dados');
    const refreshButton = document.getElementById('refresh-button');

    top10Link.addEventListener('click', (event) => {
        event.preventDefault();
        currentView = 'top10';
        loadTop10Faturamento();
    });

    top10CidadesLink.addEventListener('click', (event) => {
        event.preventDefault();
        currentView = 'cidades';
        loadTop10Cidades();
    });

    top10CategoriasLink.addEventListener('click', (event) => {
        event.preventDefault();
        currentView = 'categorias';
        loadTop10Categorias();
    });

    gerarDadosLink.addEventListener('click', (event) => {
        event.preventDefault();
        loadGerarDados();
    });

    refreshButton.addEventListener('click', () => {
        if (currentView === 'top10') {
            loadTop10Faturamento(true);
        } else if (currentView === 'cidades') {
            loadTop10Cidades(true);
        } else if (currentView === 'categorias') {
            loadTop10Categorias(true);
        }
    });

    // Carrega o dashboard principal por padrão
    loadTop10Faturamento();
});

async function loadTop10Faturamento(refresh = false) {
    const dashboard = document.getElementById('dashboard');
    dashboard.innerHTML = '<h2>Carregando...</h2>';

    try {
        const url = refresh ? '/top10-faturamento?refresh=true' : '/top10-faturamento';
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error(`Erro na API: ${response.statusText}`);
        }
        const data = await response.json();

        renderTop10Table(data);
    } catch (error) {
        dashboard.innerHTML = `<p style="color: red;">Erro ao carregar os dados: ${error.message}</p>`;
    }
}

function renderTop10Table(data) {
    const dashboard = document.getElementById('dashboard');
    const refreshButton = document.getElementById('refresh-button');

    if (!data.top_10_comercios || data.top_10_comercios.length === 0) {
        dashboard.innerHTML = '<h2>Nenhum dado encontrado.</h2>';
        refreshButton.style.display = 'none';
        return;
    }

    const table = `
        <table>
            <thead>
                <tr>
                    <th>ID</th>
                    <th>Nome Fantasia</th>
                    <th>Faturamento Anual (R$)</th>
                    <th>Cidade</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
                ${data.top_10_comercios.map(comercio => `
                    <tr>
                        <td>${comercio.id_comercio}</td>
                        <td>${comercio.nome_fantasia}</td>
                        <td>${comercio.dados_financeiros.faturamento_anual_bruto.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}</td>
                        <td>${comercio.localizacao.cidade}</td>
                        <td>${comercio.status_operacional}</td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
        <p class="processing-time">Fonte: ${data.fonte_dados} | Tempo de processamento: ${data.tempo_processamento}</p>
    `;

    dashboard.innerHTML = `<h2>Top 10 Empresas por Faturamento</h2>${table}`;
    refreshButton.style.display = 'block';
}

async function loadTop10Cidades(refresh = false) {
    const dashboard = document.getElementById('dashboard');
    dashboard.innerHTML = '<h2>Carregando...</h2>';

    try {
        const url = refresh ? '/top10-cidades?refresh=true' : '/top10-cidades';
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error(`Erro na API: ${response.statusText}`);
        }
        const data = await response.json();

        renderTop10CidadesTable(data);
    } catch (error) {
        dashboard.innerHTML = `<p style="color: red;">Erro ao carregar os dados: ${error.message}</p>`;
    }
}

function renderTop10CidadesTable(data) {
    const dashboard = document.getElementById('dashboard');
    const refreshButton = document.getElementById('refresh-button');

    if (!data.top_10_cidades || data.top_10_cidades.length === 0) {
        dashboard.innerHTML = '<h2>Nenhum dado encontrado.</h2>';
        refreshButton.style.display = 'none';
        return;
    }

    const table = `
        <table>
            <thead>
                <tr>
                    <th>Cidade</th>
                    <th>Faturamento Total (R$)</th>
                </tr>
            </thead>
            <tbody>
                ${data.top_10_cidades.map(cidade => `
                    <tr>
                        <td>${cidade.cidade}</td>
                        <td>${cidade.faturamento.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}</td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
        <p class="processing-time">Fonte: ${data.fonte_dados} | Tempo de processamento: ${data.tempo_processamento}</p>
    `;

    dashboard.innerHTML = `<h2>Top 10 Cidades por Faturamento</h2>${table}`;
    refreshButton.style.display = 'block';
}

async function loadGerarDados() {
    const dashboard = document.getElementById('dashboard');
    dashboard.innerHTML = '<h2>Gerando novos dados... Por favor, aguarde.</h2>';
    const refreshButton = document.getElementById('refresh-button');
    refreshButton.style.display = 'none';

    try {
        const response = await fetch('/gerar-dados');
        if (!response.ok) {
            throw new Error(`Erro na API: ${response.statusText}`);
        }
        const data = await response.json();

        dashboard.innerHTML = `
            <h2>Dados Gerados com Sucesso!</h2>
            <p>${data.mensagem}</p>
            <p><strong>Registros Gerados:</strong> ${data.registros_gerados}</p>
            <p><strong>Tempo de Processamento:</strong> ${data.tempo_processamento}</p>
        `;
    } catch (error) {
        dashboard.innerHTML = `<p style="color: red;">Erro ao gerar os dados: ${error.message}</p>`;
    }
}

async function loadTop10Categorias(refresh = false) {
    const dashboard = document.getElementById('dashboard');
    dashboard.innerHTML = '<h2>Carregando...</h2>';

    try {
        const url = refresh ? '/top10-categorias?refresh=true' : '/top10-categorias';
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error(`Erro na API: ${response.statusText}`);
        }
        const data = await response.json();

        renderTop10CategoriasTable(data);
    } catch (error) {
        dashboard.innerHTML = `<p style="color: red;">Erro ao carregar os dados: ${error.message}</p>`;
    }
}

function renderTop10CategoriasTable(data) {
    const dashboard = document.getElementById('dashboard');
    const refreshButton = document.getElementById('refresh-button');

    if (!data.top_10_categorias || data.top_10_categorias.length === 0) {
        dashboard.innerHTML = '<h2>Nenhum dado encontrado.</h2>';
        refreshButton.style.display = 'none';
        return;
    }

    const table = `
        <table>
            <thead>
                <tr>
                    <th>Categoria</th>
                    <th>Faturamento Total (R$)</th>
                </tr>
            </thead>
            <tbody>
                ${data.top_10_categorias.map(categoria => `
                    <tr>
                        <td>${categoria.categoria}</td>
                        <td>${categoria.faturamento.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}</td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
        <p class="processing-time">Fonte: ${data.fonte_dados} | Tempo de processamento: ${data.tempo_processamento}</p>
    `;

    dashboard.innerHTML = `<h2>Top 10 Categorias por Faturamento</h2>${table}`;
    refreshButton.style.display = 'block';
}
