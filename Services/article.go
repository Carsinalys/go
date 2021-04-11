package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"
)

type Article struct {
	Id      string `json:"id,omitempty" validate:"omitempty,uuid"`
	Author  string `json:"author,omitempty" validate:"isdefault"`
	Title   string `json:"title,omitempty" validate:"required"`
	Content string `json:"content,omitempty" validate:"required"`
}

func ArticlesGetAllEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	json.NewEncoder(res).Encode(articles)
}

func ArticleGetEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	params := mux.Vars(req)
	for _, article := range articles {
		if article.Id == params["id"] {
			json.NewEncoder(res).Encode(article)
			return
		}
	}
	res.WriteHeader(404)
	res.Write([]byte(`{ "error": "not found" }`))
}

func ArticleDeleteEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	params := mux.Vars(req)
	user := context.Get(req, "user").(CustomJWTClaim)
	for index, article := range articles {
		if article.Id == params["id"] && article.Author == user.Id {
			articles = append(articles[:index], articles[index+1:]...)
			res.Write([]byte(`{ "error": "deleted" }`))
			return
		}
	}
	res.WriteHeader(404)
	res.Write([]byte(`{ "error": "not found" }`))
}

func ArticleUpdateEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	params := mux.Vars(req)
	user := context.Get(req, "user").(CustomJWTClaim)
	var changes Article
	json.NewDecoder(req.Body).Decode(&changes)
	for i, article := range articles {
		if article.Id == params["id"] && article.Author == user.Id {
			if changes.Author != "" {
				article.Author = changes.Author
			}
			if changes.Title != "" {
				article.Title = changes.Title
			}
			if changes.Content != "" {
				article.Content = changes.Content
			}
			articles[i] = article
			json.NewEncoder(res).Encode(article)
			return
		}
	}
	res.WriteHeader(404)
	res.Write([]byte(`{ "error": "not found" }`))
}

func ArticleCreateEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	var data Article
	json.NewDecoder(req.Body).Decode(&data)
	user := context.Get(req, "user").(CustomJWTClaim)
	validate := validator.New()
	err := validate.Struct(data)
	if err != nil {
		res.WriteHeader(400)
		res.Write([]byte(`{ "error": "` + err.Error() + `" }`))
		return
	}
	data.Id = uuid.NewV4().String()
	data.Author = user.Id
	articles = append(articles, data)
	json.NewEncoder(res).Encode(data)
}
