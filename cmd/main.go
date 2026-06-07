package main

import (
	"bufio"
	"library-api/internal"
	"log"
	"net/http"
	"os"
	"strings"
)

func loadEnv() map[string]string {
	env := make(map[string]string)
	file, err := os.Open(".env")
	if err != nil {

		return map[string]string{"PORT": "8080", "DB_PATH": "library.db"}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	return env
}

func main() {

	config := loadEnv()
	port := config["PORT"]
	dbPath := config["DB_PATH"]

	if port == "" {
		port = "8080"
	}
	if dbPath == "" {
		dbPath = "library.db"
	}

	internal.InitDB(dbPath)
	defer internal.DB.Close()

	http.HandleFunc("/books", internal.GetBooksHandler)
	http.HandleFunc("/books/", internal.GetBooksHandler)
	http.HandleFunc("/books/create", internal.CreateBookHandler)

	http.HandleFunc("/users", internal.CreateUserHandler)
	http.HandleFunc("/users/", internal.GetUserBooksHandler)

	http.HandleFunc("/issues", internal.CreateIssueHandler)
	http.HandleFunc("/returns", internal.ReturnBookHandler)

	log.Printf("Библиотечный сервер запущен на порту :%s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
