package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestZip(t *testing.T, entries map[string]string) []byte {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	for name, content := range entries {
		fw, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %q: %v", name, err)
		}
		fw.Write([]byte(content))
	}
	zw.Close()
	return buf.Bytes()
}

func TestHandleImport_ValidZip(t *testing.T) {
	setupTestData()
	// Clear existing data
	foldersMu.Lock()
	foldersData = make(map[string]folder)
	notesMu.Lock()
	notesData = make(map[string]noteRecord)
	foldersMu.Unlock()
	notesMu.Unlock()
	saveFolders()
	saveNotes()

	entries := map[string]string{
		"工作/":           "", // directory entry
		"工作/笔记A.md": `---
title: "笔记A"
created: "2024-01-15T10:30:00Z"
updated: "2024-01-16T08:20:00Z"
---
这是内容。`,
	}
	zipData := createTestZip(t, entries)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.zip")
	part.Write(zipData)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	withTestToken(req)
	w := httptest.NewRecorder()
	wrapHandler(handleImport)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var result importResult
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.FoldersImported != 1 {
		t.Errorf("folders_imported: got %d, want 1", result.FoldersImported)
	}
	if result.NotesImported != 1 {
		t.Errorf("notes_imported: got %d, want 1", result.NotesImported)
	}
}

func TestHandleImport_PlainMarkdownNoFrontmatter(t *testing.T) {
	setupTestData()
	foldersMu.Lock()
	foldersData = make(map[string]folder)
	notesMu.Lock()
	notesData = make(map[string]noteRecord)
	foldersMu.Unlock()
	notesMu.Unlock()
	saveFolders()
	saveNotes()

	entries := map[string]string{
		"工作/":                  "", // directory entry
		"工作/无前置atter笔记.md": "这只是普通 markdown 内容，没有任何 YAML。",
	}
	zipData := createTestZip(t, entries)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.zip")
	part.Write(zipData)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	withTestToken(req)
	w := httptest.NewRecorder()
	wrapHandler(handleImport)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var result importResult
	json.NewDecoder(w.Body).Decode(&result)
	if result.NotesImported != 1 {
		t.Errorf("notes_imported: got %d, want 1", result.NotesImported)
	}
}

func TestHandleImport_FolderNameCollision(t *testing.T) {
	setupTestData()
	foldersMu.Lock()
	foldersData = make(map[string]folder)
	notesMu.Lock()
	notesData = make(map[string]noteRecord)

	// Create existing "工作" folder
	fID := "existing-work-folder"
	foldersData[fID] = folder{ID: fID, UserID: "default-user", Name: "工作", CreatedAt: "2024-01-01T00:00:00Z"}
	foldersMu.Unlock()
	notesMu.Unlock()
	saveFolders()
	saveNotes()

	// Import a zip that also has "工作"
	entries := map[string]string{
		"工作/":           "",
		"工作/笔记B.md": "---\ntitle: \"B\"\n---\ncontent",
	}
	zipData := createTestZip(t, entries)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.zip")
	part.Write(zipData)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	withTestToken(req)
	w := httptest.NewRecorder()
	wrapHandler(handleImport)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	// Should have 2 folders: original "工作" + imported "工作_1"
	foldersMu.RLock()
	count := 0
	for _, f := range foldersData {
		if f.UserID == "default-user" {
			count++
		}
	}
	foldersMu.RUnlock()

	if count != 2 {
		t.Errorf("total folders: got %d, want 2", count)
	}
}

func TestHandleImport_NonMdFiles(t *testing.T) {
	setupTestData()
	foldersMu.Lock()
	foldersData = make(map[string]folder)
	notesMu.Lock()
	notesData = make(map[string]noteRecord)
	foldersMu.Unlock()
	notesMu.Unlock()
	saveFolders()
	saveNotes()

	entries := map[string]string{
		"图片/image.png": string([]byte{0x89, 0x50, 0x4E, 0x47}), // fake PNG header
		"工作/":           "",
		"工作/笔记C.md": "---\ntitle: C\n---\ncontent",
	}
	zipData := createTestZip(t, entries)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.zip")
	part.Write(zipData)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	withTestToken(req)
	w := httptest.NewRecorder()
	wrapHandler(handleImport)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var result importResult
	json.NewDecoder(w.Body).Decode(&result)
	// image.png should be skipped
	if len(result.Skipped) != 1 {
		t.Errorf("skipped count: got %d, want 1", len(result.Skipped))
	}
	if result.Skipped[0].File != "图片/image.png" {
		t.Errorf("skipped file: got %q, want %q", result.Skipped[0].File, "图片/image.png")
	}
}

func TestHandleImport_Unauthenticated(t *testing.T) {
	setupTestData()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/import", body)
	// No Authorization header
	w := httptest.NewRecorder()
	wrapHandler(handleImport)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleImport_NoFile(t *testing.T) {
	setupTestData()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	withTestToken(req)
	w := httptest.NewRecorder()
	wrapHandler(handleImport)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleImport_MalformedZip(t *testing.T) {
	setupTestData()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.zip")
	part.Write([]byte("this is not a zip file"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	withTestToken(req)
	w := httptest.NewRecorder()
	wrapHandler(handleImport)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleImport_EmptyFolder(t *testing.T) {
	setupTestData()
	foldersMu.Lock()
	foldersData = make(map[string]folder)
	notesMu.Lock()
	notesData = make(map[string]noteRecord)
	foldersMu.Unlock()
	notesMu.Unlock()
	saveFolders()
	saveNotes()

	// Empty folder — directory entry but no .md files
	entries := map[string]string{
		"空文件夹/": "",
	}
	zipData := createTestZip(t, entries)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.zip")
	part.Write(zipData)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	withTestToken(req)
	w := httptest.NewRecorder()
	wrapHandler(handleImport)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var result importResult
	json.NewDecoder(w.Body).Decode(&result)
	// Empty folder should be skipped (no notes imported)
	if result.FoldersImported != 0 {
		t.Errorf("folders_imported: got %d, want 0 (empty folder skipped)", result.FoldersImported)
	}
	if result.NotesImported != 0 {
		t.Errorf("notes_imported: got %d, want 0", result.NotesImported)
	}
}

// Helper to create multipart body from zip bytes
func TestHelper_CreateZip(t *testing.T) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	fw, _ := zw.Create("test.txt")
	fw.Write([]byte("hello"))
	zw.Close()

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("failed to read back zip: %v", err)
	}
	if len(zr.File) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(zr.File))
	}
	rc, _ := zr.File[0].Open()
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "hello" {
		t.Errorf("content: got %q, want %q", string(data), "hello")
	}
}

// Test that files at root (no folder) go to _root
func TestHandleImport_RootNotes(t *testing.T) {
	setupTestData()
	foldersMu.Lock()
	foldersData = make(map[string]folder)
	notesMu.Lock()
	notesData = make(map[string]noteRecord)
	foldersMu.Unlock()
	notesMu.Unlock()
	saveFolders()
	saveNotes()

	// Note at root (no folder prefix)
	entries := map[string]string{
		"根笔记.md": "---\ntitle: root\n---\ncontent",
	}
	zipData := createTestZip(t, entries)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.zip")
	part.Write(zipData)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	withTestToken(req)
	w := httptest.NewRecorder()
	wrapHandler(handleImport)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	// Should create a folder literally named "_root"
	foldersMu.RLock()
	var rootFolder folder
	for _, f := range foldersData {
		if f.UserID == "default-user" && f.Name == "_root" {
			rootFolder = f
			break
		}
	}
	foldersMu.RUnlock()

	if rootFolder.ID == "" {
		t.Error("_root folder was not created")
	}
}

// TestSubdirectoryFlattening: subfolders in ZIP should be ignored
func TestHandleImport_SubdirectoryFlatten(t *testing.T) {
	setupTestData()
	foldersMu.Lock()
	foldersData = make(map[string]folder)
	notesMu.Lock()
	notesData = make(map[string]noteRecord)
	foldersMu.Unlock()
	notesMu.Unlock()
	saveFolders()
	saveNotes()

	// Note in subdirectory — should be placed in top-level folder, not nested
	entries := map[string]string{
		"工作/":                     "",
		"工作/子文件夹/深层笔记.md": "---\ntitle: 深层笔记\n---\ncontent",
	}
	zipData := createTestZip(t, entries)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.zip")
	part.Write(zipData)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	withTestToken(req)
	w := httptest.NewRecorder()
	wrapHandler(handleImport)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var result importResult
	json.NewDecoder(w.Body).Decode(&result)
	// Should have 1 folder (工作), 1 note (子文件夹/深层笔记.md stripped to 深层笔记.md)
	if result.FoldersImported != 1 {
		t.Errorf("folders_imported: got %d, want 1", result.FoldersImported)
	}
	if result.NotesImported != 1 {
		t.Errorf("notes_imported: got %d, want 1", result.NotesImported)
	}

	// Verify the note was placed in "工作" folder with title "深层笔记"
	notesMu.RLock()
	var foundNote noteRecord
	for _, n := range notesData {
		if n.UserID == "default-user" {
			foundNote = n
			break
		}
	}
	notesMu.RUnlock()

	if foundNote.Title != "深层笔记" {
		t.Errorf("note title: got %q, want %q", foundNote.Title, "深层笔记")
	}
}

func TestWriteFileAtomic_Fallback(t *testing.T) {
	// Test that fallback to regular WriteFile works when temp file creation fails
	// (This is hard to trigger in a test without mocking, so we just verify basic behavior)
	dir := t.TempDir()
	path := filepath.Join(dir, "atomic.json")
	data := []byte(`{"test": true}`)

	writeFileAtomic(path, data)

	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "test") {
		t.Errorf("content mismatch: got %q", string(got))
	}
}
