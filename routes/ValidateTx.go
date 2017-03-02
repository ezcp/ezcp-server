package routes

import (
	"net/http"

	"ezcp.io/ezcp-server/db"

	"github.com/gorilla/mux"
)

// ValidateTx checks if a bitcoin Tx can be used as an identifier
func (h *Handler) ValidateTx(res http.ResponseWriter, req *http.Request) {

	txhash := mux.Vars(req)["tx"]

	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var err error
	var tx *db.Transaction

	tx, err = h.db.LoadTransaction(txhash)
	if err != nil {
		h.internalError(res, err)
		return
	}
	if tx == nil { // not in cache
		tx, err = h.db.BitgoTransaction(txhash)
		if err != nil {
			h.internalError(res, err)
			return
		}
		h.db.StoreTransaction(tx) // put in cache
	}

	// check tx
	if valid, err := tx.IsValid(h.bitgoWallet); !valid {
		if err != nil {
			res.WriteHeader(401)
			res.Write([]byte(err.Error()))
			return

		}
	}

	res.WriteHeader(200)
	res.Write([]byte("Valid transaction! Many thanks from EZCP's team"))
}
