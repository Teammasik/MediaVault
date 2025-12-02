package mediavault

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type CaptionRequest struct {
	FileID  int    `json:"file_id"`
	Caption string `json:"caption"`
}

func EditCaptionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Только POST", http.StatusMethodNotAllowed)
		return
	}

	currentUserIDStr := r.Header.Get("X-User-ID")
	if currentUserIDStr == "" {
		http.Error(w, "Требуется X-User-ID", http.StatusUnauthorized)
		return
	}

	var currentUserID int
	_, err := fmt.Sscanf(currentUserIDStr, "%d", &currentUserID)
	if err != nil || currentUserID <= 0 {
		http.Error(w, "Неверный X-User-ID", http.StatusBadRequest)
		return
	}

	var req CaptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	if req.FileID <= 0 {
		http.Error(w, "Неверный file_id", http.StatusBadRequest)
		return
	}

	var ownerID int
	var isAdmin bool
	var oldCaption string
	err = DB.QueryRow(`
		SELECT mf.user_id, u.is_admin, mf.caption
		FROM media_files mf
		JOIN users u ON u.id = ?
		WHERE mf.id = ?
	`, currentUserID, req.FileID).Scan(&ownerID, &isAdmin, &oldCaption)

	if err == sql.ErrNoRows {
		http.Error(w, "Файл не найден", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}

	if ownerID != currentUserID && !isAdmin {
		http.Error(w, "Нет прав", http.StatusForbidden)
		return
	}

	newCaption := strings.TrimSpace(req.Caption)
	_, err = DB.Exec("UPDATE media_files SET caption = ? WHERE id = ?", newCaption, req.FileID)
	if err != nil {
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}

	action := "edit_caption"
	if newCaption == "" && oldCaption != "" {
		action = "delete_caption"
	}
	details := fmt.Sprintf("old: %q, new: %q", oldCaption, newCaption)
	logActivity(currentUserID, action, "file", req.FileID, details)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Подпись обновлена"})
}
