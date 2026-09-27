package images

import (
	"net/http"
	"os"
	"path/filepath"
)

func ServeWithMux(mux *http.ServeMux) {
	mux.HandleFunc("GET /uploads/{filename}", func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename")
		targetPath := filepath.Join(os.Getenv("PATH_UPLOADS"), filename)
		http.ServeFile(w, r, targetPath)
	})
}
