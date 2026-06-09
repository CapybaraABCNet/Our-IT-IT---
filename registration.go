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

	http.HandleFunc("/reg", handle)

	fmt.Println("🚀 Сервер запущен на http://localhost:8031/reg")
	log.Fatal(http.ListenAndServe(":8031", nil))
}

func handle(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("it_reg.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка загрузки страницы", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		name := strings.TrimSpace(r.FormValue("name"))
		pass := strings.TrimSpace(r.FormValue("pass"))

		if name == "" || pass == "" {
			http.Error(w, "Вы должны ввести все поля ввода!", http.StatusBadRequest)
			return
		}

		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE name = ?", name).Scan(&count)
		if err != nil {
			log.Println("Ошибка базы данных при проверке:", err)
			http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
			return
		}

		if count > 0 {
			http.Error(w, "Ошибка: такое имя уже есть!", http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка шифрования пароля", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("INSERT INTO users(name, pass) VALUES(?, ?)", name, hashedPassword)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка регистрации аккаунта", http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl.Execute(w, nil)
}
