package domain

import (
	"errors"
	"time"
)

// Transaction representa uma linha do arquivo de clearing
type Transaction struct {
	ID        string    `json:"id"`
	Origin    string    `json:"origin"`    // Banco de Origem (ex: 341)
	Destiny   string    `json:"destiny"`   // Banco de Destino (ex: 001)
	Amount    int64     `json:"amount"`    // Valor em centavos (evita float!)
	CreatedAt time.Time `json:"created_at"`
}

// NewTransaction cria uma transação validada
func NewTransaction(id, origin, destiny string, amount int64) (*Transaction, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}
	
	return &Transaction{
		ID:        id,
		Origin:    origin,
		Destiny:   destiny,
		Amount:    amount,
		CreatedAt: time.Now(),
	}, nil
}