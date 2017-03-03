package routes

import (
	"net/http"

	"ezcp.io/ezcp-server/db"

	"github.com/gorilla/mux"
)

// GetTokenTx returns a permanent token
func (h *Handler) GetTokenTx(res http.ResponseWriter, req *http.Request) {
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

		// check tx
		if err := tx.Check(); err != nil {
			res.WriteHeader(401)
			res.Write([]byte(err.Error()))
			return
		}

		// not in cache also means no token yet
		_tok := h.getToken(req)
		err = h.db.CreateDurableToken(_tok, txhash)
		if err != nil {
			h.internalError(res, err)
			return
		}
		tx.Token = &_tok

		err := h.db.StoreTransaction(tx) // put in cache
		if err != nil {
			h.internalError(res, err)
			return
		}
		res.WriteHeader(201)
		res.Write([]byte(*tx.Token))
		return
	}

	// check tx, it could have expired
	if err := tx.Check(); err != nil {
		res.WriteHeader(401)
		res.Write([]byte(err.Error()))
		return
	}
	res.WriteHeader(200)
	res.Write([]byte(*tx.Token))
}
