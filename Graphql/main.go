package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"github.com/mitchellh/mapstructure"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
)

type GraphQLPayload struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type CustomJWTClaims struct {
	Id string `json:"id"`
	jwt.StandardClaims
}

var (
	authors = []Author{
		{Id: "1", FirstName: "Test1", LastName: "WTF1", UserName: "Cardinal1", Password: "123"},
		{Id: "2", FirstName: "Test2", LastName: "WTF2", UserName: "Cardinal2", Password: "1234"},
		{Id: "3", FirstName: "Test3", LastName: "WTF3", UserName: "Cardinal3", Password: "12345"},
	}
	articles = []Article{
		{Id: "1", Author: "1", Title: "Test title1", Content: "fjkabfjhabhjabshjkasbk"},
		{Id: "2", Author: "2", Title: "Test title2", Content: "daefklcjiuaibuyabuybcauy"},
		{Id: "3", Author: "3", Title: "Test title3", Content: "ciugiusydgvuyasgvuyawyvu"},
	}
	JWT_SECRET []byte = []byte("some strong key")
)

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
		var tokenData CustomJWTClaims
		mapstructure.Decode(claims, &tokenData)
		return tokenData, nil
	} else {
		return nil, errors.New(`{ "message": "invalid token" }`)
	}
}

var rootQuery *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"authors": &graphql.Field{
			Type: graphql.NewList(authorType),
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return authors, nil
			},
		},
		"author": &graphql.Field{
			Type: authorType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id := params.Args["id"].(string)
				for _, author := range authors {
					if author.Id == id {
						return author, nil
					}
				}
				return nil, nil
			},
		},
		"articles": &graphql.Field{
			Type: graphql.NewList(articleType),
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return articles, nil
			},
		},
		"article": &graphql.Field{
			Type: articleType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id := params.Args["id"].(string)
				for _, article := range authors {
					if article.Id == id {
						return article, nil
					}
				}
				return nil, nil
			},
		},
	},
})

var rootMutation *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		"deleteAutor": &graphql.Field{
			Type: graphql.NewList(authorType),
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id := params.Args["id"].(string)

				for index, author := range authors {
					if author.Id == id {
						authors = append(authors[:index], authors[index+1:]...)

						return authors, nil
					}
				}
				return nil, errors.New("not found")
			},
		},
		"updateAutor": &graphql.Field{
			Type: graphql.NewList(authorType),
			Args: graphql.FieldConfigArgument{
				"author": &graphql.ArgumentConfig{
					Type: authorInputType,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				var changes Author
				mapstructure.Decode(params.Args["author"], &changes)
				validate := validator.New()
				for index, author := range authors {
					if author.Id == changes.Id {
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
							err := validate.Var(changes.Password, "gte=4")
							if err != nil {
								return nil, err
							}
							hash, _ := bcrypt.GenerateFromPassword([]byte(changes.Password), 10)
							author.Password = string(hash)
						}

						authors[index] = author

						return authors, nil
					}
				}
				return nil, errors.New("not found")
			},
		},
		"createArticle": &graphql.Field{
			Type: graphql.NewList(articleType),
			Args: graphql.FieldConfigArgument{
				"article": &graphql.ArgumentConfig{
					Type: articleInputType,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				var article Article
				mapstructure.Decode(params.Args["article"], &article)
				decoded, err := ValidateJWT(params.Context.Value("token").(string))
				if err != nil {
					return nil, err
				}
				validate := validator.New()
				err = validate.Struct(article)
				if err != nil {
					return nil, err
				}
				article.Id = uuid.NewV4().String()
				article.Author = decoded.(CustomJWTClaims).Id
				articles = append(articles, article)

				return articles, nil
			},
		},
		"deleteArticle": &graphql.Field{
			Type: graphql.NewList(articleType),
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id := params.Args["id"].(string)

				for index, article := range articles {
					if article.Id == id {
						articles = append(articles[:index], articles[index+1:]...)

						return articles, nil
					}
				}
				return nil, errors.New("not found")
			},
		},
		"updateArticle": &graphql.Field{
			Type: graphql.NewList(articleType),
			Args: graphql.FieldConfigArgument{
				"article": &graphql.ArgumentConfig{
					Type: articleInputType,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				var changes Article
				mapstructure.Decode(params.Args["article"], &changes)
				for index, article := range articles {
					if article.Id == changes.Id {
						if changes.Title != "" {
							article.Title = changes.Title
						}
						if changes.Content != "" {
							article.Content = changes.Content
						}

						articles[index] = article

						return articles, nil
					}
				}
				return nil, errors.New("not found")
			},
		},
	},
})

func main() {
	fmt.Println("Starting app...")
	router := mux.NewRouter()
	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
	router.HandleFunc("/graphql", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("content-type", "application/json")
		var payload GraphQLPayload
		json.NewDecoder(req.Body).Decode(&payload)
		result := graphql.Do(graphql.Params{
			Schema:         schema,
			RequestString:  payload.Query,
			VariableValues: payload.Variables,
			Context:        context.WithValue(context.Background(), "token", req.URL.Query().Get("token")),
		})
		json.NewEncoder(res).Encode(result)
	})
	router.HandleFunc("/register", RegisterEndpoint).Methods("POST")
	router.HandleFunc("/login", LoginEndpoint).Methods("POST")
	headers := handlers.AllowedHeaders(
		[]string{
			"Content-Type",
			"Aouthorization",
			"X-Requested-With",
		},
	)
	methods := handlers.AllowedMethods([]string{
		"POST", "PUT", "DELETE", "GET",
	})
	origins := handlers.AllowedOrigins([]string{"*"})
	http.ListenAndServe(":8080", handlers.CORS(headers, methods, origins)(router))
}
