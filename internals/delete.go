package mediavault

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
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

	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		http.Error(w, "Требуется параметр id", http.StatusBadRequest)
		return
	}

	var ownerID int
	var filePath string
	var currentUserIsAdmin bool

	err = DB.QueryRow(`
        SELECT mf.user_id, mf.file_path, u.is_admin
        FROM media_files mf
        JOIN users u ON u.id = ?
        WHERE mf.id = ?
    `, currentUserID, fileID).Scan(&ownerID, &filePath, &currentUserIsAdmin)

	if err == sql.ErrNoRows {
		http.Error(w, "Файл не найден", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}

	if ownerID != currentUserID && !currentUserIsAdmin {
		http.Error(w, "Нет прав на удаление", http.StatusForbidden)
		return
	}

	var originalFilename, caption string
	err = DB.QueryRow(
		"SELECT filename, caption FROM media_files WHERE id = ?", fileID,
	).Scan(&originalFilename, &caption)
	if err != nil {
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}

	_, err = DB.Exec("DELETE FROM media_files WHERE id = ?", fileID)
	if err != nil {
		http.Error(w, "Ошибка удаления из БД", http.StatusInternalServerError)
		return
	}

	fileIDint, _ := strconv.Atoi(fileID)
	details := fmt.Sprintf("filename: %s, caption: %s", originalFilename, caption)
	logActivity(currentUserID, "delete_file", "file", fileIDint, details)

	fullPath := filepath.Join("./uploads", filePath)
	if err := os.Remove(fullPath); err != nil {
		fmt.Printf("Предупреждение: не удалось удалить файл %s: %v\n", fullPath, err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}
