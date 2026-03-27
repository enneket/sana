package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func init() {
	// Skip auth for tests by using a test-only JWT secret
	os.Setenv("JWT_SECRET", "test-secret-for-tests-only")
	os.Setenv("SANA_PASSWORD", "testpassword")
	os.Setenv("NOTES_DIR", os.TempDir() + "/sana_test_notes")
}

func getTestToken() string {
	secret := []byte("test-secret-for-tests-only")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "default-user"})
	tokenStr, _ := token.SignedString(secret)
	return tokenStr
}

func withTestToken(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+getTestToken())
}

// wrapHandler simulates the router wrapping with withAuth for testing
func wrapHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		token, err := jwt.Parse(auth[7:], func(t *jwt.Token) (interface{}, error) {
			return []byte("test-secret-for-tests-only"), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		userID := token.Claims.(jwt.MapClaims)["sub"].(string)
		ctx := context.WithValue(r.Context(), "userID", userID)
		h(w, r.WithContext(ctx))
	}
}

func setupTestData() {
	// Ensure test notes dir
	os.MkdirAll(os.TempDir()+"/sana_test_notes/default-user", 0755)
	loadData()
}

func TestParseFrontmatter_Valid(t *testing.T) {
	content := `---
title: "Test Note"
created: "2024-01-15T10:30:00Z"
updated: "2024-01-16T08:20:00Z"
---
This is the body.`

	fm := parseFrontmatter(content)
	if fm == nil {
		t.Fatal("expected frontmatter, got nil")
	}
	if fm.Title != "Test Note" {
		t.Errorf("title: got %q, want %q", fm.Title, "Test Note")
	}
	if fm.Created != "2024-01-15T10:30:00Z" {
		t.Errorf("created: got %q, want %q", fm.Created, "2024-01-15T10:30:00Z")
	}
	if fm.Updated != "2024-01-16T08:20:00Z" {
		t.Errorf("updated: got %q, want %q", fm.Updated, "2024-01-16T08:20:00Z")
	}
	if fm.Body != "This is the body." {
		t.Errorf("body: got %q, want %q", fm.Body, "This is the body.")
	}
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	content := `This is just a plain note without any YAML.`

	fm := parseFrontmatter(content)
	if fm != nil {
		t.Errorf("expected nil for plain markdown, got %+v", fm)
	}
}

func TestParseFrontmatter_Malformed(t *testing.T) {
	content := `---
title: "Unclosed frontmatter
This is the body.`

	fm := parseFrontmatter(content)
	if fm != nil {
		t.Errorf("expected nil for malformed frontmatter, got %+v", fm)
	}
}

func TestParseFrontmatter_OnlyTitle(t *testing.T) {
	content := `---
title: "Only Title"
---
Body here.`

	fm := parseFrontmatter(content)
	if fm == nil {
		t.Fatal("expected frontmatter, got nil")
	}
	if fm.Title != "Only Title" {
		t.Errorf("title: got %q, want %q", fm.Title, "Only Title")
	}
	if fm.Created != "" {
		t.Errorf("created: expected empty, got %q", fm.Created)
	}
}

func TestParseFrontmatter_SpecialChars(t *testing.T) {
	content := `---
title: "Note with \"quotes\" and colon: and backslash\\"
created: "2024-01-15T10:30:00Z"
updated: "2024-01-16T08:20:00Z"
---
Body content.`

	fm := parseFrontmatter(content)
	if fm == nil {
		t.Fatal("expected frontmatter, got nil")
	}
	// YAML double-quoted strings handle \" and \\ correctly
	if fm.Title == "" {
		t.Error("title should not be empty")
	}
}

func TestWriteFileAtomic_Normal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	data := []byte(`{"key": "value"}`)

	writeFileAtomic(path, data)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("content: got %q, want %q", string(got), string(data))
	}
}

func TestHandleExport_EmptyVault(t *testing.T) {
	setupTestData()
	// Clear existing data
	foldersMu.Lock()
	foldersData = make(map[string]folder)
	foldersMu.Unlock()
	notesMu.Lock()
	notesData = make(map[string]noteRecord)
	notesMu.Unlock()
	saveFolders()
	saveNotes()

	req := httptest.NewRequest("GET", "/api/export", nil)
	withTestToken(req)
	w := httptest.NewRecorder()
	wrapHandler(handleExport)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/zip" {
		t.Errorf("Content-Type: got %q, want %q", ct, "application/zip")
	}
	if disp := w.Header().Get("Content-Disposition"); disp == "" {
		t.Error("Content-Disposition should be set")
	}
}

func TestHandleExport_WithFoldersAndNotes(t *testing.T) {
	setupTestData()
	// Clear and set up
	foldersMu.Lock()
	foldersData = make(map[string]folder)
	notesMu.Lock()
	notesData = make(map[string]noteRecord)

	// Create a folder
	now := "2024-01-15T10:30:00Z"
	fID := "folder-1"
	foldersData[fID] = folder{ID: fID, UserID: "default-user", Name: "工作", CreatedAt: now}

	// Create a note in that folder
	nID := "note-1"
	notesData[nID] = noteRecord{ID: nID, UserID: "default-user", FolderID: fID, Title: "测试笔记", Filename: "note-1.md", CreatedAt: now, UpdatedAt: now}

	foldersMu.Unlock()
	notesMu.Unlock()

	// Write the note file
	dir := os.TempDir() + "/sana_test_notes/default-user/" + fID
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "note-1.md"), []byte("# Hello\n\nContent here."), 0644)

	req := httptest.NewRequest("GET", "/api/export", nil)
	withTestToken(req)
	w := httptest.NewRecorder()
	wrapHandler(handleExport)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleExport_Unauthenticated(t *testing.T) {
	setupTestData()
	req := httptest.NewRequest("GET", "/api/export", nil)
	// No Authorization header
	w := httptest.NewRecorder()
	wrapHandler(handleExport)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
