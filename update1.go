package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	var err error
	dsn := "root:root@tcp(localhost:3309)/prog"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/update", update)

	fmt.Println("🚀 Сервер запущен на http://localhost:8031/update")
	log.Fatal(http.ListenAndServe(":8031", nil))
}

func update(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("it_update.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка загрузки страницы", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Сначала авторизуйтесь!", http.StatusBadRequest)
		return
	}

	Name := cookie.Value

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE name = ?", Name).Scan(&count)
	if err != nil {
		log.Println("Ошибка базы данных при проверке:", err)
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	if count < 1 {
		http.Error(w, "Такого пользователя нет!", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		name := strings.TrimSpace(r.FormValue("name"))

		if name == "" {
			http.Error(w, "Вы должны ввести новое имя", http.StatusBadRequest)
			return
		}

		if !isValidEnglishNickname(name) {
			http.Error(w, "Никнейм должен содержать только английские буквы (a-z, A-Z)", http.StatusBadRequest)
			return
		}

		var count1 int
		err = db.QueryRow("SELECT COUNT(*) FROM users WHERE name = ?", name).Scan(&count1)
		if err != nil {
			log.Println("Ошибка базы данных при проверке:", err)
			http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
			return
		}

		if count1 > 0 {
			http.Error(w, "Ошибка: такое имя уже есть!", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("UPDATE users SET name = ? WHERE name = ?", name, Name)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка обновления имени", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("UPDATE posts SET name = ? WHERE TRIM(name) = ?", name, Name)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка обновления имени", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("UPDATE answer SET name = ? WHERE TRIM(name) = ?", name, Name)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка обновления имени", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("UPDATE answer SET name_post = ? WHERE TRIM(name_post) = ?", name, Name)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка обновления имени", http.StatusBadRequest)
			return
		}

		updatedCookie := &http.Cookie{
			Name:     "session_id",
			Value:    name,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   2678400,
		}
		http.SetCookie(w, updatedCookie)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return

	}
	tmpl.Execute(w, nil)
}

func isValidEnglishNickname(name string) bool {
	for i := 0; i < len(name); i++ {
		b := name[i]
		if !(b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z') {
			return false
		}
	}
	return len(name) > 0
}
