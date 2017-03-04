package routes

import (
	"io/ioutil"
	"log"
	"net/http"
	"text/template"

	"path/filepath"

	"os"

	db "ezcp.io/ezcp-server/db"
)

// EZCPstorage is the storage location path for EZCP
const EZCPstorage = "ezcp-storage/"

// Handler handles HTTP routes
type Handler struct {
	db           *db.DB
	homeTemplate *template.Template
}

// NewHandler returns a routes handler
func NewHandler(db *db.DB) *Handler {

	indexHTMLFile, err := ioutil.ReadFile("index.html")
	if err != nil {
		panic(err)
	}
	tmpl, err := template.New("Home").Parse(string(indexHTMLFile))
	handler := &Handler{db, tmpl}

	return handler
}

func (h *Handler) internalError(res http.ResponseWriter, err error) {
	log.Print(err)

	res.WriteHeader(500)
	res.Write([]byte(err.Error()))
}

func (h *Handler) getFilePath(token string) string {
	first3 := token[0:3]
	next2 := token[3:5]
	path := filepath.Join(EZCPstorage, first3, next2)

	fileinfo, err := os.Stat(path)
	if err != nil || !fileinfo.IsDir() {
		err = os.MkdirAll(path, 0700)
		if err != nil {
			log.Print(err)
		}
	}
	return filepath.Join(path, token)
}
