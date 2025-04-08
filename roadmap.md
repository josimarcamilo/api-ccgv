# V1
## Funcionalidades
- [x] se registrar no sistema
- [x] cadastrar usuários no time
## Financeiro
- [] cadastrar contas contábeis
- [] cadastrar categoriasaazxc
- [] criar lançamentos para as contas
## relatórios
- [] receitas por categoria
- [] despesas por categoria
- [] saldo das contas
- [] extrato de conta

# V2
# Usuário
- [x] cadastro de usuários no time

# Categorias
Categorias podem ser do tipo entrada ou saída.
Categorias podem ou nao serem usadas no mapa mensal.
- [] cadastro
- [x] editar categoria
- [] exluir categoria
- [] filtro por: tipo, uso_mapa

# Contas
- [] saldo inicial do banco ou caixa (verificar de qual conta foi paga) deve desconsiderar a décima do mes anterior
- [x] editar conta
- [x] calcular saldo das contas de acordo com uma data
- [x] escolher se a conta vai ser "A receber" (conta desconsiderada nas somas)
Contas para criar:
    - CCRD (taxas de lixo, IPTU) a receber
    - CCVI (taxas de lixo, IPTU) a receber
    - Lar dos Velhinhos a receber
    - CMGV a receber
    - Aplicacao financeira a resgatar

# Transações
- [x] colocar comprovante nas receitas como opicional
- [] excluir uma transacao
- [x] importar planilha
- [x] importar ofx
- [] escolher a conta para criar as transacoes ao importar os arquivos
- [] marcar transacao para nao considerar no mapa (exemplo é décima)
- [] marcar transacao como transferencia para nao contar como uma nova receita/despesa (exemplo saque do banco para caixa)
- [] criar padrao de descricoes
- [] configurar impressao
- [] filtrar por:
    - [] status
    - [] conta
    - [] descricao
    - [] periodo
    - [] tipo (entrada/saida)

# Tesoureiros e Conselho Fiscal
- [] recurso para aprovar/reprovar uma transacao (talvez na tela de visualizar/editar já resolva). Colocar observacao caso reprove
Acho que a transacao pode ter um status (criada, correcao, revisao, aprovada), salvar o histórico de mudança de status o usuário que mudou e a observacao.
De correcao só pode ir para revisao
listar os registros nessa mesma tela

# Relatórios
- [] Mapa Mensal (calculado)
- [] relatorio de despesas e receitas por categia de um periodo de data

# Cadatro de diretoria
- [] cadastro de membros diretoria (nome, telefone, email)
- [] cadastro membros do conselho fiscal (nome, telefone, email)

# Sitema
- [] colocar alerts de sucesso e falha
