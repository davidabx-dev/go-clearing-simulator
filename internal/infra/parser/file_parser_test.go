package parser_test

import (
	"strings"
	"testing"

	"github.com/davidabx-dev/go-clearing-simulator/internal/infra/parser"
)

func TestParseFile_Success(t *testing.T) {
	// SIMULAÇÃO DO ARQUIVO (Conteúdo na memória)
	// Layout: [0-36]ID [36-40]Origin [40-44]Destiny [44-54]Amount
	// Linha 1: Válida
	// Linha 2: Válida
	// Linha 3: Inválida (curta demais, deve ser ignorada)
	fileContent := `550e8400-e29b-41d4-a716-446655440000000100020000001000
550e8400-e29b-41d4-a716-446655440001000300040000005500
LINHA_INVALIDA_CURTA`

	// Transformamos a string em um io.Reader (como se fosse um arquivo aberto)
	reader := strings.NewReader(fileContent)

	transactions, err := parser.ParseFile(reader)

	if err != nil {
		t.Fatalf("Esperava sucesso, mas deu erro: %v", err)
	}

	// VERIFICAÇÕES (ASSERTS)
	if len(transactions) != 2 {
		t.Errorf("Esperava 2 transações, encontrou %d", len(transactions))
	}

	// Valida a primeira transação
	t1 := transactions[0]
	if t1.Amount != 1000 { // 1000 centavos = R$ 10,00
		t.Errorf("Esperava valor 1000, encontrou %d", t1.Amount)
	}
	if t1.Origin != "0001" {
		t.Errorf("Esperava origem 0001, encontrou %s", t1.Origin)
	}
}