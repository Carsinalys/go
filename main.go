package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"shit/quotes"
	"strings"
)

type Quote struct {
	Author string `json:"author"`
	Text   string `json:"text"`
	Source string `json:"source,omitempty"`
}

type App struct {
	storage map[string]*Quote
}

func hello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello, world!\n")
	io.WriteString(w, "URL:"+r.URL.Path)
}

func (app *App) handlerQoute(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var body Quote
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		app.storage[body.Author] = &body
		fmt.Printf("%v", app.storage[body.Author])
		io.WriteString(w, "Created")
	case "GET":
		record := app.storage[app.getQouteKey(r.URL.Path)]
		if record == nil {
			http.Error(w, "Quote doesn`t exist", http.StatusBadRequest)
			return
		}
		data, err := json.Marshal(record)
		if err != nil {
			fmt.Println(err)
			return
		}
		io.WriteString(w, string(data))
	case "PUT":
		var body Quote
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		app.storage[body.Author] = &body
		fmt.Printf("%v", app.storage[body.Author])
		io.WriteString(w, "Updated")
	case "DELETE":
		delete(app.storage, app.getQouteKey(r.URL.Path))
		io.WriteString(w, "Deleted\n")
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (app *App) handleQoutesList(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var data []*Quote
		for k, v := range app.storage {
			fmt.Println(k)
			fmt.Println(v)
			local := v
			data = append(data, local)
		}

		fmt.Println(data)
		value, err := json.Marshal(data)
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
	fmt.Println(quotes.Init())
	quotes.Init()
	app := &App{
		storage: map[string]*Quote{},
	}
	prefix := "/api/v1/"

	http.HandleFunc(prefix+"qoute/", app.handlerQoute)
	http.HandleFunc(prefix+"qoutes/", app.handleQoutesList)
	http.HandleFunc("/", hello)

	err := http.ListenAndServe("localhost:8000", nil)

	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func (app *App) getQouteKey(path string) string {
	arr := strings.Split(path, "/")
	last := arr[len(arr)-1]

	return last
}
