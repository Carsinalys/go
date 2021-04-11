package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

// struct for JWT token
type CustomJWTClaim struct {
	Id string `json:"id"`
	jwt.StandardClaims
}

var (
	authors = []Author{
		{Id: "1", FirstName: "Test1", LastName: "WTF1", UserName: "Cardinal1", Password: "123"},
		{Id: "2", FirstName: "Test2", LastName: "WTF2", UserName: "Cardinal2", Password: "123"},
		{Id: "3", FirstName: "Test3", LastName: "WTF3", UserName: "Cardinal3", Password: "123"},
	}
	articles = []Article{
		{Id: "1", Author: "1", Title: "Test title1", Content: "fjkabfjhabhjabshjkasbk"},
		{Id: "2", Author: "2", Title: "Test title2", Content: "daefklcjiuaibuyabuybcauy"},
		{Id: "3", Author: "3", Title: "Test title3", Content: "ciugiusydgvuyasgvuyawyvu"},
	}
	JWT_SECRET []byte = []byte("some secret key")
)

// JWT validation logic
func ValidateJWT(t string) (interface{}, error) {
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method %v", token.Header["alg"])
		}
		return JWT_SECRET, nil
	})
	if err != nil {
		return nil, errors.New(`{ "message": "` + err.Error() + `" }`)
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var tokenData CustomJWTClaim
		mapstructure.Decode(claims, &tokenData)
		return tokenData, nil
	} else {
		return nil, errors.New(`{ "message": "invalid token" }`)
	}
}

// token validation middleware
func ValidateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		authHeader := req.Header.Get("Authorization")
		if authHeader != "" {
			beaerToken := strings.Split(authHeader, " ")
			if len(beaerToken) == 2 {
				decoded, err := ValidateJWT(beaerToken[1])
				if err != nil {
					res.Header().Add("content-type", "application/json")
					res.WriteHeader(500)
					res.Write([]byte(`{ "message": "` + err.Error() + `" }`))
					return
				}

				context.Set(req, "user", decoded)
				next(res, req)
			}
		} else {
			res.Header().Add("content-type", "application/json")
			res.WriteHeader(400)
			res.Write([]byte(`{ "message": "auth header is required" }`))
			return
		}
	})
}

// root route func for testing
func RootRoute(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	res.Write([]byte(`{ "status": "ok" }`))
}

func main() {
	fmt.Println("Starting...")
	router := mux.NewRouter()
	router.HandleFunc("/", RootRoute).Methods("GET")
	router.HandleFunc("/authors", AuthorsGetAllEndpoint).Methods("GET")
	router.HandleFunc("/author/{id}", AuthorGetEndpoint).Methods("GET")
	router.HandleFunc("/author/{id}", AuthorDeleteEndpoint).Methods("DELETE")
	router.HandleFunc("/author/{id}", AuthorUpdateEndpoint).Methods("PUT")
	router.HandleFunc("/author/signup", AuthorCreateEndpoint).Methods("POST")
	router.HandleFunc("/author/login", AuthorLoginEndpoint).Methods("POST")
	router.HandleFunc("/articles", ArticlesGetAllEndpoint).Methods("GET")
	router.HandleFunc("/article/{id}", ArticleGetEndpoint).Methods("GET")
	router.HandleFunc("/article/{id}", ValidateMiddleware(ArticleDeleteEndpoint)).Methods("DELETE")
	router.HandleFunc("/article/{id}", ValidateMiddleware(ArticleUpdateEndpoint)).Methods("PUT")
	router.HandleFunc("/article", ValidateMiddleware(ArticleCreateEndpoint)).Methods("POST")
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})
	headers := handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-Requested-With"})
	origins := handlers.AllowedOrigins([]string{"*"})
	http.ListenAndServe(":8080", handlers.CORS(headers, methods, origins)(router))
}
