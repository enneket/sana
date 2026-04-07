package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var sanaCtx = context.Background()

// handleListMemos handles GET /api/sanas
func handleListMemos(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	cursorStr := r.URL.Query().Get("cursor")
	var cursor time.Time
	if cursorStr != "" {
		ts, err := strconv.ParseInt(cursorStr, 10, 64)
		if err == nil {
			cursor = time.Unix(ts, 0)
		}
	}

	var rows, err = db.QueryContext(sanaCtx, `
		SELECT id, uid, user_id, content, created_at, updated_at
		FROM sanas
		WHERE user_id = $1 AND ($2 OR updated_at < $3)
		ORDER BY updated_at DESC
		LIMIT $4
	`, userID, cursor.IsZero(), cursor, limit)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var sanas []SanaResponse
	var lastUpdated *time.Time
	for rows.Next() {
		var s Sana
		if err := rows.Scan(&s.ID, &s.UID, &s.UserID, &s.Content, &s.CreatedAt, &s.UpdatedAt); err != nil {
			continue
		}
		sanas = append(sanas, s.ToResponse())
		lastUpdated = &s.UpdatedAt
	}

	response := map[string]interface{}{
		"sanas": sanas,
	}
	if lastUpdated != nil {
		response["next_cursor"] = lastUpdated.Unix()
		response["has_more"] = len(sanas) == limit
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleCreateMemo handles POST /api/sanas
func handleCreateMemo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		http.Error(w, "content cannot be empty", http.StatusBadRequest)
		return
	}

	uid := uuid.New().String()
	now := time.Now()

	var id int
	err := db.QueryRowContext(sanaCtx, `
		INSERT INTO sanas (uid, user_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, uid, userID, req.Content, now, now).Scan(&id)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	memo := Sana{
		ID:        id,
		UID:       uid,
		UserID:    userID,
		Content:   req.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(memo.ToResponse())
}

// handleGetMemo handles GET /api/sanas/:id
func handleGetMemo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	uid := strings.TrimPrefix(r.URL.Path, "/api/sanas/")

	var s Sana
	err := db.QueryRowContext(sanaCtx, `
		SELECT id, uid, user_id, content, created_at, updated_at
		FROM sanas WHERE uid = $1 AND user_id = $2
	`, uid, userID).Scan(&s.ID, &s.UID, &s.UserID, &s.Content, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.ToResponse())
}

// handleUpdateMemo handles PUT /api/sanas/:id
func handleUpdateMemo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	uid := strings.TrimPrefix(r.URL.Path, "/api/sanas/")

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		http.Error(w, "content cannot be empty", http.StatusBadRequest)
		return
	}

	now := time.Now()
	result, err := db.ExecContext(sanaCtx, `
		UPDATE sanas SET content = $1, updated_at = $2
		WHERE uid = $3 AND user_id = $4
	`, req.Content, now, uid, userID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         uid,
		"content":    req.Content,
		"updated_ts": now.Unix(),
	})
}

// handleDeleteMemo handles DELETE /api/sanas/:id
func handleDeleteMemo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	uid := strings.TrimPrefix(r.URL.Path, "/api/sanas/")

	result, err := db.ExecContext(sanaCtx, `DELETE FROM sanas WHERE uid = $1 AND user_id = $2`, uid, userID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleGetStats handles GET /api/sanas/stats
func handleGetStats(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	var memoCount int
	db.QueryRowContext(r.Context(),
		"SELECT COUNT(*) FROM sanas WHERE user_id = ?", userID).Scan(&memoCount)

	var activeDays int
	db.QueryRowContext(r.Context(),
		"SELECT COUNT(DISTINCT DATE(created_at)) FROM sanas WHERE user_id = ?", userID).Scan(&activeDays)

	var totalChars int
	db.QueryRowContext(r.Context(),
		"SELECT COALESCE(SUM(LENGTH(content)), 0) FROM sanas WHERE user_id = ?", userID).Scan(&totalChars)

	rows, _ := db.QueryContext(r.Context(), `
		SELECT strftime('%Y-%m-%d', created_at) as day, COUNT(*) as count
		FROM sanas
		WHERE user_id = ? AND created_at >= datetime('now', '-90 days')
		GROUP BY day
	`, userID)

	heatmap := make(map[string]int)
	for rows.Next() {
		var day string
		var count int
		rows.Scan(&day, &count)
		heatmap[day] = count
	}
	rows.Close()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"memo_count":  memoCount,
		"active_days": activeDays,
		"total_chars": totalChars,
		"heatmap":     heatmap,
	})
}
