package mediavault

import (
	"encoding/json"
	"net/http"
)

type FileInfo struct {
	ID        int    `json:"id"`
	FilePath  string `json:"file_path"`
	Filename  string `json:"filename"`
	Caption   string `json:"caption"`
	CreatedBy string `json:"created_by"`
	UserID    int    `json:"user_id"`
}

func FilesHandler(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT mf.id, mf.file_path, mf.filename, mf.caption, u.username, mf.user_id
		FROM media_files mf
		JOIN users u ON mf.user_id = u.id
		ORDER BY mf.id DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []FileInfo
	for rows.Next() {
		var f FileInfo
		err := rows.Scan(&f.ID, &f.FilePath, &f.Filename, &f.Caption, &f.CreatedBy, &f.UserID)
		if err != nil {
			http.Error(w, "Ошибка чтения", http.StatusInternalServerError)
			return
		}
		files = append(files, f)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}
