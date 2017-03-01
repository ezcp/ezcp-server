package routes

import (
	"log"
	"net/http"

	db "ezcp.io/ezcp-server/db"
)

// Handler handles HTTP routes
type Handler struct {
	db db.DB
}

// NewHandler returns a routes handler
func NewHandler(db db.DB) *Handler {
	return &Handler{db}
}

func (h *Handler) internalError(res http.ResponseWriter, err error) {
	log.Print(err)

	res.WriteHeader(500)
	res.Write([]byte(err.Error()))
}
