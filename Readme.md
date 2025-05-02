# infraestrutura heroku
Subir alterações direto
```bash
git push heroku main
```
# Categorias
- Décima CP N. Sra Aparecida Ilha
- Décima CP N. Sra Lourdes
- Duocentésima Lar dos Velhinhos
- Empréstimo

# Movimentacao Interna
É uma movimentacao em uma unidade e nao altera o saldo das contas financeiras
## Estrutura da tabela
| Conta | Unidade | Tipo | descricao | Valor |
| ----------- | ----- | ------- | -------------- | ------- |
|             | conf. SS  | Entrada   | saldo inicial | 500.00  |
|             | conf. x  | Saida | correcao | 500.00  |

# Transferencia Interna
É a transferencia entre unidades e nao altera o saldo das contas financeiras
## Estrutura da tabela
| Conta  | Unidade | Tipo | descricao | Valor |
| ----------- | ----- | ------- | -------------- | ------- |
| Banco Itau  | CCGV  | Entrada | aluguel chonim | 5000.00 |
|             | CCGV  | Saida   | aluguel chonim | 500.00  |
|             | conf  | Entrada | aluguel chonim | 500.00  |

# Transferencia entre contas financeiras
Altera o saldo nas contas financeiras
## Estrutura da tabela
| Conta  | Unidade | Tipo | descricao | Valor |
| ----------- | ----- | ------- | -------------- | ------- |
| Banco Itau  | CCGV  | Saída | aplicacao CDB | 5000.00 |
| Aplicacao  | CCGV  | Entrada | aplicacao CDB | 5000.00 |

# Emprestimos
- vincular transacao a um emprestimo
- no emprestimo listar todas as transcacoes com aquele emprestimo_id
- toda vez que eu vincular uma transacao a um emprestimo eu calculo o valor pago, se o valor pago for >= valor do emprestimo altero o status dele para quitado

## Estrutura da tabela emprestimos
| Unidade Credora  | Unidade Devedora | Valor | Valor Pago | Status | Observacao |
| ----------- | ----- | ------- | -------------- | ------- | ------- |
| CCGV | Lar dos Velhinhos  | 10.000,00 | 1.000,00 | perdoado | para pagar funcionarios. Perdoado dia x ata numero y
### Possiveis status
| Status do emprestimo  |
| ----------- |
| ativo |
| quitado |
| perdoado |
## Estrutura da tabela transacao
| Conta  | Unidade | Tipo | descricao | Valor | Emprestimo_id
| ----------- | ----- | ------- | -------------- | ------- | ------- |
| Banco Itau  | CCGV  | Entrada | pix recebido | 5000.00 | 34