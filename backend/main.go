package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var defaultUserID = "default-user"
var defaultPasswordHash string

// JSON storage
var (
	usersMu      sync.RWMutex
	usersData    = make(map[string]user)

	foldersMu      sync.RWMutex
	foldersData    = make(map[string]folder)

	notesMu      sync.RWMutex
	notesData    = make(map[string]noteRecord)
)

type user struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	CreatedAt   string `json:"created_at"`
}

type folder struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Name      string  `json:"name"`
	ParentID  *string `json:"parent_id,omitempty"`
	CreatedAt string  `json:"created_at"`
}

type noteRecord struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	FolderID  string `json:"folder_id"`
	Title     string `json:"title"`
	Filename  string `json:"filename"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getJWTSecret() []byte {
	secret := getEnv("JWT_SECRET", "")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable must be set")
	}
	return []byte(secret)
}

func main() {
	notesDir := getEnv("NOTES_DIR", "./notes")
	os.MkdirAll(notesDir, 0755)

	loadData()

	initDefaultUser()

	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("POST /api/auth/login", handleLogin)
	mux.HandleFunc("POST /api/auth/logout", handleLogout)
	mux.HandleFunc("GET /api/auth/me", withAuth(handleMe))

	// Folders
	mux.HandleFunc("GET /api/folders", withAuth(handleListFolders))
	mux.HandleFunc("POST /api/folders", withAuth(handleCreateFolder))
	mux.HandleFunc("PUT /api/folders/", withAuth(handleUpdateFolder))
	mux.HandleFunc("DELETE /api/folders/", withAuth(handleDeleteFolder))

	// Notes
	mux.HandleFunc("GET /api/notes", withAuth(handleListNotes))
	mux.HandleFunc("GET /api/notes/", withAuth(handleGetNote))
	mux.HandleFunc("POST /api/notes", withAuth(handleCreateNote))
	mux.HandleFunc("PUT /api/notes/", withAuth(handleUpdateNote))
	mux.HandleFunc("DELETE /api/notes/", withAuth(handleDeleteNote))

	// Serve Vue frontend
	mux.Handle("/", http.FileServer(http.Dir("../frontend/dist")))

	port := getEnv("PORT", "8080")
	log.Printf("Listening on %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// --- Data persistence ---

func dataDir() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "data")
}

func ensureDataDir() {
	os.MkdirAll(dataDir(), 0755)
}

func loadData() {
	ensureDataDir()
	loadUsers()
	loadFolders()
	loadNotes()
}

func loadUsers() {
	usersMu.Lock()
	defer usersMu.Unlock()
	data, err := os.ReadFile(filepath.Join(dataDir(), "users.json"))
	if err != nil {
		return
	}
	var list []user
	if err := json.Unmarshal(data, &list); err != nil {
		return
	}
	usersData = make(map[string]user)
	for _, u := range list {
		usersData[u.ID] = u
	}
}

func saveUsers() {
	ensureDataDir()
	usersMu.RLock()
	list := make([]user, 0, len(usersData))
	for _, u := range usersData {
		list = append(list, u)
	}
	usersMu.RUnlock()
	data, _ := json.MarshalIndent(list, "", "  ")
	os.WriteFile(filepath.Join(dataDir(), "users.json"), data, 0644)
}

func loadFolders() {
	foldersMu.Lock()
	defer foldersMu.Unlock()
	data, err := os.ReadFile(filepath.Join(dataDir(), "folders.json"))
	if err != nil {
		return
	}
	var list []folder
	if err := json.Unmarshal(data, &list); err != nil {
		return
	}
	foldersData = make(map[string]folder)
	for _, f := range list {
		foldersData[f.ID] = f
	}
}

func saveFolders() {
	ensureDataDir()
	foldersMu.RLock()
	list := make([]folder, 0, len(foldersData))
	for _, f := range foldersData {
		list = append(list, f)
	}
	foldersMu.RUnlock()
	data, _ := json.MarshalIndent(list, "", "  ")
	os.WriteFile(filepath.Join(dataDir(), "folders.json"), data, 0644)
}

func loadNotes() {
	notesMu.Lock()
	defer notesMu.Unlock()
	data, err := os.ReadFile(filepath.Join(dataDir(), "notes.json"))
	if err != nil {
		return
	}
	var list []noteRecord
	if err := json.Unmarshal(data, &list); err != nil {
		return
	}
	notesData = make(map[string]noteRecord)
	for _, n := range list {
		notesData[n.ID] = n
	}
}

func saveNotes() {
	ensureDataDir()
	notesMu.RLock()
	list := make([]noteRecord, 0, len(notesData))
	for _, n := range notesData {
		list = append(list, n)
	}
	notesMu.RUnlock()
	data, _ := json.MarshalIndent(list, "", "  ")
	os.WriteFile(filepath.Join(dataDir(), "notes.json"), data, 0644)
}

func initDefaultUser() {
	password := getEnv("SANA_PASSWORD", "")
	if password == "" {
		log.Fatal("SANA_PASSWORD environment variable must be set")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	defaultPasswordHash = string(hash)

	usersMu.Lock()
	if _, ok := usersData[defaultUserID]; !ok {
		usersData[defaultUserID] = user{
			ID:           defaultUserID,
			Username:     "admin",
			PasswordHash: defaultPasswordHash,
			CreatedAt:    time.Now().Format(time.RFC3339),
		}
	} else {
		// Update password if user exists
		u := usersData[defaultUserID]
		u.PasswordHash = defaultPasswordHash
		usersData[defaultUserID] = u
	}
	usersMu.Unlock()
	saveUsers()
}

func checkPassword(password string) bool {
	usersMu.RLock()
	defer usersMu.RUnlock()
	u, ok := usersData[defaultUserID]
	if !ok {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

// --- HTTP handlers ---

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if !checkPassword(body.Password) {
		http.Error(w, "invalid password", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(defaultUserID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"id": defaultUserID, "username": "admin"})
}

func generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{"sub": userID, "exp": jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour))}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(getJWTSecret())
}

func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		token, err := jwt.Parse(auth[7:], func(t *jwt.Token) (interface{}, error) {
			return getJWTSecret(), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		userID := token.Claims.(jwt.MapClaims)["sub"].(string)
		ctx := context.WithValue(r.Context(), "userID", userID)
		next(w, r.WithContext(ctx))
	}
}

// --- Folder handlers ---

func handleListFolders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	foldersMu.RLock()
	var result []folder
	for _, f := range foldersData {
		if f.UserID == userID {
			result = append(result, f)
		}
	}
	foldersMu.RUnlock()
	json.NewEncoder(w).Encode(result)
}

func handleCreateFolder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	var body struct {
		Name     string  `json:"name"`
		ParentID *string `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if body.ParentID != nil {
		foldersMu.RLock()
		_, exists := foldersData[*body.ParentID]
		foldersMu.RUnlock()
		if !exists || *body.ParentID != "" && !folderBelongsToUser(userID, *body.ParentID) {
			http.Error(w, "parent folder not found", http.StatusBadRequest)
			return
		}
	}

	id := uuid.New().String()
	now := time.Now().Format(time.RFC3339)
	f := folder{ID: id, UserID: userID, Name: body.Name, ParentID: body.ParentID, CreatedAt: now}
	foldersMu.Lock()
	foldersData[id] = f
	foldersMu.Unlock()
	saveFolders()
	json.NewEncoder(w).Encode(f)
}

func folderBelongsToUser(userID, folderID string) bool {
	foldersMu.RLock()
	defer foldersMu.RUnlock()
	f, ok := foldersData[folderID]
	return ok && f.UserID == userID
}

func handleUpdateFolder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	id := strings.TrimPrefix(r.URL.Path, "/api/folders/")
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	foldersMu.Lock()
	f, ok := foldersData[id]
	if !ok || f.UserID != userID {
		foldersMu.Unlock()
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	f.Name = body.Name
	foldersData[id] = f
	foldersMu.Unlock()
	saveFolders()
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteFolder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	id := strings.TrimPrefix(r.URL.Path, "/api/folders/")

	deleteFolderRecursive(userID, id)
	w.WriteHeader(http.StatusNoContent)
}

func deleteFolderRecursive(userID, folderID string) {
	notesMu.Lock()
	var notesToDelete []string
	for id, n := range notesData {
		if n.UserID == userID && n.FolderID == folderID {
			notesToDelete = append(notesToDelete, id)
		}
	}
	for _, nid := range notesToDelete {
		deleteNoteFile(userID, notesData[nid].FolderID, notesData[nid].Filename)
		delete(notesData, nid)
	}
	notesMu.Unlock()
	saveNotes()

	// Find child folders
	foldersMu.Lock()
	var childIDs []string
	for fid, f := range foldersData {
		if f.UserID == userID && f.ParentID != nil && *f.ParentID == folderID {
			childIDs = append(childIDs, fid)
		}
	}
	for _, cid := range childIDs {
		deleteFolderRecursive(userID, cid)
	}
	delete(foldersData, folderID)
	foldersMu.Unlock()
	saveFolders()
}

// --- Note handlers ---

func handleListNotes(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	folderID := r.URL.Query().Get("folder_id")
	if folderID == "" {
		http.Error(w, "folder_id required", http.StatusBadRequest)
		return
	}

	if !folderBelongsToUser(userID, folderID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	notesMu.RLock()
	var result []noteMeta
	for _, n := range notesData {
		if n.UserID == userID && n.FolderID == folderID {
			result = append(result, noteMeta{ID: n.ID, Title: n.Title, FolderID: n.FolderID, UpdatedAt: n.UpdatedAt})
		}
	}
	notesMu.RUnlock()
	json.NewEncoder(w).Encode(result)
}

func handleGetNote(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	id := strings.TrimPrefix(r.URL.Path, "/api/notes/")

	notesMu.RLock()
	n, ok := notesData[id]
	notesMu.RUnlock()
	if !ok || n.UserID != userID {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if !folderBelongsToUser(userID, n.FolderID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	content := readNoteFile(userID, n.FolderID, n.Filename)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": n.ID, "title": n.Title, "content": content, "folder_id": n.FolderID,
	})
}

func handleCreateNote(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	var body struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		FolderID string `json:"folder_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" || body.FolderID == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if !folderBelongsToUser(userID, body.FolderID) {
		http.Error(w, "folder not found", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	filename := id + ".md"
	now := time.Now().Format(time.RFC3339)
	writeNoteFile(userID, body.FolderID, filename, body.Content)

	notesMu.Lock()
	notesData[id] = noteRecord{ID: id, UserID: userID, FolderID: body.FolderID, Title: body.Title, Filename: filename, CreatedAt: now, UpdatedAt: now}
	notesMu.Unlock()
	saveNotes()

	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "title": body.Title, "folder_id": body.FolderID})
}

func handleUpdateNote(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	id := strings.TrimPrefix(r.URL.Path, "/api/notes/")
	var body struct {
		Title   *string `json:"title"`
		Content *string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	notesMu.Lock()
	n, ok := notesData[id]
	if !ok || n.UserID != userID {
		notesMu.Unlock()
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if body.Title != nil {
		n.Title = *body.Title
	}
	if body.Content != nil {
		writeNoteFile(userID, n.FolderID, n.Filename, *body.Content)
	}
	n.UpdatedAt = time.Now().Format(time.RFC3339)
	notesData[id] = n
	notesMu.Unlock()
	saveNotes()
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteNote(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	id := strings.TrimPrefix(r.URL.Path, "/api/notes/")

	notesMu.Lock()
	n, ok := notesData[id]
	if !ok || n.UserID != userID {
		notesMu.Unlock()
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	deleteNoteFile(userID, n.FolderID, n.Filename)
	delete(notesData, id)
	notesMu.Unlock()
	saveNotes()
	w.WriteHeader(http.StatusNoContent)
}

// --- File helpers ---

func notesDir(userID, folderID string) string {
	return filepath.Join(getEnv("NOTES_DIR", "./notes"), userID, folderID)
}

func writeNoteFile(userID, folderID, filename, content string) {
	dir := notesDir(userID, folderID)
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644)
}

func readNoteFile(userID, folderID, filename string) string {
	data, _ := os.ReadFile(filepath.Join(notesDir(userID, folderID), filename))
	return string(data)
}

func deleteNoteFile(userID, folderID, filename string) {
	os.Remove(filepath.Join(notesDir(userID, folderID), filename))
}

// --- Types ---

type noteMeta struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	FolderID  string `json:"folder_id"`
	UpdatedAt string `json:"updated_at"`
}
