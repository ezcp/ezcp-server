package routes

import (
	"io"
	"net/http"
	"os"
	"time"

	"ezcp.io/ezcp-server/db"

	"errors"

	"github.com/gorilla/mux"
)

// Upload is used to receive a post'ed document from the CLI
// It stores the resulting file
func (h *Handler) Upload(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	token := mux.Vars(req)["token"]

	tok, err := h.db.GetToken(token)
	if err != nil {
		h.internalError(res, err)
		return
	}
	if tok == nil {
		res.WriteHeader(404)
		res.Write([]byte("Token not found"))
		return
	}
	if tok.Uploaded != nil && !tok.Permanent {
		res.WriteHeader(404)
		res.Write([]byte("Token already uploaded"))
		return
	}

	file, err := os.Create(db.GetFilePath(token))
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

	if !tok.IsValidSize(size) {
		os.Remove(db.GetFilePath(token))
		h.internalError(res, errors.New("File too large"))
		return
	}

	err = h.db.TokenUploaded(token, size, time.Now())
	if err != nil {
		h.internalError(res, err)
		return
	}

	res.WriteHeader(201)
}
