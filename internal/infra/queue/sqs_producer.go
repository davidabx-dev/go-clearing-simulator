package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/davidabx-dev/go-clearing-simulator/internal/domain"
)

type SQSProducer struct {
	client   *sqs.Client
	queueURL string
}

// NewSQSProducer inicializa a conexão com a AWS (ou LocalStack)
func NewSQSProducer(ctx context.Context, queueName string) (*SQSProducer, error) {
	// Carrega config padrão
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	// TRUQUE PARA O LOCALSTACK: Força o endpoint local
	client := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566") 
	})

	// Cria a fila se não existir
	result, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create queue: %v", err)
	}

	return &SQSProducer{
		client:   client,
		queueURL: *result.QueueUrl,
	}, nil
}

// Publish envia a transação para a fila
func (p *SQSProducer) Publish(ctx context.Context, transaction *domain.Transaction) error {
	body, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %v", err)
	}

	_, err = p.client.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: aws.String(string(body)),
		QueueUrl:    &p.queueURL,
	})

	if err != nil {
		return fmt.Errorf("failed to send message to SQS: %v", err)
	}

	log.Printf("🚀 Transação enviada para fila: %s | Valor: %d", transaction.ID, transaction.Amount)
	return nil
}