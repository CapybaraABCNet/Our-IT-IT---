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

type Otvets struct {
	NameOtvet string
	NameTo    string
	MesOtvet  string
	CodeOtvet string
	TimeOtvet string
}

type Message struct {
	ID1    int
	Name   string
	Mes    string
	Code   string
	Time   string
	Otvets []Otvets
}

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

	http.HandleFunc("/forum", forum)
	http.HandleFunc("/otvet", otvet)

	fmt.Println("🚀 Сервер запущен на http://localhost:8031/forum")
	log.Fatal(http.ListenAndServe(":8031", nil))
}

func forum(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("it_forum.html")
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

	name := cookie.Value

	if r.Method == http.MethodPost {
		mes := strings.TrimSpace(r.FormValue("mes"))
		code := r.FormValue("code")

		if mes == "" {
			http.Error(w, "Заполните форму ввода сообщения!", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("INSERT INTO posts(name, message, code) VALUES(?, ?, ?)", name, mes, code)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка базы данных", http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	rows, err := db.Query("SELECT id, name, message, code, create_at FROM posts ORDER BY create_at DESC")
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка взятия сообщений", http.StatusBadRequest)
		return
	}

	defer rows.Close()

	mess := []Message{}

	for rows.Next() {
		var c Message
		rows.Scan(&c.ID1, &c.Name, &c.Mes, &c.Code, &c.Time)

		addRows, err := db.Query("SELECT name_post, name, message, code, create_at FROM answer WHERE post_id = ?", c.ID1)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка взятия сообщений N2", http.StatusBadRequest)
			return
		}

		defer addRows.Close()

		var answer []Otvets
		for addRows.Next() {
			var a Otvets
			addRows.Scan(&a.NameTo, &a.NameOtvet, &a.MesOtvet, &a.CodeOtvet, &a.TimeOtvet)
			answer = append(answer, a)
		}

		c.Otvets = answer
		mess = append(mess, c)

	}

	data := map[string]interface{}{
		"Message": mess,
	}

	tmpl.Execute(w, data)
}

func otvet(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("it_otvet.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка загрузки страницы", http.StatusBadRequest)
		return
	}

	IdOtvet := r.URL.Query().Get("id")
	if IdOtvet == "" {
		http.Error(w, "Нету id!", http.StatusBadRequest)
		return
	}

	var name_post string

	err = db.QueryRow("SELECT name FROM posts WHERE id = ?", IdOtvet).Scan(&name_post)
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

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Вы не авторизовались", http.StatusBadRequest)
		return
	}

	name := cookie.Value

	if r.Method == http.MethodPost {
		mes := strings.TrimSpace(r.FormValue("mess"))
		code := r.FormValue("code")

		if mes == "" {
			http.Error(w, "Заполните форму ввода сообщения!", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("INSERT INTO answer(post_id, name_post, name, message, code) VALUES(?, ?, ?, ?, ?)", IdOtvet, name_post, name, mes, code)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка базы данных", http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	tmpl.Execute(w, IdOtvet)
}
