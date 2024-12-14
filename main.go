package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Serve the HTML file
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		htmlPath := filepath.Join("static", "index.html")
		htmlFile, err := os.ReadFile(htmlPath)
		if err != nil {
			http.Error(w, "Could not read HTML file", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(htmlFile)
	})

	// Start the server
	fmt.Println("Server is running at :8080")
	http.ListenAndServe(":8080", nil)
}
