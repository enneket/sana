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

var memoCtx = context.Background()

// handleListMemos handles GET /api/memos
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

	var rows, err = db.Query(memoCtx, `
		SELECT id, uid, user_id, content, created_at, updated_at
		FROM memos
		WHERE user_id = $1 AND ($2 = false OR updated_at < $3)
		ORDER BY updated_at DESC
		LIMIT $4
	`, userID, cursor.IsZero(), cursor, limit)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var memos []MemoResponse
	var lastUpdated *time.Time
	for rows.Next() {
		var m Memo
		if err := rows.Scan(&m.ID, &m.UID, &m.UserID, &m.Content, &m.CreatedAt, &m.UpdatedAt); err != nil {
			continue
		}
		memos = append(memos, m.ToResponse())
		lastUpdated = &m.UpdatedAt
	}

	response := map[string]interface{}{
		"memos": memos,
	}
	if lastUpdated != nil {
		response["next_cursor"] = lastUpdated.Unix()
		response["has_more"] = len(memos) == limit
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleCreateMemo handles POST /api/memos
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
	err := db.QueryRow(memoCtx, `
		INSERT INTO memos (uid, user_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, uid, userID, req.Content, now, now).Scan(&id)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	memo := Memo{
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

// handleGetMemo handles GET /api/memos/:id
func handleGetMemo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	uid := strings.TrimPrefix(r.URL.Path, "/api/memos/")

	var m Memo
	err := db.QueryRow(memoCtx, `
		SELECT id, uid, user_id, content, created_at, updated_at
		FROM memos WHERE uid = $1 AND user_id = $2
	`, uid, userID).Scan(&m.ID, &m.UID, &m.UserID, &m.Content, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.ToResponse())
}

// handleUpdateMemo handles PUT /api/memos/:id
func handleUpdateMemo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	uid := strings.TrimPrefix(r.URL.Path, "/api/memos/")

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
	result, err := db.Exec(memoCtx, `
		UPDATE memos SET content = $1, updated_at = $2
		WHERE uid = $3 AND user_id = $4
	`, req.Content, now, uid, userID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
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

// handleDeleteMemo handles DELETE /api/memos/:id
func handleDeleteMemo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	uid := strings.TrimPrefix(r.URL.Path, "/api/memos/")

	result, err := db.Exec(memoCtx, `DELETE FROM memos WHERE uid = $1 AND user_id = $2`, uid, userID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleGetStats handles GET /api/memos/stats
func handleGetStats(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	var memoCount int
	db.QueryRow(r.Context(),
		"SELECT COUNT(*) FROM memos WHERE user_id = $1", userID).Scan(&memoCount)

	var activeDays int
	db.QueryRow(r.Context(),
		"SELECT COUNT(DISTINCT DATE(created_at)) FROM memos WHERE user_id = $1", userID).Scan(&activeDays)

	rows, _ := db.Query(r.Context(), `
		SELECT DATE(created_at)::text as day, COUNT(*) as count
		FROM memos
		WHERE user_id = $1 AND created_at >= NOW() - INTERVAL '90 days'
		GROUP BY DATE(created_at)
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
		"heatmap":     heatmap,
	})
}
