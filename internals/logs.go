package mediavault

import "fmt"

func logActivity(userID int, action, targetType string, targetID int, details string) {
	_, err := DB.Exec(`
		INSERT INTO activity_log (user_id, action, target_type, target_id, details)
		VALUES (?, ?, ?, ?, ?)
	`, userID, action, targetType, targetID, details)
	if err != nil {
		fmt.Printf("Ошибка логирования: %v\n", err)
	}
}
