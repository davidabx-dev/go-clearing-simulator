<div align="center">

# 🏦 Go Clearing Simulator (Financial System)

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go)
![AWS](https://img.shields.io/badge/AWS-SQS-FF9900?style=for-the-badge&logo=amazon-aws)
![Docker](https://img.shields.io/badge/Docker-LocalStack-2496ED?style=for-the-badge&logo=docker)
![Architecture](https://img.shields.io/badge/Clean-Architecture-green?style=for-the-badge)

</div>

> **Simulador de Sistema de Compensação Bancária de Alta Performance**
> Focado em processamento assíncrono, resiliência e garantia de idempotência.

## 🎯 O Desafio
Em sistemas financeiros (Clearing), o processamento de arquivos de remessa (como CNAB/BASE II) exige:
1.  **Alta Performance:** Ler arquivos gigantes sem estourar a memória RAM.
2.  **Resiliência:** Garantir que nenhuma transação seja perdida.
3.  **Idempotência:** Garantir que **pagamentos duplicados sejam bloqueados**, mesmo se o sistema falhar ou receber o arquivo 2x.

## 🏗️ Arquitetura (Event-Driven)

O projeto segue **Clean Architecture** e é dividido em microsserviços desacoplados:

```flowchart LR
flowchart LR
    A[Arquivo Legacy .txt] -->|Upload Stream| B(API Gateway)
    B -->|Parser O(1) Mem| C{Validação}
    C -->|JSON| D[AWS SQS]
    D -->|Pull| E(Worker Service)
    E -->|Check| F[(Idempotency Store)]
    E -->|Process| G[Finalização]
```
<br>

**Componentes:**

- **cmd/api:** Recebe o upload e faz o streaming do arquivo (não carrega tudo na memória).
- **internal/parser:** Lê arquivos posicionais (formato fixo) linha a linha.
- **cmd/worker:** Consumidor concorrente que processa a fila SQS
- **Idempotency Layer:** Implementação Thread-Safe (Mutex) que impede processamento duplicado.

---

## Como Rodar Localmente

Pré-requisitos: `Go 1.22+`, `Docker` e `Docker Compose

**1. Subir a Infraestrutura (AWS LocalStack)**

Simulamos o SQS localmente para não gerar custos.

```Bash
docker-compose up -d
```

---

**2. Iniciar a API (Producer)**
Em um terminal:

```Bash
go run cmd/api/main.go
# Output: 🚀 Servidor rodando na porta 8080...
```

---

**3. Iniciar o Worker (Consumer)**
Em outro terminal:

```Bash
go run cmd/worker/main.go
# Output: 👷 Worker Iniciando... (Com Idempotência e Mutex)
```

---

**4. Enviar um Arquivo de Teste**
>Simule o envio de uma remessa bancária:

```Bash
# Windows (PowerShell)
curl.exe -F "file=@remessa_teste.txt" http://localhost:8080/upload

# Linux/Mac
curl -F "file=@remessa_teste.txt" http://localhost:8080/upload
```

---

## 🛡️ Teste de Idempotência (Prova de Fogo)

O sistema é protegido contra falhas de rede que enviam o mesmo arquivo duas vezes.

1. Envie o arquivo `remessa_teste.txt`.

2.  Envie **novamente** o mesmo arquivo logo em seguida.

**Resultado no Log do Worker:**

```Plaintext
✅ Sucesso! Transação 550e8400... finalizada.
...
🛑 DUPLICIDADE: Transação 550e8400... já foi processada. Ignorando.
```
>O sistema detecta o ID duplicado e descarta a mensagem sem processar o pagamento novamente.

---

## 🛠️ Tech Stack & Decisões Técnicas

| Tecnologia | Motivo da Escolha |
|------------|------------|
| **Golang** |  Concorrência nativa (Goroutines) e baixo uso de memória para High Throughput. |
| **AWS SDK v2** | Padrão de mercado para integração com serviços Cloud. |
| **LocalStack** | Simulação fiel da AWS para ambiente de desenvolvimento (DX). |
| **Mutex (Sync)** | Controle de concorrência para garantir consistência de dados em memória. |
| **Clean Arch** | Isolamento entre Regra de Negócio (Domain) e Infraestrutura (AWS/Web). |

---
## 👨‍💻 Autor

Desenvolvido por **DavidABx** Projeto desenvolvido como POC para sistemas de Clearing Bancária.
