document.addEventListener("DOMContentLoaded", () => {
    const tableBody = document.getElementById("accounts-table");

    // Função para carregar contas
    async function loadAccounts() {
        try {
            const response = await fetch("/accounts", {
                credentials: "include", // Inclui os cookies na requisição
            });

            if (!response.ok) {
                tableBody.innerHTML = `<tr><td colspan="3">Erro ao carregar contas</td></tr>`;
                return;
            }

            const accounts = await response.json();

            // Limpar tabela antes de adicionar novas linhas
            tableBody.innerHTML = "";

            if (accounts.length === 0) {
                tableBody.innerHTML = `<tr><td colspan="3">Nenhuma conta encontrada</td></tr>`;
                return;
            }

            // Preencher tabela com contas
            accounts.forEach((account) => {
                const row = document.createElement("tr");

                row.innerHTML = `
                    <td>${account.ID}</td>
                    <td>${account.name}</td>
                    <td>${account.balance.toFixed(2)}</td>
                `;

                tableBody.appendChild(row);
            });
        } catch (error) {
            tableBody.innerHTML = `<tr><td colspan="3">Erro: ${error.message}</td></tr>`;
        }
    }

    // Carregar contas ao carregar a página
    loadAccounts();
});
