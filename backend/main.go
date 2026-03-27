package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
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

	// Import/Export
	mux.HandleFunc("GET /api/export", withAuth(handleExport))
	mux.HandleFunc("POST /api/import", withAuth(handleImport))

	// Serve Vue frontend (SPA)
	spa := spaHandler{root: "frontend/dist"}
	fs := http.FileServer(http.Dir("frontend/dist"))
	mux.Handle("/assets/", fs)
	mux.Handle("/favicon.svg", fs)
	mux.Handle("/icons.svg", fs)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		spa.ServeHTTP(w, r)
	})

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
	writeFileAtomic(filepath.Join(dataDir(), "folders.json"), data)
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
	writeFileAtomic(filepath.Join(dataDir(), "notes.json"), data)
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
	result := []folder{}
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
	result := []noteMeta{}
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

// --- Export/Import handlers ---

func handleExport(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	// Collect all folders and notes for the user
	foldersMu.RLock()
	var userFolders []folder
	for _, f := range foldersData {
		if f.UserID == userID {
			userFolders = append(userFolders, f)
		}
	}
	foldersMu.RUnlock()

	notesMu.RLock()
	var userNotes []noteRecord
	for _, n := range notesData {
		if n.UserID == userID {
			userNotes = append(userNotes, n)
		}
	}
	notesMu.RUnlock()

	// Build a map of folder name -> folder for quick lookup
	folderNameMap := make(map[string]string) // name -> id
	for _, f := range userFolders {
		folderNameMap[f.ID] = f.Name
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"sana_export_%s.zip\"", time.Now().Format("20060102")))

	zw := zip.NewWriter(w)
	defer zw.Close()

	// Group notes by folder
	notesByFolder := make(map[string][]noteRecord)
	for _, n := range userNotes {
		notesByFolder[n.FolderID] = append(notesByFolder[n.FolderID], n)
	}

	// Write each folder and its notes
	for _, f := range userFolders {
		folderName := f.Name
		// If folder has no notes, still create the entry
		if notes, ok := notesByFolder[f.ID]; ok {
			for _, n := range notes {
				content := readNoteFile(userID, n.FolderID, n.Filename)
				frontmatter := fmt.Sprintf("---\ntitle: %q\ncreated: %q\nupdated: %q\n---\n%s",
					n.Title, n.CreatedAt, n.UpdatedAt, content)

				fileName := fmt.Sprintf("%s/%s.md", folderName, n.Title)
				fw, err := zw.Create(fileName)
				if err != nil {
					continue
				}
				fw.Write([]byte(frontmatter))
			}
		} else {
			// Create empty folder marker (directory entry)
			// zip entries for directories end with /
			fw, err := zw.Create(folderName + "/")
			if err != nil {
				continue
			}
			fw.Write(nil)
		}
	}

	// Handle notes with no folder (root-level) — put in _root folder
	var rootNotes []noteRecord
	for _, n := range notesData {
		// Find notes whose folder doesn't exist or whose folder name is effectively root
		if n.UserID == userID {
			if _, exists := folderNameMap[n.FolderID]; !exists || n.FolderID == "" {
				rootNotes = append(rootNotes, n)
			}
		}
	}
	for _, n := range rootNotes {
		content := readNoteFile(userID, n.FolderID, n.Filename)
		frontmatter := fmt.Sprintf("---\ntitle: %q\ncreated: %q\nupdated: %q\n---\n%s",
			n.Title, n.CreatedAt, n.UpdatedAt, content)
		fileName := fmt.Sprintf("_root/%s.md", n.Title)
		fw, err := zw.Create(fileName)
		if err != nil {
			continue
		}
		fw.Write([]byte(frontmatter))
	}
}

// frontmatterData holds parsed YAML frontmatter fields
type frontmatterData struct {
	Title    string
	Created  string
	Updated  string
	Body     string // content after frontmatter
}

// parseFrontmatter extracts YAML frontmatter from markdown content.
// If no frontmatter is found, returns nil (caller should use filename as title).
func parseFrontmatter(content string) *frontmatterData {
	content = strings.TrimLeft(content, "\r\n ")
	if !strings.HasPrefix(content, "---") {
		return nil // plain markdown, no frontmatter
	}
	// Find the closing ---
	end := strings.Index(content[3:], "---")
	if end < 0 {
		return nil // malformed frontmatter
	}
	yamlContent := content[3 : end+3]
	body := strings.TrimLeft(content[end+6:], "\r\n ")

	var fm map[string]string
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return nil
	}

	return &frontmatterData{
		Title:   fm["title"],
		Created: fm["created"],
		Updated: fm["updated"],
		Body:    body,
	}
}

type importResult struct {
	FoldersImported int           `json:"folders_imported"`
	NotesImported   int           `json:"notes_imported"`
	Skipped         []skippedFile `json:"skipped"`
}

type skippedFile struct {
	File   string `json:"file"`
	Reason string `json:"reason"`
}

func handleImport(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	// Enforce max upload size (50MB)
	maxBytes := int64(50 << 20)
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "no file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read into memory (multipart temp file is already in mem via ParseMultipartForm)
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		http.Error(w, "malformed zip", http.StatusBadRequest)
		return
	}

	result := &importResult{Skipped: []skippedFile{}}

	// Collect all top-level directory names from ZIP
	// ZIP entries are flat paths like "folder/file.md"
	type zipFolder struct {
		name    string
		notes   []zipNote
	}
	zipFolders := make(map[string]*zipFolder) // folder name -> folder struct

	for _, entry := range zr.File {
		name := entry.Name
		if name == "" {
			continue
		}

		// Determine if this is a directory entry (ends with /)
		// Directory entries are skipped — folder creation is handled lazily when .md files are found.
		if strings.HasSuffix(name, "/") {
			continue
		}

		// It's a file — check if it's a .md file
		if !strings.HasSuffix(name, ".md") {
			result.Skipped = append(result.Skipped, skippedFile{File: name, Reason: "not a .md file"})
			continue
		}

		// Parse the path to get folder and filename
		// Subdirectories within a folder are ignored: "folder/subfolder/note.md" → folder="folder", filename="note.md"
		parts := strings.Split(name, "/")
		var folderName, fileName string
		if len(parts) >= 2 {
			folderName = parts[0]
			fileName = parts[len(parts)-1] // take only the last path component
		} else {
			// File at root — no folder
			folderName = ""
			fileName = name
		}

		// Open and read the entry
		rc, err := entry.Open()
		if err != nil {
			result.Skipped = append(result.Skipped, skippedFile{File: name, Reason: "cannot read entry"})
			continue
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			result.Skipped = append(result.Skipped, skippedFile{File: name, Reason: "cannot read entry"})
			continue
		}

		fm := parseFrontmatter(string(content))

		// If no frontmatter or plain .md, use filename (without .md) as title
		title := strings.TrimSuffix(fileName, ".md")
		created := time.Now().Format(time.RFC3339)
		updated := time.Now().Format(time.RFC3339)
		body := string(content)
		if fm != nil {
			if fm.Title != "" {
				title = fm.Title
			}
			if fm.Created != "" {
				created = fm.Created
			}
			if fm.Updated != "" {
				updated = fm.Updated
			}
			body = fm.Body
		}

		if folderName == "" {
			folderName = "_root"
		}

		if _, ok := zipFolders[folderName]; !ok {
			zipFolders[folderName] = &zipFolder{name: folderName, notes: []zipNote{}}
		}
		zipFolders[folderName].notes = append(zipFolders[folderName].notes, zipNote{
			title:    title,
			body:     body,
			created:  created,
			updated:  updated,
			origName: name,
		})
	}

	// Sort folder names for deterministic conflict resolution
	sortedFolderNames := make([]string, 0, len(zipFolders))
	for fn := range zipFolders {
		sortedFolderNames = append(sortedFolderNames, fn)
	}
	sort.Strings(sortedFolderNames)

	// Collect existing folder names for conflict check
	foldersMu.RLock()
	existingFolderNames := make(map[string]bool)
	for _, f := range foldersData {
		if f.UserID == userID {
			existingFolderNames[strings.ToLower(f.Name)] = true
		}
	}
	foldersMu.RUnlock()

	// Create folders with conflict resolution
	folderNameToID := make(map[string]string) // ZIP folder name -> Sana folder ID
	for _, folderName := range sortedFolderNames {
		resolvedName := folderName
		counter := 1
		lowerName := strings.ToLower(resolvedName)
		for existingFolderNames[lowerName] {
			resolvedName = fmt.Sprintf("%s_%d", folderName, counter)
			lowerName = strings.ToLower(resolvedName)
			counter++
		}
		existingFolderNames[lowerName] = true

		// Create the folder in data store
		id := uuid.New().String()
		now := time.Now().Format(time.RFC3339)
		f := folder{ID: id, UserID: userID, Name: resolvedName, CreatedAt: now}
		foldersMu.Lock()
		foldersData[id] = f
		foldersMu.Unlock()
		folderNameToID[folderName] = id
		result.FoldersImported++
	}

	// Collect existing note titles per folder for conflict check
	notesMu.RLock()
	existingNotesPerFolder := make(map[string]map[string]bool) // folderID -> title(lowercase) -> true
	for _, n := range notesData {
		if n.UserID == userID {
			if existingNotesPerFolder[n.FolderID] == nil {
				existingNotesPerFolder[n.FolderID] = make(map[string]bool)
			}
			existingNotesPerFolder[n.FolderID][strings.ToLower(n.Title)] = true
		}
	}
	notesMu.RUnlock()

	// Write notes
	for _, folderName := range sortedFolderNames {
		zf := zipFolders[folderName]
		folderID := folderNameToID[folderName]

		// Sort notes by filename for deterministic processing
		sort.Slice(zf.notes, func(i, j int) bool {
			return zf.notes[i].origName < zf.notes[j].origName
		})

		if existingNotesPerFolder[folderID] == nil {
			existingNotesPerFolder[folderID] = make(map[string]bool)
		}
		usedTitles := make(map[string]bool) // lowercase title -> true for this import session
		for _, nz := range zf.notes {
			title := nz.title
			counter := 1
			lowerTitle := strings.ToLower(title)
			for existingNotesPerFolder[folderID][lowerTitle] || usedTitles[lowerTitle] {
				title = fmt.Sprintf("%s_%d", nz.title, counter)
				lowerTitle = strings.ToLower(title)
				counter++
			}
			usedTitles[lowerTitle] = true
			existingNotesPerFolder[folderID][lowerTitle] = true

			id := uuid.New().String()
			filename := id + ".md"
			writeNoteFile(userID, folderID, filename, nz.body)

			n := noteRecord{
				ID:        id,
				UserID:    userID,
				FolderID:  folderID,
				Title:     title,
				Filename:  filename,
				CreatedAt: nz.created,
				UpdatedAt: nz.updated,
			}
			notesMu.Lock()
			notesData[id] = n
			notesMu.Unlock()
			result.NotesImported++
		}
	}

	saveNotes()
	saveFolders()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

type zipNote struct {
	title    string
	body     string
	created  string
	updated  string
	origName string // original ZIP entry name for sorting
}

// --- SPA handler ---

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

// writeFileAtomic writes data to a temp file then atomically renames it.
// This prevents corruption if a crash occurs mid-write.
func writeFileAtomic(path string, data []byte) {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp_"+filepath.Base(path))
	if err != nil {
		os.WriteFile(path, data, 0644) // fallback
		return
	}
	tmp.Write(data)
	tmp.Close()
	os.Rename(tmp.Name(), path)
}

// --- SPA handler ---

type spaHandler struct {
	root string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.root+"/index.html")
}

type noteMeta struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	FolderID  string `json:"folder_id"`
	UpdatedAt string `json:"updated_at"`
}
