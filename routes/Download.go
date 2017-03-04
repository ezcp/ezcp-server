package routes

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"strings"

	"github.com/gorilla/mux"
)

// Download is used to retrieve the file from the CLI
func (h *Handler) Download(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	token := mux.Vars(req)["token"]

	hostName := req.Host
	if last := strings.LastIndex(req.Host, ":"); last != -1 {
		hostName = req.Host[:last]
	}
	expectedHostName := "api" + string(token[0]) + ".ezcp.io"
	if hostName != expectedHostName && req.Host != "localhost:8000" {
		res.Header().Set("Location", "https://"+expectedHostName+"/download/"+token)
		res.WriteHeader(301)
		return
	}

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

	file, err := os.Open(h.getFilePath(token))
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

	err = os.Remove(h.getFilePath(token))
	if err != nil {
		log.Print("Can't remove file ", token, err)
	}
}
