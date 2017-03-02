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
	db          db.DB
	bitgoToken  string
	bitgoWallet string
}

// NewHandler returns a routes handler
func NewHandler(db db.DB, token string, wallet string) *Handler {
	return &Handler{db, token, wallet}
}

func (h *Handler) internalError(res http.ResponseWriter, err error) {
	log.Print(err)

	res.WriteHeader(500)
	res.Write([]byte(err.Error()))
}
