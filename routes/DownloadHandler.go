package routes

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// DownloadHandler is used to retrieve the file from the CLI
func (h *Handler) DownloadHandler(res http.ResponseWriter, req *http.Request) {
	token := mux.Vars(req)["token"]

	exists, err := h.db.TokenExists(token, false)
	if err != nil {
		h.internalError(res, err)
	}
	if !exists {
		res.WriteHeader(404)
		return
	}

	file, err := os.Open(EZCPstorage + token)
	if err != nil {
		h.internalError(res, err)
		return
	}

	res.WriteHeader(200)
	res.Header().Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(res, file)
	if err != nil {
		h.internalError(res, err)
		return
	}

	err = h.db.TokenDownloaded(token, time.Now())
	if err != nil {
		h.internalError(res, err)
	}
}
