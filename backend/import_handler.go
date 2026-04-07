package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

var importCtx = context.Background()

type ImportResult struct {
	MemosImported int      `json:"sanas_imported"`
	Errors        []string `json:"errors,omitempty"`
}

type MemosImportFormat struct {
	App    string           `json:"app"`
	Version string          `json:"version"`
	Memos  []MemoExportItem `json:"sanas"`
}

func handleImportMemos(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	maxBytes := int64(50 << 20)
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "no file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	zr, err := zip.NewReader(file, maxBytes)
	if err != nil {
		http.Error(w, "invalid zip file", http.StatusBadRequest)
		return
	}

	var sanasJSON *MemosImportFormat
	contentMap := make(map[string]string)

	for _, f := range zr.File {
		if f.Name == "sanas.json" {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			data, _ := io.ReadAll(rc)
			rc.Close()
			json.Unmarshal(data, &sanasJSON)
		} else if strings.HasSuffix(f.Name, ".md") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			content, _ := io.ReadAll(rc)
			rc.Close()
			uid := strings.TrimSuffix(f.Name, ".md")
			contentMap[uid] = string(content)
		}
	}

	if sanasJSON == nil {
		http.Error(w, "invalid sanas format: sanas.json not found", http.StatusBadRequest)
		return
	}

	if sanasJSON.App != "sana" && sanasJSON.App != "sanas" {
		http.Error(w, "unsupported format: app must be 'sana' or 'sanas'", http.StatusBadRequest)
		return
	}

	result := ImportResult{MemosImported: 0}
	now := time.Now()

	for _, m := range sanasJSON.Memos {
		content := m.Content
		if content == "" {
			content, _ = contentMap[m.UID]
		}
		if content == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("memo %s: empty content", m.UID))
			continue
		}

		uid := m.UID
		if uid == "" {
			uid = uuid.New().String()
		}

		createdTs := time.Unix(m.CreatedTs, 0)
		if createdTs.Year() < 2000 {
			createdTs = now
		}
		updatedTs := time.Unix(m.UpdatedTs, 0)
		if updatedTs.Year() < 2000 {
			updatedTs = now
		}

		_, err := db.ExecContext(importCtx, `
			INSERT INTO sanas (uid, user_id, content, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (uid) DO UPDATE SET content = EXCLUDED.content, updated_at = EXCLUDED.updated_at
		`, uid, userID, content, createdTs, updatedTs)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("memo %s: %v", m.UID, err))
			continue
		}
		result.MemosImported++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
