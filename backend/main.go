package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

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
	var err error
	db, err = sql.Open("sqlite3", "./sana.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	initDB()

	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("POST /api/auth/register", handleRegister)
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

func initDB() {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS folders (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		parent_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (parent_id) REFERENCES folders(id)
	);
	CREATE TABLE IF NOT EXISTS notes (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		folder_id TEXT NOT NULL,
		title TEXT NOT NULL,
		filename TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (folder_id) REFERENCES folders(id)
	);
	CREATE INDEX IF NOT EXISTS idx_folders_user ON folders(user_id);
	CREATE INDEX IF NOT EXISTS idx_notes_user ON notes(user_id);
	CREATE INDEX IF NOT EXISTS idx_notes_folder ON notes(folder_id);
	`
	_, err := db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}
}

// --- Auth handlers ---

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Username == "" || body.Password == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	id := uuid.New().String()
	_, err = db.Exec("INSERT INTO users (id, username, password_hash) VALUES (?, ?, ?)", id, body.Username, string(hash))
	if err != nil {
		http.Error(w, "username taken", http.StatusConflict)
		return
	}

	token, err := generateToken(id)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	var id, hash string
	err := db.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", body.Username).Scan(&id, &hash)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(id)
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
	userID := r.Context().Value("userID").(string)
	var username string
	err := db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"id": userID, "username": username})
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
	rows, err := db.Query("SELECT id, name, parent_id FROM folders WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var folders []folder
	for rows.Next() {
		var f folder
		var parentID sql.NullString
		if err := rows.Scan(&f.ID, &f.Name, &parentID); err != nil {
			continue
		}
		if parentID.Valid {
			f.ParentID = &parentID.String
		}
		folders = append(folders, f)
	}
	json.NewEncoder(w).Encode(folders)
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
		var count int
		db.QueryRow("SELECT COUNT(*) FROM folders WHERE id = ? AND user_id = ?", *body.ParentID, userID).Scan(&count)
		if count == 0 {
			http.Error(w, "parent folder not found", http.StatusBadRequest)
			return
		}
	}

	id := uuid.New().String()
	_, err := db.Exec("INSERT INTO folders (id, user_id, name, parent_id) VALUES (?, ?, ?, ?)", id, userID, body.Name, body.ParentID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(folder{ID: id, Name: body.Name, ParentID: body.ParentID})
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

	result, err := db.Exec("UPDATE folders SET name = ? WHERE id = ? AND user_id = ?", body.Name, id, userID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteFolder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	id := strings.TrimPrefix(r.URL.Path, "/api/folders/")

	deleteFolderRecursive(userID, id)
	w.WriteHeader(http.StatusNoContent)
}

func deleteFolderRecursive(userID, folderID string) {
	// Delete all notes in this folder
	db.Exec("DELETE FROM notes WHERE folder_id = ? AND user_id = ?", folderID, userID)
	// Find and delete child folders (iterating until no more children found)
	for {
		rows, err := db.Query("SELECT id FROM folders WHERE parent_id = ? AND user_id = ?", folderID, userID)
		if err != nil {
			break
		}
		var childIDs []string
		for rows.Next() {
			var id string
			rows.Scan(&id)
			childIDs = append(childIDs, id)
		}
		rows.Close()
		if len(childIDs) == 0 {
			break
		}
		for _, childID := range childIDs {
			deleteFolderRecursive(userID, childID)
		}
	}
	// Delete the folder itself
	db.Exec("DELETE FROM folders WHERE id = ? AND user_id = ?", folderID, userID)
}

// --- Note handlers ---

func handleListNotes(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	folderID := r.URL.Query().Get("folder_id")
	if folderID == "" {
		http.Error(w, "folder_id required", http.StatusBadRequest)
		return
	}

	// Verify folder belongs to user
	var count int
	db.QueryRow("SELECT COUNT(*) FROM folders WHERE id = ? AND user_id = ?", folderID, userID).Scan(&count)
	if count == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	rows, err := db.Query("SELECT id, title, folder_id, updated_at FROM notes WHERE user_id = ? AND folder_id = ? ORDER BY updated_at DESC", userID, folderID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notes []noteMeta
	for rows.Next() {
		var n noteMeta
		if err := rows.Scan(&n.ID, &n.Title, &n.FolderID, &n.UpdatedAt); err != nil {
			continue
		}
		notes = append(notes, n)
	}
	json.NewEncoder(w).Encode(notes)
}

func handleGetNote(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	id := strings.TrimPrefix(r.URL.Path, "/api/notes/")

	var n note
	err := db.QueryRow("SELECT id, title, folder_id, filename FROM notes WHERE id = ? AND user_id = ?", id, userID).Scan(&n.ID, &n.Title, &n.FolderID, &n.Filename)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Verify folder belongs to user
	var count int
	db.QueryRow("SELECT COUNT(*) FROM folders WHERE id = ? AND user_id = ?", n.FolderID, userID).Scan(&count)
	if count == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	content, err := os.ReadFile(notePath(userID, n.FolderID, n.Filename))
	if err != nil {
		content = []byte("")
	}
	n.Content = string(content)
	json.NewEncoder(w).Encode(n)
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

	id := uuid.New().String()
	filename := id + ".md"
	dir := notesDir(userID, body.FolderID)
	os.MkdirAll(dir, 0755)
	os.WriteFile(dir+"/"+filename, []byte(body.Content), 0644)

	_, err := db.Exec("INSERT INTO notes (id, user_id, folder_id, title, filename) VALUES (?, ?, ?, ?, ?)", id, userID, body.FolderID, body.Title, filename)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
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

	var folderID, filename string
	err := db.QueryRow("SELECT folder_id, filename FROM notes WHERE id = ? AND user_id = ?", id, userID).Scan(&folderID, &filename)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if body.Title != nil {
		db.Exec("UPDATE notes SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", *body.Title, id)
	}
	if body.Content != nil {
		os.WriteFile(notesDir(userID, folderID)+"/"+filename, []byte(*body.Content), 0644)
		db.Exec("UPDATE notes SET updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteNote(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	id := strings.TrimPrefix(r.URL.Path, "/api/notes/")

	var folderID, filename string
	err := db.QueryRow("SELECT folder_id, filename FROM notes WHERE id = ? AND user_id = ?", id, userID).Scan(&folderID, &filename)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	os.Remove(notesDir(userID, folderID) + "/" + filename)
	db.Exec("DELETE FROM notes WHERE id = ? AND user_id = ?", id, userID)
	w.WriteHeader(http.StatusNoContent)
}

// --- Helpers ---

func notesDir(userID, folderID string) string {
	wd, _ := os.Getwd()
	return fmt.Sprintf("%s/notes/%s/%s", wd, userID, folderID)
}

func notePath(userID, folderID, filename string) string {
	return notesDir(userID, folderID) + "/" + filename
}

type folder struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id,omitempty"`
}

type noteMeta struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	FolderID  string `json:"folder_id"`
	UpdatedAt string `json:"updated_at"`
}

type note struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	FolderID string `json:"folder_id"`
	Filename string `json:"-"`
}
