package db

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

const (
	satoshi  = 100000000
	priceBtc = 0.01
)

// Account is an account involved in a bitcoin transaction
type Account struct {
	Account string  `json:"account"`
	Value   float64 `json:"value"`
	IsMine  bool    `json:"isMine"`
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
	Pending bool      `json:"pending"`

	Token *string `bson:"token,omitempty"` // only present in mongodb
}

// Check checks if a transaction is valid
func (t *Transaction) Check() error {
	acc := t.GetOurAccountEntry()
	if acc == nil {
		return errors.New("EZCP wasn't the recepient of the transaction")
	}
	txdate := t.GetDate()
	expires := txdate.AddDate(1, 0, 0)
	now := time.Now()
	if expires.Before(now) {
		return errors.New("Subscription has expired")
	}
	amount := acc.ValueBTC()
	if amount >= priceBtc {
		return nil
	}
	return errors.New("Transaction amount wasn't enough")
}

// GetDate returns a time.Time object from json date
func (t *Transaction) GetDate() time.Time {
	tim, _ := time.Parse("2006-01-02T15:04:05.000Z", t.Date)
	return tim
}

// GetOurAccountEntry returns the account entry that belongs to us, or nil if not found
func (t *Transaction) GetOurAccountEntry() *Account {
	log.Print(t)
	for _, a := range t.Outputs {
		log.Print(a)
		if a.IsMine {
			return &a
		}
	}
	return nil
}

// BitgoTransaction returns a transaction from Bitgo
func (db *DB) BitgoTransaction(tx string) (*Transaction, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.bitgo.com/api/v1/wallet/"+db.bitgoWallet+"/tx/"+tx, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+string(db.bitgoToken))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	log.Print(resp.StatusCode)

	transaction := &Transaction{}
	if err := json.NewDecoder(resp.Body).Decode(transaction); err != nil {
		log.Print("Can't parse transaction")
		return nil, nil
	}
	return transaction, nil
}

// NewAddress returns a new Bitgo address
func (db *DB) NewAddress() (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://www.bitgo.com/api/v1/wallet/"+db.bitgoWallet+"/address/0", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+string(db.bitgoToken))

	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		return "", err
	}
	var result = make(map[string]interface{})
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result["address"].(string), nil
}
