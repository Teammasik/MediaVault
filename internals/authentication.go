package mediavault

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UserID  int    `json:"user_id,omitempty"`
	IsAdmin bool   `json:"is_admin"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var userID int
	var isAdmin bool
	var storedHash string
	query := "SELECT id, password_hash, is_admin FROM users WHERE username = ?"
	err := DB.QueryRow(query, req.Username).Scan(&userID, &storedHash, &isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			bcrypt.CompareHashAndPassword([]byte("$2a$10$invalidhash"), []byte("invalid"))
			http.Error(w, "Неверные данные", http.StatusUnauthorized)
		} else {
			http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
		bcrypt.CompareHashAndPassword([]byte("$2a$10$invalidhash"), []byte("invalid"))
		http.Error(w, "Неверные данные", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(LoginResponse{
		Success: true,
		Message: "Login successful",
		UserID:  userID,
		IsAdmin: isAdmin,
	})
}
