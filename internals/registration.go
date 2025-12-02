package mediavault

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}

	if strings.Count(email, "@") != 1 {
		return false
	}

	parts := strings.Split(email, "@")
	local, domain := parts[0], parts[1]

	if local == "" {
		return false
	}

	if domain == "" || !strings.Contains(domain, ".") {
		return false
	}

	match, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
	return match
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка хеширования пароля", http.StatusInternalServerError)
		return
	}

	if !isValidEmail(req.Email) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Неверный формат email",
		})
		return
	}

	query := "INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)"
	_, err = DB.Exec(query, req.Username, req.Email, passwordHash)
	if err != nil {
		if isDuplicate(err) {
			http.Error(w, "User already exists", http.StatusConflict)
		} else {
			http.Error(w, "DB error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered"})
}

func isDuplicate(err error) bool {
	if err == nil {
		return false
	}

	var mysqlErr interface{ ErrorCode() uint16 }
	if errors.As(err, &mysqlErr) {
		return mysqlErr.ErrorCode() == 1062 // ER_DUP_ENTRY
	}

	return false
}
