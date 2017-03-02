package db

import (
	"errors"
	"net/http"
	"time"

	"log"

	"encoding/json"
)

const (
	satoshi  = 100000000
	priceBtc = 0.01
)

// Account is an account involved in a bitcoin transaction
type Account struct {
	Account string  `json:"account"`
	Value   float64 `json:"value"`
}

// ValueBTC returns the account value in BTC
func (a *Account) ValueBTC() float64 {
	return a.Value / satoshi
}

// Transaction models a bitcoin transaction
type Transaction struct {
	ID      string    `json:"id"`
	Date    string    `json:"date"`
	Outputs []Account `json:"outputs"`
	Entries []Account `json:"entries"`
	Pending bool      `json:"pending"`
}

// IsValid checks if a transaction is valid
func (t *Transaction) IsValid(ourWallet string) (bool, error) {
	acc := t.GetAccountEntry(ourWallet)
	if acc == nil {
		return false, errors.New("EZCP wasn't the recepient of the transaction")
	}
	txdate := t.GetDate()
	validationRequired := txdate.AddDate(0, 0, 1) // one day to validate transaction
	expires := txdate.AddDate(1, 0, 0)
	now := time.Now()
	if t.Pending && validationRequired.After(now) {
		return false, errors.New("Transaction has been pending validation for more than 24h") // no validation in one day
	}
	// could add tests about validation count too...
	if expires.Before(now) {
		return false, errors.New("Subscription has expired")
	}
	amount := acc.ValueBTC()
	if amount > priceBtc {
		return true, nil
	}
	return false, errors.New("Transaction amount wasn't enough")
}

// GetDate returns a time.Time object from json date
func (t *Transaction) GetDate() time.Time {
	tim, _ := time.Parse("2006-01-02T15:04:05.000Z", t.Date)
	return tim
}

// GetAccountEntry returns the account entry for the specified account, or nil if not found
func (t *Transaction) GetAccountEntry(account string) *Account {
	for _, a := range t.Entries {
		if a.Account == account {
			return &a
		}
	}
	return nil
}

// BitgoTransaction returns a transaction from Bitgo
func (db *DB) BitgoTransaction(tx string) (*Transaction, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.bitgo.com/api/v1/tx/"+tx, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	transaction := &Transaction{}
	if err := json.NewDecoder(resp.Body).Decode(transaction); err != nil {
		log.Print("Can't parse transaction")
		return nil, nil
	}

	return transaction, nil
}
