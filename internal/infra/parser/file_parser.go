package parser

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/davidabx-dev/go-clearing-simulator/internal/domain"
)

// ParseFile lê um arquivo streamado linha a linha (Economiza RAM!)
func ParseFile(reader io.Reader) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		
		// Pular linhas vazias ou curtas demais
		if len(line) < 54 {
			continue
		}

		// PARSING POSICIONAL (O segredo da vaga)
		// Supondo layout: 
		// [0-36] ID
		// [36-40] Origin
		// [40-44] Destiny
		// [44-54] Amount
		
		id := strings.TrimSpace(line[0:36])
		origin := strings.TrimSpace(line[36:40])
		destiny := strings.TrimSpace(line[40:44])
		amountStr := strings.TrimSpace(line[44:54])

		amount, err := strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			// Em produção, logaríamos o erro e continuaríamos (resiliência)
			continue
		}

		transaction, err := domain.NewTransaction(id, origin, destiny, amount)
		if err == nil {
			transactions = append(transactions, transaction)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}