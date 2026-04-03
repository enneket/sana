package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// handleSearchMemos handles GET /api/sanas/search?q=<keyword>
func handleSearchMemos(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"sanas": []SanaResponse{}, "total": 0})
		return
	}

	ctx := context.Background()
	rows, err := db.Query(ctx, `
		SELECT id, uid, user_id, content, created_at, updated_at
		FROM sanas
		WHERE user_id = $1 AND content ILIKE '%' || $2 || '%'
		ORDER BY updated_at DESC
	`, userID, q)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var sanas []SanaResponse
	for rows.Next() {
		var s Sana
		if err := rows.Scan(&s.ID, &s.UID, &s.UserID, &s.Content, &s.CreatedAt, &s.UpdatedAt); err != nil {
			continue
		}
		sanas = append(sanas, s.ToResponse())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sanas": sanas,
		"total": len(sanas),
	})
}
