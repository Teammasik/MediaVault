package main

import (
	"net/http"
	ds "vault/internals"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	ds.DBConnectionAndInit()

	http.HandleFunc("/register", ds.RegisterHandler)
	http.HandleFunc("/login", ds.LoginHandler)
	http.HandleFunc("/upload", ds.UploadHandler)
	http.HandleFunc("/files", ds.FilesHandler)
	http.HandleFunc("/delete-file", ds.DeleteFileHandler)
	http.HandleFunc("/download", ds.DownloadFileHandler)
	http.HandleFunc("/edit-caption", ds.EditCaptionHandler)

	http.ListenAndServe(":8080", nil)
}
