package mediavault

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

const (
	dbUser = "root"
	dbPwd  = ""
	dbHost = "127.0.0.1"
	dbPort = "3306"
	dbName = "vault"
)

var DB *sql.DB

func DBConnectionAndInit() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", dbUser, dbPwd, dbHost, dbPort)

	tempDB, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Ошибка подключения к MySQL:", err)
	}
	defer tempDB.Close()

	_, err = tempDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName))
	if err != nil {
		log.Fatal("Ошибка создания БД:", err)
	}

	dsnWithDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPwd, dbHost, dbPort, dbName)
	DB, err = sql.Open("mysql", dsnWithDB)
	if err != nil {
		log.Fatal("Не удалось создать пул подключений:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Не удалось подключиться к БД:", err)
	}

	DB.SetMaxOpenConns(20)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(300)

	if err := createTables(); err != nil {
		log.Fatal("Ошибка создания таблиц:", err)
	}

	fmt.Println("База данных 'vault' и таблицы готовы.")
}

func createTables() error {
	queries := []string{
		`
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			email VARCHAR(100) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			is_admin BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
		`,
		`
		CREATE TABLE IF NOT EXISTS media_files (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			filename VARCHAR(255) NOT NULL,
			file_path VARCHAR(500) NOT NULL,
			caption TEXT,
			uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
		`,
		`
		CREATE TABLE IF NOT EXISTS activity_log (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			action VARCHAR(50) NOT NULL,
			target_type VARCHAR(20) NOT NULL,
			target_id INT NOT NULL,
			details TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
		`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			return err
		}
	}
	return nil
}
