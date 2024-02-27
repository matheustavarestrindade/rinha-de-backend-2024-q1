# README para Submissão da Rinha de Backend - 2024/Q1

## Sobre o Projeto

Este repositório contém a implementação da API para o desafio "Rinha de Backend - 2024/Q1", focando no controle de concorrência com operações de créditos e débitos (crébitos). Utilizei Golang para a criação das APIs, PostgreSQL para o banco de dados, e Nginx como load balancer para gerenciar o tráfego e promover alta disponibilidade.

## Tecnologias Utilizadas

- **Golang:** Linguagem de programação escolhida para o desenvolvimento das APIs, aproveitando sua performance e facilidade de uso para sistemas concorrentes.
- **PostgreSQL:** Banco de dados relacional para armazenar dados dos clientes e transações, escolhido por sua robustez e confiabilidade.
- **Nginx:** Utilizado como load balancer para distribuir as requisições de forma eficiente entre as instâncias da aplicação, garantindo balanceamento de carga e alta disponibilidade.
- **Docker:** Todos os componentes são conteinerizados, facilitando o deployment e a execução em diferentes ambientes.

## Funcionalidades

A API suporta as seguintes operações:

- **POST /clientes/[id]/transacoes:** Registra uma nova transação (crédito ou débito) para o cliente especificado.
- **GET /clientes/[id]/extrato:** Recupera o extrato das últimas transações do cliente, incluindo saldo total, limite e detalhes das transações.

## Regras de Negócio

- Transações de débito não podem exceder o limite disponível do cliente.
- Requisições que resultariam em saldo abaixo do limite devem retornar HTTP 422.
- IDs de clientes inexistentes devem retornar HTTP 404.

## Como Executar

Para executar a aplicação, você precisará do Docker e do Docker Compose instalados em sua máquina. Siga os passos abaixo:

1. Clone o repositório para sua máquina local.
2. Navegue até o diretório raiz do projeto.
3. Execute o comando `docker-compose up` para iniciar todos os serviços.

A API estará acessível na porta 9999 conforme configurado no Nginx.

## Estrutura do Repositório

- **docker-compose.yml:** Arquivo de configuração do Docker Compose para iniciar a aplicação, incluindo a API, banco de dados e Nginx.
- **nginx.conf:** Configuração do Nginx utilizada para balanceamento de carga.
- **/sql:** Diretório contendo scripts SQL para inicialização do banco de dados.
- **/src:** Código fonte da API desenvolvida em Golang.

## Contato

- **E-mail:** matheustavarestrindade@hotmail.com

Sinta-se livre para entrar em contato para dúvidas ou contribuições ao projeto. Agradecemos pelo interesse e boa sorte a todos os participantes da Rinha de Backend!
