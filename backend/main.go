package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var defaultUserID = "default-user"
var defaultPasswordHash string

type user struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	CreatedAt    string `json:"created_at"`
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

	initDB()
	defer closeDB()

	// Ensure default user exists in database
	initDefaultUser()

	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("POST /api/auth/login", handleLogin)
	mux.HandleFunc("POST /api/auth/logout", handleLogout)
	mux.HandleFunc("GET /api/auth/me", withAuth(handleMe))

	// Memos (Timeline)
	mux.HandleFunc("GET /api/memos", withAuth(handleListMemos))
	mux.HandleFunc("POST /api/memos", withAuth(handleCreateMemo))
	mux.HandleFunc("GET /api/memos/", withAuth(handleGetMemo))
	mux.HandleFunc("PUT /api/memos/", withAuth(handleUpdateMemo))
	mux.HandleFunc("DELETE /api/memos/", withAuth(handleDeleteMemo))
	mux.HandleFunc("GET /api/memos/search", withAuth(handleSearchMemos))

	// Import/Export (Memos format) - pending implementation

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

// --- Default user (database-based) ---

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

	// Create user in database if not exists
	ctx := context.Background()
	var exists bool
	err = db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", defaultUserID).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check user: %v", err)
	}

	if !exists {
		_, err = db.Exec(ctx, `
			INSERT INTO users (id, username, password_hash, created_at)
			VALUES ($1, $2, $3, $4)
		`, defaultUserID, "admin", string(hash), time.Now())
		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
	}
}

func checkPassword(password string) bool {
	ctx := context.Background()
	var storedHash string
	err := db.QueryRow(ctx, "SELECT password_hash FROM users WHERE id = $1", defaultUserID).Scan(&storedHash)
	if err != nil {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)) == nil
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

// --- SPA handler ---

type spaHandler struct {
	root string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.root+"/index.html")
}
