package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

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

	http.HandleFunc("/personal", personal)

	fmt.Println("🚀 Сервер запущен на http://localhost:8031/personal")
	log.Fatal(http.ListenAndServe(":8031", nil))
}

func personal(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("personal.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка загрузки страницы", http.StatusBadRequest)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Нету id!", http.StatusBadRequest)
		return
	}

	var Name string
	var Time string

	err = db.QueryRow("SELECT name, create_at FROM users WHERE id = ?", idStr).Scan(&Name, &Time)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Нет такого пользователя", http.StatusBadRequest)
			return
		} else {
			log.Println(err)
			http.Error(w, "Ошибка выполнения запроса", http.StatusBadRequest)
			return
		}
	}

	data := map[string]interface{}{
		"Name": Name,
		"Time": Time,
	}

	tmpl.Execute(w, data)
}
