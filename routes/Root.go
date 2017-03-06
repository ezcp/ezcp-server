package routes

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// Root serves the home page
func (h *Handler) Root(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	res.Header().Set("Content-type", "text/html")
	res.WriteHeader(200)
	hash := h.getToken(req)

	exists, err := h.db.TokenExists(hash, false)
	if err != nil {
		h.internalError(res, err)
	}
	if exists {
		h.internalError(res, errors.New("SHA1 collision"))
	}

	err = h.db.CreateToken(hash)
	if err != nil {
		h.internalError(res, err)
	}
	vars := struct {
		Token string
	}{hash}

	err = h.homeTemplate.Execute(res, vars)
	if err != nil {
		log.Print(err)
	}
}

func (h *Handler) getToken(req *http.Request) string {
	now := time.Now().String()
	now2 := strconv.Itoa(time.Now().Nanosecond())
	random := strconv.Itoa(rand.Int())
	useragent := req.UserAgent()

	sha1 := sha1.New()
	sha1.Write([]byte(now))
	sha1.Write([]byte(now2))
	sha1.Write([]byte(random))
	sha1.Write([]byte(useragent))
	bytes := sha1.Sum(nil)
	hash := fmt.Sprintf("%x", bytes)
	return hash
}
