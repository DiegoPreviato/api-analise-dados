document.addEventListener('DOMContentLoaded', () => {
    const top10Link = document.getElementById('top10-faturamento');
    const dashboard = document.getElementById('dashboard');

    top10Link.addEventListener('click', (event) => {
        event.preventDefault();
        loadTop10Faturamento();
    });

    // Carrega o dashboard principal por padrão
    loadTop10Faturamento();
});

async function loadTop10Faturamento() {
    const dashboard = document.getElementById('dashboard');
    dashboard.innerHTML = '<h2>Carregando...</h2>';

    try {
        const response = await fetch('/top10');
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

    if (!data.top_10_comercios || data.top_10_comercios.length === 0) {
        dashboard.innerHTML = '<h2>Nenhum dado encontrado.</h2>';
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
        <p class="processing-time">Tempo de processamento: ${data.tempo_processamento}</p>
    `;

    dashboard.innerHTML = `<h2>Top 10 Empresas por Faturamento</h2>${table}`;
}
