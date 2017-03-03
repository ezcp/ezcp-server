package routes

import "net/http"

// Bitcoin returns a new bitcoin address
func (h *Handler) Bitcoin(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	address, err := h.db.NewAddress()
	if err != nil {
		h.internalError(res, err)
		return
	}

	res.WriteHeader(200)
	res.Write([]byte(address))
}
