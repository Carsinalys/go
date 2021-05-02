package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/graphql-go/graphql"
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

var authorType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Author",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"firstname": &graphql.Field{
			Type: graphql.String,
		},
		"lastname": &graphql.Field{
			Type: graphql.String,
		},
		"username": &graphql.Field{
			Type: graphql.String,
		},
		"password": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var authorInputType *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AuthorInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"firstname": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"lastname": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"username": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"password": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})

func RegisterEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	var author Author
	json.NewDecoder(req.Body).Decode(&author)
	validate := validator.New()
	err := validate.Struct(author)
	if err != nil {
		res.WriteHeader(400)
		res.Write([]byte(`{ "error": "` + err.Error() + `"}`))
		return
	}
	author.Id = uuid.NewV4().String()
	hash, _ := bcrypt.GenerateFromPassword([]byte(author.Password), 10)
	author.Password = string(hash)
	_, error := DBConnection.Exec(context.Background(), "insert into authors(id, firstname, lastname, username, password) values($1, $2, $3, $4, $5)", author.Id, author.FirstName, author.LastName, author.UserName, author.Password)
	if error != nil {
		fmt.Fprintf(os.Stderr, "Unable to insert record to database: %v\n", error)
		json.NewEncoder(res).Encode(error.Error())
		return
	}
	json.NewEncoder(res).Encode(author)
}

func LoginEndpoint(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	var data Author
	json.NewDecoder(req.Body).Decode(&data)
	validate := validator.New()
	err := validate.StructExcept(data, "FirstName", "LastName")
	if err != nil {
		res.WriteHeader(400)
		res.Write([]byte(`{ "error": "` + err.Error() + `"}`))
		return
	}
	var userDB Author
	query := fmt.Sprintf("select * from authors where username='%v'", data.UserName)
	error := DBConnection.QueryRow(context.Background(), query).Scan(&userDB.Id, &userDB.FirstName, &userDB.LastName, &userDB.UserName, &userDB.Password)
	if error != nil {
		fmt.Fprintf(os.Stderr, "Unable to find user in database: %v\n", error)
		json.NewEncoder(res).Encode(Author{})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(userDB.Password), []byte(data.Password))
	if err != nil {
		res.WriteHeader(400)
		res.Write([]byte(`{ "error": "invalid password"`))
		return
	}
	claims := CustomJWTClaims{
		Id: userDB.Id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour).Unix(),
			Issuer:    "something you can indetify",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(JWT_SECRET)
	//uncomment fields for working in browser
	cookie := &http.Cookie{
		Name:  "jwt",
		Value: tokenString,
		// Secure:  true,
		Expires: time.Now().Add(24 * time.Hour),
		// HttpOnly: true,
	}
	http.SetCookie(res, cookie)
	res.Write([]byte(`{ "id": "` + claims.Id + `"}`))
}
