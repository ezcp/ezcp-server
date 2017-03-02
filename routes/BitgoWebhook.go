package routes

import (
	"net/http"

	"log"

	"encoding/json"

	"ezcp.io/ezcp-server/db"
)

const (
	satoshi  = 1000000
	priceBtc = 0.01
)

// BitgoWebhook is called each time Bitgo is sending us something
func (h *Handler) BitgoWebhook(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var transaction db.BitgoTransaction
	json.NewDecoder(req.Body).Decode(&transaction)

	err := h.db.StoreTransaction(&transaction)
	if err != nil {
		log.Print(err)
		log.Print("transaction", transaction)
		h.internalError(res, err)
	}

	if isSimulation, ok := transaction["simulation"].(bool); ok && isSimulation {
		log.Print("Simulated transaction", transaction)
	}
	if hash, ok := transaction["hash"].(string); !ok {
		paid, err := h.getAmount(hash)
		if err != nil || paid == 0 {
			log.Print("Can't get amount for transaction ", hash)
		}
		if paid >= priceBtc*satoshi {
			log.Print("Customer did pay")

			// TODO store transaction is OK

		} else {
			log.Print("Customer didn't pay enough")
		}

	} else {
		log.Print("transaction without hash ", transaction)
	}

	res.WriteHeader(200)
}

type walletAccount struct {
	Account string  `json:"account"`
	Value   float64 `json:"value"`
}
type walletTransaction struct {
	ID      string          `json:"id"`
	Date    string          `json:"date"`
	Outputs []walletAccount `json:"outputs"`
	Entries []walletAccount `json:"entries"`
	Pending bool            `json:"pending"`
}

func (h *Handler) getAmount(tx string) (float64, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.bitgo.com/api/v1/wallet/"+h.bitgoWallet+"/tx/"+tx, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+h.bitgoToken)
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}

	transaction := walletTransaction{}
	if err := json.NewDecoder(resp.Body).Decode(&transaction); err != nil {
		log.Print("Can't parse transaction")
	}

	var amount float64
	for _, entry := range transaction.Entries {
		if entry.Account == h.bitgoWallet {
			amount = entry.Value
		}
	}
	return amount, nil
}
