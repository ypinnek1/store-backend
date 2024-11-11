package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// UploadHandler handles file uploads and saves the file at a custom location
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form to retrieve file data
	err := r.ParseMultipartForm(10 << 20) // Limit file size to 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form data", http.StatusBadRequest)
		return
	}

	// Retrieve the file from the form
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to retrieve file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Define the custom folder to save the uploaded file
	uploadDir := "./uploads" // Change this to your desired location
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		http.Error(w, "Unable to create upload directory", http.StatusInternalServerError)
		return
	}

	// Create the destination file in the specified folder
	dstFile, err := os.Create(filepath.Join(uploadDir, fileHeader.Filename))
	if err != nil {
		http.Error(w, "Unable to create file in the destination folder", http.StatusInternalServerError)
		return
	}
	defer dstFile.Close()

	// Copy the file content to the destination file
	_, err = io.Copy(dstFile, file)
	if err != nil {
		http.Error(w, "Unable to save the file", http.StatusInternalServerError)
		return
	}

	// Send a response back to the client
	w.Write([]byte("File uploaded successfully"))
}

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Enable CORS for requests from the frontend (http://localhost:5173)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"}, // Allow frontend origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the Store!"))
	})

	// Add the /upload route for handling file uploads
	r.Post("/upload", UploadHandler)

	// Start the server
	http.ListenAndServe(":8080", r)
}
