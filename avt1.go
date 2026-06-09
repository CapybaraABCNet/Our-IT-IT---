package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
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

	http.HandleFunc("/avt", login)

	fmt.Println("🚀 Сервер запущен на http://localhost:8031/avt")
	log.Fatal(http.ListenAndServe(":8031", nil))
}

func login(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("it_avt.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка загрузки страницы", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("session_user")
	if err == nil {
		http.Error(w, "Вы уже вошли в систему", http.StatusAccepted)
		return
	}

	if r.Method == http.MethodPost {
		name := strings.TrimSpace(r.FormValue("name"))
		pass := strings.TrimSpace(r.FormValue("pass"))

		if name == "" || pass == "" {
			http.Error(w, "Вы должны заполнить все поля ввода", http.StatusBadRequest)
			return
		}

		var hashedPassword string

		err = db.QueryRow("SELECT pass FROM users WHERE name = ?", name).Scan(&hashedPassword)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Пользователь не найден", http.StatusBadRequest)
				return
			} else {
				log.Println(err)
				http.Error(w, "Ошибка проверки имени", http.StatusBadRequest)
				return
			}
		}

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(pass))
		if err != nil {
			log.Println(err)
			http.Error(w, "Неверный пароль", http.StatusBadRequest)
			return

		} else {
			cookie = &http.Cookie{
				Name:     "session_id",
				Value:    name,
				Path:     "/",
				HttpOnly: true,
				MaxAge:   2678400,
			}
			http.SetCookie(w, cookie)
			log.Println("🍪 Куки установлены!")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	tmpl.Execute(w, nil)
}
