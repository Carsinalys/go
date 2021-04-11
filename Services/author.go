package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
)

type Author struct {
	Id        string `json:"id,omitempty" validate:"omitempty,uuid"`
	FirstName string `json:"firstname,omitempty" validate:"required"`
	LastName  string `json:"lastname,omitempty" validate:"required"`
	UserName  string `json:"username,omitempty" validate:"required"`
	Password  string `json:"password,omitempty" validate:"required,gte=4"`
}

func AuthorsGetAllEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	json.NewEncoder(res).Encode(authors)
}

func AuthorGetEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	params := mux.Vars(req)
	for _, author := range authors {
		if author.Id == params["id"] {
			json.NewEncoder(res).Encode(author)
			return
		}
	}
	res.WriteHeader(404)
	res.Write([]byte(`{ "error": "not found" }`))
}

func AuthorDeleteEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	params := mux.Vars(req)
	for index, author := range authors {
		if author.Id == params["id"] {
			authors = append(authors[:index], authors[index+1:]...)
			res.Write([]byte(`{ "error": "deleted" }`))
			return
		}
	}
	res.WriteHeader(404)
	res.Write([]byte(`{ "error": "not found" }`))
}

func AuthorUpdateEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	params := mux.Vars(req)
	var changes Author
	json.NewDecoder(req.Body).Decode(&changes)
	validate := validator.New()
	err := validate.StructExcept(changes, "FirstName", "LastName", "UserName", "Password")
	if err != nil {
		res.WriteHeader(400)
		res.Write([]byte(`{ "error": "` + err.Error() + `" }`))
		return
	}
	for i, author := range authors {
		if author.Id == params["id"] {
			if changes.FirstName != "" {
				author.FirstName = changes.FirstName
			}
			if changes.LastName != "" {
				author.LastName = changes.LastName
			}
			if changes.UserName != "" {
				author.UserName = changes.UserName
			}
			if changes.Password != "" {
				err = validate.Var(changes.Password, "gte=4")
				if err != nil {
					res.WriteHeader(400)
					res.Write([]byte(`{ "error": "` + err.Error() + `" }`))
					return
				}
				hash, _ := bcrypt.GenerateFromPassword([]byte(changes.Password), 10)
				author.Password = string(hash)
			}
			authors[i] = author
			json.NewEncoder(res).Encode(author)
			return
		}
	}
	res.WriteHeader(404)
	res.Write([]byte(`{ "error": "not found" }`))
}

func AuthorCreateEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	var data Author
	json.NewDecoder(req.Body).Decode(&data)
	validate := validator.New()
	err := validate.Struct(data)
	if err != nil {
		res.WriteHeader(400)
		res.Write([]byte(`{ "error": "` + err.Error() + `" }`))
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(data.Password), 10)
	data.Id = uuid.NewV4().String()
	data.Password = string(hash)
	authors = append(authors, data)
	json.NewEncoder(res).Encode(data)
}

func AuthorLoginEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	var data Author
	json.NewDecoder(req.Body).Decode(&data)
	validate := validator.New()
	err := validate.StructExcept(data, "FirstName", "LastName")
	if err != nil {
		res.WriteHeader(400)
		res.Write([]byte(`{ "error": "` + err.Error() + `" }`))
		return
	}
	for _, author := range authors {
		if author.UserName == data.UserName {
			err := bcrypt.CompareHashAndPassword([]byte(author.Password), []byte(data.Password))
			if err != nil {
				res.WriteHeader(500)
				res.Write([]byte(`{ "error": "wrong credentials" }`))
				return
			}
			claims := CustomJWTClaim{
				Id: author.Id,
				StandardClaims: jwt.StandardClaims{
					ExpiresAt: time.Now().Local().Add(time.Hour).Unix(),
					Issuer:    "The polyglot developer",
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, _ := token.SignedString(JWT_SECRET)
			res.Write([]byte(`{ "token": "` + tokenString + `" }`))
			return
		}
	}
	res.WriteHeader(404)
	res.Write([]byte(`{ "error": "user not found" }`))
}
