package internal

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func CreateBookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		logger.Error("Ошибка декодирования", "error", err)
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат запроса"})
		return
	}

	b.ID = uuid.New().String()
	b.Status = "Available"

	_, err := DB.Exec("INSERT INTO books (id, title, author, isbn, year, status) VALUES (?, ?, ?, ?, ?, ?)",
		b.ID, b.Title, b.Author, b.ISBN, b.Year, b.Status)
	if err != nil {
		logger.Error("Ошибка вставки книги в БД", "error", err)
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка сервера"})
		return
	}

	logger.Info("Книга успешно добавлена", "book_id", b.ID)
	respondWithJSON(w, http.StatusCreated, b)
}

func GetBooksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) == 2 {
		GetBookByIDHandler(w, r, pathParts[1])
		return
	}

	author := r.URL.Query().Get("author")
	status := r.URL.Query().Get("status")

	query := "SELECT id, title, author, isbn, year, status FROM books WHERE 1=1"
	var args []interface{}

	if author != "" {
		query += " AND author = ?"
		args = append(args, author)
	}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " LIMIT 10"

	rows, err := DB.Query(query, args...)
	if err != nil {
		logger.Error("Ошибка получения списка книг", "error", err)
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка сервера"})
		return
	}
	defer rows.Close()

	books := []Book{}
	for rows.Next() {
		var b Book
		rows.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Year, &b.Status)
		books = append(books, b)
	}

	respondWithJSON(w, http.StatusOK, books)
}

func GetBookByIDHandler(w http.ResponseWriter, r *http.Request, id string) {
	var b Book
	err := DB.QueryRow("SELECT id, title, author, isbn, year, status FROM books WHERE id = ?", id).
		Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Year, &b.Status)

	if err == sql.ErrNoRows {
		respondWithJSON(w, http.StatusNotFound, map[string]string{"error": "Книга не найдена"})
		return
	} else if err != nil {
		logger.Error("Ошибка БД при поиске книги", "error", err)
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка сервера"})
		return
	}

	respondWithJSON(w, http.StatusOK, b)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат"})
		return
	}

	u.ID = uuid.New().String()
	u.RegistrationDate = time.Now()

	_, err := DB.Exec("INSERT INTO users (id, name, email, registration_date) VALUES (?, ?, ?, ?)",
		u.ID, u.Name, u.Email, u.RegistrationDate)
	if err != nil {
		logger.Error("Ошибка регистрации пользователя", "error", err)
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка сервера"})
		return
	}

	logger.Info("Пользователь зарегистрирован", "user_id", u.ID)
	respondWithJSON(w, http.StatusCreated, u)
}

func CreateIssueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		BookID string `json:"book_id"`
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат"})
		return
	}

	var status string
	err := DB.QueryRow("SELECT status FROM books WHERE id = ?", req.BookID).Scan(&status)
	if err == sql.ErrNoRows {
		respondWithJSON(w, http.StatusNotFound, map[string]string{"error": "Книга не найдена"})
		return
	}

	if status != "Available" {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Книга уже выдана или недоступна"})
		return
	}

	issueDate := time.Now()
	dueDate := issueDate.AddDate(0, 0, 14)

	tx, err := DB.Begin()
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка транзакции"})
		return
	}

	_, err = tx.Exec("INSERT INTO issues (book_id, user_id, issue_date, due_date, return_date) VALUES (?, ?, ?, ?, NULL)",
		req.BookID, req.UserID, issueDate, dueDate)
	if err != nil {
		tx.Rollback()
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка оформления выдачи"})
		return
	}

	_, err = tx.Exec("UPDATE books SET status = 'Issued' WHERE id = ?", req.BookID)
	if err != nil {
		tx.Rollback()
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка обновления статуса книги"})
		return
	}

	tx.Commit()
	logger.Info("Книга выдана", "book_id", req.BookID, "user_id", req.UserID)
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "Книга успешно выдана", "due_date": dueDate.Format("2006-01-02")})
}

func ReturnBookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		BookID string `json:"book_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат"})
		return
	}

	tx, err := DB.Begin()
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка транзакции"})
		return
	}

	_, err = tx.Exec("UPDATE issues SET return_date = ? WHERE book_id = ? AND return_date IS NULL", time.Now(), req.BookID)
	if err != nil {
		tx.Rollback()
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка оформления возврата"})
		return
	}

	_, err = tx.Exec("UPDATE books SET status = 'Available' WHERE id = ?", req.BookID)
	if err != nil {
		tx.Rollback()
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка обновления книги"})
		return
	}

	tx.Commit()
	logger.Info("Книга возвращена в библиотеку", "book_id", req.BookID)
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "Книга успешно возвращена"})
}

func GetUserBooksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Неверный URL"})
		return
	}
	userID := pathParts[1]

	query := `
		SELECT b.id, b.title, b.author, b.isbn, b.year, b.status 
		FROM books b
		JOIN issues i ON b.id = i.book_id
		WHERE i.user_id = ? AND i.return_date IS NULL`

	rows, err := DB.Query(query, userID)
	if err != nil {
		logger.Error("Ошибка получения книг пользователя", "error", err)
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка сервера"})
		return
	}
	defer rows.Close()

	books := []Book{}
	for rows.Next() {
		var b Book
		rows.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Year, &b.Status)
		books = append(books, b)
	}

	respondWithJSON(w, http.StatusOK, books)
}
