package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// Serve static files (HTML, CSS, JS)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/upload", uploadHandler)

	// Serve uploaded media
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("media"))))

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// uploadHandler processes the image upload
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // Limit upload size to 10MB
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	if !isValidImageType(handler) {
		http.Error(w, "Invalid file type. Only images are allowed.", http.StatusBadRequest)
		return
	}

	// Ensure media directory exists
	if err := os.MkdirAll("media", os.ModePerm); err != nil {
		http.Error(w, "Failed to create media directory", http.StatusInternalServerError)
		return
	}

	// Save the file
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(handler.Filename))
	filepath := filepath.Join("media", filename)

	out, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Failed to save the file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "Failed to save the file", http.StatusInternalServerError)
		return
	}

	// Respond with the file URL
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "File uploaded successfully", "url": "/media/%s"}`, filename)
}

// isValidImageType validates the uploaded file's MIME type
func isValidImageType(fileHeader *multipart.FileHeader) bool {
	allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}
	contentType := fileHeader.Header.Get("Content-Type")
	for _, t := range allowedTypes {
		if strings.EqualFold(contentType, t) {
			return true
		}
	}
	return false
}
