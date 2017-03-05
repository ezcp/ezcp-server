package routes

import (
	"io"
	"net/http"
	"os"
	"path"
)

// DownloadOS allows easy downloading of binaries
func (h *Handler) DownloadOS(osname string, binName string) {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			res.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		file, err := os.Open(path.Join("bin", "ezcp."+osname))
		defer file.Close()
		if err != nil {
			res.WriteHeader(404)
			return
		}
		res.Header().Set("Content-Type", "application/octet-stream")
		res.Header().Set("Content-Disposition", `attachment; filename="`+binName+`"`)
		io.Copy(res, file)
	}
}
