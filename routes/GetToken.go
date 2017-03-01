package routes

import (
	"crypto"
	"encoding/hex"
	"errors"
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

	now := time.Now().String()
	random := strconv.Itoa(rand.Int())
	useragent := req.UserAgent()

	sha1 := crypto.SHA1.New()
	sha1.Write([]byte(now))
	sha1.Write([]byte(random))
	bytes := sha1.Sum([]byte(useragent))
	hash := hex.EncodeToString(bytes)

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
