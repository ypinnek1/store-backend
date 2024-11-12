package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

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

	// Get the folder name from the form or URL parameters (if provided)
	folderName := r.FormValue("folder") // Use "folder" form field for folder name
	if folderName == "" {
		folderName = chi.URLParam(r, "folderName") // Fallback to folderName parameter in URL
	}

	// If no folder name is provided, default to "misc"
	if folderName == "" || folderName == "undefined" {
		folderName = "misc"
	}

	// Define the base directory for file uploads
	uploadDir := "./uploads"

	// If a folder name is provided, create the folder path
	uploadDir = filepath.Join(uploadDir, folderName)

	// Create the target directory if it doesn't exist
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

func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	// Define the directory where files are uploaded
	uploadDir := "./uploads"

	// Check if the uploads directory exists
	_, err := os.Stat(uploadDir)
	if os.IsNotExist(err) {
		// If the directory doesn't exist, return an empty list
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	// Open the upload directory
	dir, err := os.Open(uploadDir)
	if err != nil {
		http.Error(w, "Unable to open the directory", http.StatusInternalServerError)
		return
	}
	defer dir.Close()

	// Read all files and folders in the directory
	entries, err := dir.Readdir(0) // 0 means read all files/folders
	if err != nil {
		http.Error(w, "Unable to read files in the directory", http.StatusInternalServerError)
		return
	}

	// If no files or folders exist, return an empty array
	if len(entries) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	// Create a slice to store file/folder info
	var fileItems []map[string]string

	// Loop through each entry in the directory
	for _, entry := range entries {
		item := make(map[string]string)
		item["name"] = entry.Name()

		// Determine if it's a folder or file
		if entry.IsDir() {
			item["type"] = "folder"
		} else {
			item["type"] = "file"
		}

		// Append the item to the list
		fileItems = append(fileItems, item)
	}

	// Send the response as JSON
	w.Header().Set("Content-Type", "application/json")
	response, err := json.Marshal(fileItems)
	if err != nil {
		http.Error(w, "Unable to marshal JSON", http.StatusInternalServerError)
		return
	}
	w.Write(response)
}

// CreateFolderHandler creates a new folder on the server
func CreateFolderHandler(w http.ResponseWriter, r *http.Request) {
	// Get the folder name from the URL parameters
	folderName := chi.URLParam(r, "folderName")
	if folderName == "" {
		http.Error(w, "Folder name is required", http.StatusBadRequest)
		return
	}

	// Define the parent directory for folder creation
	uploadDir := "./uploads" // Change this to your desired parent directory

	// Create the folder at the specified location
	newFolderPath := filepath.Join(uploadDir, folderName)
	err := os.MkdirAll(newFolderPath, os.ModePerm)
	if err != nil {
		http.Error(w, "Unable to create the folder", http.StatusInternalServerError)
		return
	}

	// Send a response back to the client
	w.Write([]byte(fmt.Sprintf("Folder '%s' created successfully", folderName)))
}

// ListFilesInFolderHandler lists all files and folders inside a specified folder
func ListFilesInFolderHandler(w http.ResponseWriter, r *http.Request) {
	// Get the folder name from the URL parameters
	folderName := chi.URLParam(r, "folderName")
	if folderName == "" {
		http.Error(w, "Folder name is required", http.StatusBadRequest)
		return
	}

	// Define the parent directory where folders are stored
	uploadDir := "./uploads" // Change this to your desired parent directory

	// Get the full path of the folder
	folderPath := filepath.Join(uploadDir, folderName)

	// Check if the folder exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		http.Error(w, "Folder does not exist", http.StatusNotFound)
		return
	}

	// Open the folder
	dir, err := os.Open(folderPath)
	if err != nil {
		http.Error(w, "Unable to open the folder", http.StatusInternalServerError)
		return
	}
	defer dir.Close()

	// Read all files and folders in the directory
	entries, err := dir.Readdir(0) // 0 means read all files/folders
	if err != nil {
		http.Error(w, "Unable to read files in the folder", http.StatusInternalServerError)
		return
	}

	// Get the query parameter for sorting by date (optional)
	sortByDate := r.URL.Query().Get("sortByDate")

	// Sort entries by creation time if requested
	if sortByDate == "true" {
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].ModTime().Before(entries[j].ModTime())
		})
	}

	// Create a slice to store file/folder info
	var fileItems []map[string]string

	// Loop through each entry in the directory
	for _, entry := range entries {
		item := make(map[string]string)
		item["name"] = entry.Name()

		// Determine if it's a folder or file
		if entry.IsDir() {
			item["type"] = "folder"
		} else {
			item["type"] = "file"
		}

		// Append the item to the list
		fileItems = append(fileItems, item)
	}

	// Send the response as JSON
	w.Header().Set("Content-Type", "application/json")
	if len(fileItems) == 0 {
		// Return an empty array if no files or folders found
		w.Write([]byte("[]"))
		return
	}

	// Return the JSON response
	response, err := json.Marshal(fileItems)
	if err != nil {
		http.Error(w, "Unable to marshal JSON", http.StatusInternalServerError)
		return
	}
	w.Write(response)
}

// SearchHandler handles searching for files and folders
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	// Define the base directory for search
	baseDir := "./uploads"

	var results []map[string]string

	// Walk through the directory to find matching files/folders
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Match the file/folder name with the search query
		if strings.Contains(strings.ToLower(info.Name()), strings.ToLower(query)) {
			item := map[string]string{
				"name": info.Name(),
				"type": "folder",
			}
			if !info.IsDir() {
				item["type"] = "file"
			}
			results = append(results, item)
		}
		return nil
	})

	if err != nil {
		http.Error(w, "Error searching for files", http.StatusInternalServerError)
		return
	}

	// Return the search results as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	// Get the file name from the URL parameters
	fileName := chi.URLParam(r, "fileName")
	if fileName == "" {
		http.Error(w, "File name is required", http.StatusBadRequest)
		return
	}

	// Define the base directory where files are stored
	uploadDir := "./uploads" // Change this to your desired directory

	// Construct the full file path
	filePath := filepath.Join(uploadDir, fileName)

	// Check if the file exists and is not a directory
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) || fileInfo.IsDir() {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Disposition", "attachment; filename="+fileInfo.Name())
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Unable to open the file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Copy the file content to the response writer
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Error serving the file", http.StatusInternalServerError)
		return
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
			"https://aa8d-67-170-199-42.ngrok-free.app",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "ngrok-skip-browser-warning"},
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

	// Add the /createFolder route to create a new folder
	r.Post("/createFolder/{folderName}", CreateFolderHandler)

	r.Get("/files/{folderName}", ListFilesInFolderHandler)

	// Add the /search route to search for files
	r.Get("/search", SearchHandler)

	r.Get("/download/{fileName}", DownloadHandler)

	// Start the server
	http.ListenAndServe(":8080", r)
}
