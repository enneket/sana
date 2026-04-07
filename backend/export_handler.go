package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// MemosExport Memos格式导出结构
type MemosExport struct {
	App        string           `json:"app"`
	Version    string           `json:"version"`
	ExportedAt string           `json:"exported_at"`
	Memos      []MemoExportItem `json:"memos"`
	Sanas      []MemoExportItem `json:"sanas,omitempty"`
}

// MemoExportItem 单条导出项
type MemoExportItem struct {
	UID        string `json:"uid"`
	Content    string `json:"content"`
	Visibility string `json:"visibility"`
	Pinned     bool   `json:"pinned"`
	CreatedTs  int64  `json:"created_ts"`
	UpdatedTs  int64  `json:"updated_ts"`
}

func handleExportMemos(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `
		SELECT uid, content, created_at, updated_at
		FROM sanas WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []MemoExportItem
	for rows.Next() {
		var s Sana
		if err := rows.Scan(&s.UID, &s.Content, &s.CreatedAt, &s.UpdatedAt); err != nil {
			continue
		}
		items = append(items, MemoExportItem{
			UID:        s.UID,
			Content:    s.Content,
			Visibility: "private",
			Pinned:     false,
			CreatedTs:  s.CreatedAt.Unix(),
			UpdatedTs:  s.UpdatedAt.Unix(),
		})
	}

	export := MemosExport{
		App:        "sana",
		Version:    "1.0",
		ExportedAt: time.Now().Format(time.RFC3339),
		Memos:      items,
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"sana_export_%s.zip\"", time.Now().Format("20060102")))

	zw := zip.NewWriter(w)
	defer zw.Close()

	// Write sanas.json
	fw, err := zw.Create("sanas.json")
	if err != nil {
		return
	}
	json.NewEncoder(fw).Encode(export)

	// Write individual .md files
	for _, item := range items {
		fw, err := zw.Create(item.UID + ".md")
		if err != nil {
			continue
		}
		fw.Write([]byte(item.Content))
	}
}
