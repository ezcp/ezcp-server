package routes

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// GetToken shows the homepage for ezcp.io with a Token ready for use
// It's using fields in the request to determine a unique SHA1 hash
// and renders a template
//
// It stores the token in a database with a timestamp. Tokens which weren't used
// are removed after some time.
func (h *Handler) GetToken(res http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

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
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(hash))
}
