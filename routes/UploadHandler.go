package routes

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// UploadHandler is used to receive a post'ed document from the CLI
// It stores the resulting file
func (h *Handler) UploadHandler(res http.ResponseWriter, req *http.Request) {
	token := mux.Vars(req)["token"]

	exists, err := h.db.TokenExists(token, true)
	if err != nil {
		h.internalError(res, err)
		return
	}
	if !exists {
		res.WriteHeader(404)
		return
	}

	file, err := os.Create("storage/" + token)
	if err != nil {
		h.internalError(res, err)
		return
	}

	var size int64
	size, err = io.Copy(file, req.Body)
	if err != nil {
		h.internalError(res, err)
		return
	}

	err = h.db.TokenUploaded(token, size, time.Now())
	if err != nil {
		h.internalError(res, err)
		return
	}

	res.WriteHeader(201)
}
