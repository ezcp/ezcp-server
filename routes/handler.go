package routes

import (
	"log"
	"net/http"

	"io/ioutil"
	"text/template"

	db "ezcp.io/ezcp-server/db"
)

// EZCPstorage is the storage location path for EZCP
const EZCPstorage = "ezcp-storage/"

// Handler handles HTTP routes
type Handler struct {
	db           *db.DB
	origin       string
	homeTemplate *template.Template
}

// NewHandler returns a routes handler
func NewHandler(db *db.DB, origin string) *Handler {

	indexHTMLFile, err := ioutil.ReadFile("index.html")
	if err != nil {
		panic(err)
	}
	tmpl, err := template.New("Home").Parse(string(indexHTMLFile))
	handler := &Handler{db, origin, tmpl}

	return handler
}

func (h *Handler) internalError(res http.ResponseWriter, err error) {
	log.Print(err)

	res.WriteHeader(500)
	res.Write([]byte(err.Error()))
}
