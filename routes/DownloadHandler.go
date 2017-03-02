package routes

import (
	"io"
	"net/http"
	"os"
	"time"

	"log"

	"github.com/gorilla/mux"
)

// DownloadHandler is used to retrieve the file from the CLI
func (h *Handler) DownloadHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
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
	if tok.Uploaded == nil {
		res.WriteHeader(400)
		res.Write([]byte("Token not uploaded"))
		return
	}

	file, err := os.Open(EZCPstorage + token)
	if err != nil {
		h.internalError(res, err)
		return
	}

	res.Header().Set("Content-Type", "application/octet-stream")
	res.WriteHeader(200)
	_, err = io.Copy(res, file)
	if err != nil {
		log.Print("Error during copy from disk")
		return
	}

	err = h.db.TokenDownloaded(tok, time.Now())
	if err != nil {
		log.Print("Can't remove token", token, err)
	}

	err = os.Remove(EZCPstorage + token)
	if err != nil {
		log.Print("Can't remove file ", token, err)
	}
}
