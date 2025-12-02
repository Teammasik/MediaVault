package mediavault

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		http.Error(w, "Требуется параметр id", http.StatusBadRequest)
		return
	}

	var filePath string
	err := DB.QueryRow("SELECT file_path FROM media_files WHERE id = ?", fileID).Scan(&filePath)
	if err == sql.ErrNoRows {
		http.Error(w, "Файл не найден", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}

	fullPath := filepath.Join("./uploads", filePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "Файл удалён с диска", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment")
	http.ServeFile(w, r, fullPath)
}
