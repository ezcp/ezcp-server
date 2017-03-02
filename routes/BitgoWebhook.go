package routes

import (
	"net/http"

	"log"

	"encoding/json"

	"ezcp.io/ezcp-server/db"
)

// BitgoWebhook is called each time Bitgo is sending us something
func (h *Handler) BitgoWebhook(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var transaction db.BitgoTransaction
	json.NewDecoder(req.Body).Decode(&transaction)
	log.Print(transaction)

	err := h.db.StoreTransaction(&transaction)
	if err != nil {
		log.Print(err)
		h.internalError(res, err)
	}

	res.WriteHeader(200)
}
