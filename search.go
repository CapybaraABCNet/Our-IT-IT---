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

type Message struct {
	Name string
	Mes  string
	Code string
	Time string
}

type Acc struct {
	NameAcc string
	TimeAcc string
}

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

	http.HandleFunc("/search", search)

	fmt.Println("🚀 Сервер запущен на http://localhost:8031/search")
	log.Fatal(http.ListenAndServe(":8031", nil))
}

func search(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("it_search.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка загрузки страницы", http.StatusBadRequest)
		return
	}

	var cikl []Message
	var acc []Acc

	if r.Method == http.MethodPost {
		search := strings.TrimSpace(r.FormValue("search"))

		if search == "" {
			http.Error(w, "Вы должны ввести что-то в поле ввода!", http.StatusBadRequest)
			return
		}

		search = "%" + search + "%"
		rows, err := db.Query("SELECT name, message, code, create_at FROM posts WHERE name LIKE ? OR message LIKE ? OR code LIKE ?", search, search, search)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка базы данных N1", http.StatusBadRequest)
			return
		}

		defer rows.Close()

		addRows, err := db.Query("SELECT name, create_at FROM users WHERE name LIKE ?", search)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка базы данных N2", http.StatusBadRequest)
			return
		}

		defer addRows.Close()

		for rows.Next() {
			var c Message
			rows.Scan(&c.Name, &c.Mes, &c.Code, &c.Time)
			cikl = append(cikl, c)
		}

		for addRows.Next() {
			var h Acc
			addRows.Scan(&h.NameAcc, &h.TimeAcc)
			acc = append(acc, h)
		}

	}

	data := map[string]interface{}{
		"Message": cikl,
		"Account": acc,
	}

	tmpl.Execute(w, data)
}
