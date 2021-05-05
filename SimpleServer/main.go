package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"test/quotes"
)

type Quote struct {
	Author string `json:"author"`
	Text   string `json:"text"`
	Source string `json:"source,omitempty"`
}

type App struct {
	db quotes.DB
}

// dummy handler - all routes witch not handled
func hello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello, world!\n")
	io.WriteString(w, "URL:"+r.URL.Path)
}

// quote CRUD handler
func (app *App) handlerQoute(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var body *quotes.Quote
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = app.db.Create(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		io.WriteString(w, "Created")
	case "GET":
		q, err := app.db.Get(app.getQouteKey(r.URL.Path))
		if err != nil {
			http.Error(w, "Quote doesn`t exist", http.StatusBadRequest)
			return
		}

		data, err := json.Marshal(q)
		if err != nil {
			fmt.Println(err)
			return
		}
		io.WriteString(w, string(data))
	case "PUT":
		var body *quotes.Quote
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = app.db.Update(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		io.WriteString(w, "Updated")
	case "DELETE":
		app.db.Delete(app.getQouteKey(r.URL.Path))
		io.WriteString(w, "Deleted")
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

// GET quote list handler
func (app *App) handleQoutesList(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		arr, error := app.db.List()
		if error != nil {
			fmt.Println(error)
			return
		}

		value, err := json.Marshal(arr)
		if err != nil {
			fmt.Println(err)
			return
		}

		io.WriteString(w, string(value))
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func main() {
	db, err := quotes.Open("quotes.db")
	if err != nil {
		log.Fatalln("Cannot open quotes.db:", err)
	}

	defer db.Close()

	app := &App{db: *db}

	prefix := "/api/v1/"

	http.HandleFunc(prefix+"quote/", app.handlerQoute)
	http.HandleFunc(prefix+"quotes/", app.handleQoutesList)
	http.HandleFunc("/", hello)

	error := http.ListenAndServe("localhost:8000", nil)

	if error != nil {
		log.Fatal("ListenAndServe:", error)
	}
}

// get quote author key from path
func (app *App) getQouteKey(path string) string {
	arr := strings.Split(path, "/")
	last := arr[len(arr)-1]

	return last
}
