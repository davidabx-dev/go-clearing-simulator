package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/davidabx-dev/go-clearing-simulator/internal/infra/parser"
	"github.com/davidabx-dev/go-clearing-simulator/internal/infra/queue"
)

func main() {
	// Contexto com timeout para inicialização
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Inicializa o SQS (Cria a fila "clearing-transactions" se não existir)
	fmt.Println("🔌 Conectando ao AWS SQS (LocalStack)...")
	producer, err := queue.NewSQSProducer(ctx, "clearing-transactions")
	if err != nil {
		log.Fatalf("Erro ao iniciar SQS: %v", err)
	}
	fmt.Println("✅ SQS Conectado!")

	// 2. Define o Handler da API
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		// Pega o arquivo do form-data
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Erro ao ler arquivo", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// 3. Parser: Lê o stream do arquivo
		transactions, err := parser.ParseFile(file)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro no parse: %v", err), http.StatusInternalServerError)
			return
		}

		// 4. Producer: Envia para o SQS
		// (Em produção real, isso seria feito em background/goroutines para ser mais rápido)
		count := 0
		for _, t := range transactions {
			err := producer.Publish(r.Context(), t)
			if err != nil {
				log.Printf("Erro ao enviar msg %s: %v", t.ID, err)
				continue
			}
			count++
		}

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "Processado com sucesso! %d transações enviadas para fila.", count)
	})

	fmt.Println("🚀 Servidor rodando na porta 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}