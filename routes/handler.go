package routes

import (
	"io/ioutil"
	"log"
	"net/http"
	"text/template"

	db "ezcp.io/ezcp-server/db"
)

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
