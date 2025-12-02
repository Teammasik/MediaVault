package mediavault

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const uploadDir = "./uploads"

type UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Только POST", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		http.Error(w, "Требуется заголовок X-User-ID", http.StatusUnauthorized)
		return
	}

	var userID int
	_, err := fmt.Sscanf(userIDStr, "%d", &userID)
	if err != nil || userID <= 0 {
		http.Error(w, "Неверный X-User-ID", http.StatusBadRequest)
		return
	}

	const maxFileSize = 10 << 20 // 10 МБ
	err = r.ParseMultipartForm(maxFileSize)
	if err != nil {
		http.Error(w, "Ошибка: файл слишком большой или повреждён", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Поле 'file' не найдено", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(handler.Filename))
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true,
		".pdf": true, ".doc": true, ".docx": true,
		".txt": true, ".xlsx": true, ".pptx": true,
	}
	if !allowedExts[ext] {
		http.Error(w, "Недопустимый тип файла", http.StatusBadRequest)
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Ошибка чтения файла", http.StatusInternalServerError)
		return
	}

	uniqueName := fmt.Sprintf("%d_%d%s", userID, time.Now().Unix(), ext)
	savePath := filepath.Join(uploadDir, uniqueName)

	if err := os.WriteFile(savePath, fileBytes, 0o644); err != nil {
		http.Error(w, "Ошибка сохранения файла", http.StatusInternalServerError)
		return
	}

	caption := r.FormValue("caption")

	query := `
		INSERT INTO media_files (user_id, filename, file_path, caption)
		VALUES (?, ?, ?, ?)
	`
	_, err = DB.Exec(query, userID, handler.Filename, uniqueName, caption)
	if err != nil {
		os.Remove(savePath)
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	} else {
		var newFileID int
		err = DB.QueryRow("SELECT LAST_INSERT_ID()").Scan(&newFileID)
		if err != nil {
			fmt.Printf("Не удалось получить ID файла: %v\n", err)
		} else {
			details := fmt.Sprintf("filename: %s, caption: %s", handler.Filename, caption)
			logActivity(userID, "upload", "file", newFileID, details)
		}
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"success":true,"message":"Файл загружен"}`))
}
