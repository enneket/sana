package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// handleSearchMemos handles GET /api/memos/search?q=<keyword>
func handleSearchMemos(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"memos": []MemoResponse{}, "total": 0})
		return
	}

	ctx := context.Background()
	rows, err := db.Query(ctx, `
		SELECT id, uid, user_id, content, created_at, updated_at
		FROM memos
		WHERE user_id = $1 AND content ILIKE '%' || $2 || '%'
		ORDER BY updated_at DESC
	`, userID, q)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var memos []MemoResponse
	for rows.Next() {
		var m Memo
		if err := rows.Scan(&m.ID, &m.UID, &m.UserID, &m.Content, &m.CreatedAt, &m.UpdatedAt); err != nil {
			continue
		}
		memos = append(memos, m.ToResponse())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"memos": memos,
		"total": len(memos),
	})
}
