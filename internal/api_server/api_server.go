package apiserver

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"syscall"
	"time"
)

// Handlers

var devices = make(map[string]RegisterMessage)
var activeSessions = make(map[string]UploadMetadataMessage)
var fileTokens = make(map[string]string)

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func getAvailableStorage() (uint64, error) {
	wd, err := os.Getwd()
	if err != nil {
		return 0, err
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(wd, &stat); err != nil {
		return 0, err
	}
	return uint64(stat.Bavail * uint64(stat.Bsize)), nil
}

// POST /api/localsend/v2/register
func registerDevice(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("Received device registration request")
	var msg RegisterMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Default().Printf("Device registered: %s", msg.Fingerprint)

	devices[msg.Fingerprint] = msg
	w.WriteHeader(http.StatusAccepted)
}

// POST /api/localsend/v2/prepare-upload
func prepareUpload(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("Received upload preparation request")
	var msg UploadMetadataMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Process the upload metadata
	// Create a new session for this upload.
	sessionId := generateRandomString(32)
	activeSessions[sessionId] = msg
	// Check if the file has already existed on disk
	for _, file := range msg.Files {
		if _, err := os.Stat(file.FileName); err == nil {
			http.Error(w, fmt.Sprintf("File %s already exists", file.FileName), http.StatusConflict)
			return
		} else if !os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("Error checking file %s: %v", file.FileName, err), http.StatusInternalServerError)
			return
		}
	}
	// Check if we have enough space
	totalFileSizes := uint64(0)
	for _, file := range msg.Files {
		totalFileSizes += uint64(file.Size)
	}

	availableStorage, err := getAvailableStorage()
	if err != nil {
		http.Error(w, "Failed to get available storage", http.StatusInternalServerError)
		return
	}
	if totalFileSizes > availableStorage {
		log.Default().Printf("Not enough storage space: required %d bytes, available %d bytes", totalFileSizes, availableStorage)
		http.Error(w, "Not enough storage space", http.StatusInsufficientStorage)
		return
	}

	// Create our response messages
	var response UploadMetadataResponseMessage
	response.SessionID = sessionId
	response.Files = make(map[string]string)
	for _, file := range msg.Files {
		response.Files[file.ID] = generateRandomString(32)
		fileTokens[file.ID] = response.Files[file.ID]
	}

	// Send the response back to the client
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	log.Default().Printf("Upload session prepared: %s", response)
}

// POST /api/localsend/v2/upload?sessionId=mySessionId&fileId=someFileId&token=someFileToken
func uploadFile(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("Received file upload request")
	sessionId := r.URL.Query().Get("sessionId")
	fileId := r.URL.Query().Get("fileId")
	token := r.URL.Query().Get("token")

	// Validate the query parameters
	if sessionId == "" || fileId == "" || token == "" {
		http.Error(w, "Missing query parameters", http.StatusBadRequest)
		return
	}

	// Process the file upload
	// Check if the session exists
	if _, exists := activeSessions[sessionId]; !exists {
		http.Error(w, "Session not found", http.StatusForbidden)
		return
	}

	// Check if the file exists in the session
	if _, exists := activeSessions[sessionId].Files[fileId]; !exists {
		http.Error(w, "File not found", http.StatusForbidden)
		return
	}

	// Check if the token is valid
	if token != fileTokens[fileId] {
		http.Error(w, "Invalid token", http.StatusForbidden)
		return
	}

	// Else create the file and write the body to it
	file, err := os.Create(activeSessions[sessionId].Files[fileId].FileName)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if _, err := io.Copy(file, r.Body); err != nil {
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}
	// File uploaded successfully
	// Remove the file from the session and token list
	delete(activeSessions[sessionId].Files, fileId)
	delete(fileTokens, fileId)

	w.WriteHeader(http.StatusAccepted)
}

// POST /api/localsend/v2/cancel?sessionId=mySessionId
func cancelUpload(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("Received cancel upload request")
	sessionId := r.URL.Query().Get("sessionId")

	// Validate the query parameters
	if sessionId == "" {
		http.Error(w, "Missing query parameters", http.StatusBadRequest)
		return
	}

	// Process the cancel upload request
	if _, exists := activeSessions[sessionId]; !exists {
		http.Error(w, "Session not found", http.StatusForbidden)
		return
	}

	delete(activeSessions, sessionId)
	w.WriteHeader(http.StatusNoContent)
}

// Scrubber process that scrubs any session that do not have any files left
func startScrubberProcess() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cleanupEmptySessions()
	}
}

func cleanupEmptySessions() {
	for sessionId, session := range activeSessions {
		if len(session.Files) == 0 {
			delete(activeSessions, sessionId)
		}
	}
}

// Router
func createRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/localsend/v2/register", registerDevice)
	mux.HandleFunc("/api/localsend/v2/prepare-upload", prepareUpload)
	mux.HandleFunc("/api/localsend/v2/upload", uploadFile)
	mux.HandleFunc("/api/localsend/v2/cancel", cancelUpload)
	return mux
}

// API server
var apiServer *http.Server

func StartAPIServer(addr string, port int) error {
	apiServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", addr, port),
		Handler: createRouter(),
	}

	go startScrubberProcess()

	return apiServer.ListenAndServe()
}
