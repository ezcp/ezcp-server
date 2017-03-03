package routes

import (
	"log"
	"net/http"

	db "ezcp.io/ezcp-server/db"
)

// EZCPstorage is the storage location path for EZCP
const EZCPstorage = "ezcp-storage/"

// Handler handles HTTP routes
type Handler struct {
	db          *db.DB
	bitgoWallet string
	origin      string
}

// NewHandler returns a routes handler
func NewHandler(db *db.DB, wallet string, origin string) *Handler {
	return &Handler{db, wallet, origin}
}

func (h *Handler) internalError(res http.ResponseWriter, err error) {
	log.Print(err)

	res.WriteHeader(500)
	res.Write([]byte(err.Error()))
}
