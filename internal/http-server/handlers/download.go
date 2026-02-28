package handlers

import "net/http"

func (h StreamHandler) DownloadFile() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Force file download
		w.Header().Set("Content-Disposition", "attachment")
		w.Header().Set("Content-Type", "application/octet-stream")

		// Reuse existing stream logic
		h.ServerFile().ServeHTTP(w, r)
	})
}
