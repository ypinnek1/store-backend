package main

import (
	"fmt"
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

// ListFilesHandler lists all files in the upload directory
func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	// Define the directory where files are uploaded
	uploadDir := "./uploads"

	// Open the upload directory
	dir, err := os.Open(uploadDir)
	if err != nil {
		http.Error(w, "Unable to open the directory", http.StatusInternalServerError)
		return
	}
	defer dir.Close()

	// Read all files in the directory
	files, err := dir.Readdirnames(0) // 0 means read all files
	if err != nil {
		http.Error(w, "Unable to read files in the directory", http.StatusInternalServerError)
		return
	}

	// Send a response with the file names
	if len(files) == 0 {
		w.Write([]byte("No files found"))
		return
	}

	// Format the list of files as a response
	for i, file := range files {
		// If it's not the last file, append a newline character
		if i < len(files)-1 {
			w.Write([]byte(fmt.Sprintf("%s\n", file)))
		} else {
			w.Write([]byte(fmt.Sprintf("%s", file))) // Don't append newline for the last file
		}
	}
}

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Enable CORS for requests from the frontend (http://localhost:5173)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",                 // Local development frontend
			"https://store-frontend-red.vercel.app", // Production frontend
		},
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

	// Add the /files route to list all uploaded files
	r.Get("/files", ListFilesHandler)

	// Start the server
	http.ListenAndServe(":8080", r)
}
