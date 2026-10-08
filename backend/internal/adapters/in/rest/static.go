package rest

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// spaHandler serves the built Svelte app and falls back to index.html for
// client-side routes such as /checkout/ord_1.
func spaHandler(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	index := filepath.Join(dir, "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeJSON(w, http.StatusNotFound, errorEnvelope{errorBody{"NOT_FOUND", "Ruta no encontrada"}})
			return
		}
		path := filepath.Join(dir, filepath.Clean("/"+r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, index)
	})
}
