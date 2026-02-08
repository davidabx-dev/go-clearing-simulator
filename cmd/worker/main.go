package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/davidabx-dev/go-clearing-simulator/internal/domain"
)

// --- SIMULAÇÃO DE BANCO DE DADOS (Thread-Safe) ---
// Usamos Mutex para garantir que dois processos não gravem ao mesmo tempo
type Repository struct {
	mu   sync.Mutex      
	data map[string]bool 
}

// Verifica se já existe (Leitura segura)
func (r *Repository) HasProcessed(id string) bool {
	r.mu.Lock()         
	defer r.mu.Unlock() 
	return r.data[id]
}

// Marca como processado (Escrita segura)
func (r *Repository) MarkAsProcessed(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[id] = true
}

// Instância global do nosso banco
var repo = &Repository{
	data: make(map[string]bool),
}

func main() {
	// Configura credenciais "fake" para LocalStack
	os.Setenv("AWS_ACCESS_KEY_ID", "teste")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "teste")
	os.Setenv("AWS_REGION", "us-east-1")

	ctx := context.TODO()

	fmt.Println("👷 Worker Iniciando... (Com Idempotência e Mutex)")
	
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("Erro config: %v", err)
	}

	client := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})

	queueName := "clearing-transactions"
	queueUrl := getQueueURL(client, queueName)
	fmt.Printf("✅ Worker conectado na fila: %s\n", queueName)
	fmt.Println("🚀 Aguardando mensagens...")

	for {
		// Busca mensagens (Long Polling)
		output, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            &queueUrl,
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     10,
		})

		if err != nil {
			log.Printf("Erro ao buscar mensagens: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, msg := range output.Messages {
			// Processa cada mensagem
			processMessage(ctx, client, queueUrl, msg)
		}
	}
}

func getQueueURL(client *sqs.Client, name string) string {
	res, err := client.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{QueueName: &name})
	if err != nil {
		resCreate, _ := client.CreateQueue(context.TODO(), &sqs.CreateQueueInput{QueueName: &name})
		return *resCreate.QueueUrl
	}
	return *res.QueueUrl
}

func processMessage(ctx context.Context, client *sqs.Client, queueUrl string, msg types.Message) {
	var t domain.Transaction
	
	// 1. Parse do JSON
	if err := json.Unmarshal([]byte(*msg.Body), &t); err != nil {
		log.Printf("❌ Falha no JSON: %v", err)
		deleteMessage(ctx, client, queueUrl, msg.ReceiptHandle)
		return
	}

	// --- 🛑 AQUI ESTÁ A LÓGICA DE IDEMPOTÊNCIA ---
	if repo.HasProcessed(t.ID) {
		fmt.Printf("🛑 DUPLICIDADE: Transação %s já foi processada. Ignorando.\n", t.ID)
		deleteMessage(ctx, client, queueUrl, msg.ReceiptHandle)
		return
	}

	// 2. Simula Processamento
	fmt.Printf("🔄 Processando DOC: %s | Valor: R$ %d \n", t.ID, t.Amount)
	
	// 3. Salva no "Banco"
	repo.MarkAsProcessed(t.ID)

	// 4. Deleta da fila (ACK)
	deleteMessage(ctx, client, queueUrl, msg.ReceiptHandle)
	fmt.Printf("✅ Sucesso! Transação %s finalizada.\n", t.ID)
}

func deleteMessage(ctx context.Context, client *sqs.Client, queueUrl string, receipt *string) {
	_, err := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &queueUrl,
		ReceiptHandle: receipt,
	})
	if err != nil {
		log.Printf("⚠️ Erro ao deletar msg: %v", err)
	}
}