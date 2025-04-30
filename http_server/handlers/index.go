package handlers

import (
	"fmt"
	"net/http"
	"os"
)

func Index() {
	// Handle the home page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		indexHtml, err := os.ReadFile("http_server/index.html")
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read index.html: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(indexHtml)
	})
}
