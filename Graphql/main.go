package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"github.com/jackc/pgx/v4"
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
	JWT_SECRET   []byte = []byte("some strong key")
	DBConnection *pgx.Conn
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
				rows, err := DBConnection.Query(context.Background(), "select * from authors")
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to get users from database: %v\n", err)
					return err, nil
				}
				defer rows.Close()
				var result []Author
				for rows.Next() {
					var r Author
					err = rows.Scan(&r.Id, &r.FirstName, &r.LastName, &r.UserName, &r.Password)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Unable to scan %v\n", err)
						return err, nil
					}
					result = append(result, r)
				}
				return result, nil
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
				var result Author
				query := fmt.Sprintf("select * from authors where id='%v'", id)
				err := DBConnection.QueryRow(context.Background(), query).Scan(&result.Id, &result.FirstName, &result.LastName, &result.UserName, &result.Password)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to find user in database: %v\n", err)
					return nil, err
				}
				return result, nil
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
		"deleteAuthor": &graphql.Field{
			Type: authorType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id := params.Args["id"].(string)
				query := fmt.Sprintf("delete from authors where id='%v'", id)
				_, err := DBConnection.Exec(context.Background(), query)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to delete record from database: %v\n", err)
					return nil, err
				}
				return Author{Id: id}, nil
			},
		},
		"updateAuthor": &graphql.Field{
			Type: authorType,
			Args: graphql.FieldConfigArgument{
				"author": &graphql.ArgumentConfig{
					Type: authorInputType,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				var changes Author
				mapstructure.Decode(params.Args["author"], &changes)
				validate := validator.New()
				var dbAuthor Author
				query := fmt.Sprintf("select * from authors where id='%v'", changes.Id)
				err := DBConnection.QueryRow(context.Background(), query).Scan(&dbAuthor.Id, &dbAuthor.FirstName, &dbAuthor.LastName, &dbAuthor.UserName, &dbAuthor.Password)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to find user in database: %v\n", err)
					return nil, err
				}
				if changes.FirstName != "" {
					dbAuthor.FirstName = changes.FirstName
				}
				if changes.LastName != "" {
					dbAuthor.LastName = changes.LastName
				}
				if changes.UserName != "" {
					dbAuthor.UserName = changes.UserName
				}
				if changes.Password != "" {
					err := validate.Var(changes.Password, "gte=4")
					if err != nil {
						return nil, err
					}
					hash, _ := bcrypt.GenerateFromPassword([]byte(changes.Password), 10)
					dbAuthor.Password = string(hash)
				}
				query = fmt.Sprintf("update authors set firstname='%v', lastname='%v', username='%v', password='%v' where id='%v'", dbAuthor.FirstName, dbAuthor.LastName, dbAuthor.UserName, dbAuthor.Password, dbAuthor.Id)
				_, error := DBConnection.Exec(context.Background(), query)
				if error != nil {
					fmt.Fprintf(os.Stderr, "Unable to insert record to database: %v\n", error)
					return nil, error
				}

				return dbAuthor, nil
			},
		},
		"createArticle": &graphql.Field{
			Type: articleType,
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
				_, error := DBConnection.Exec(context.Background(), "insert into articles(id, author, title, content) values($1, $2, $3, $4)", article.Id, article.Author, article.Title, article.Content)
				if error != nil {
					fmt.Fprintf(os.Stderr, "Unable to insert record to database: %v\n", error)
					return nil, error
				}
				return article, nil
			},
		},
		"deleteArticle": &graphql.Field{
			Type: articleType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id := params.Args["id"].(string)
				query := fmt.Sprintf("delete from articles where id='%v'", id)
				_, err := DBConnection.Exec(context.Background(), query)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to delete record from database: %v\n", err)
					return nil, err
				}
				return Article{Id: id}, nil
			},
		},
		"updateArticle": &graphql.Field{
			Type: articleType,
			Args: graphql.FieldConfigArgument{
				"article": &graphql.ArgumentConfig{
					Type: articleInputType,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				var changes Article
				mapstructure.Decode(params.Args["article"], &changes)
				var dbArticle Article
				query := fmt.Sprintf("select * from articles where id='%v'", changes.Id)
				err := DBConnection.QueryRow(context.Background(), query).Scan(&dbArticle.Id, &dbArticle.Author, &dbArticle.Title, &dbArticle.Content)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to find article in database: %v\n", err)
					return nil, err
				}
				if changes.Title != "" {
					dbArticle.Title = changes.Title
				}
				if changes.Content != "" {
					dbArticle.Content = changes.Content
				}
				query = fmt.Sprintf("update articles set title='%v', content='%v' where id='%v'", dbArticle.Title, dbArticle.Content, dbArticle.Id)
				_, error := DBConnection.Exec(context.Background(), query)
				if error != nil {
					fmt.Fprintf(os.Stderr, "Unable to insert record to database: %v\n", error)
					return nil, error
				}

				return dbArticle, nil
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

	urlExample := "postgres://cardinalys:cardinalys@localhost:5432/godb"
	conn, err := pgx.Connect(context.Background(), urlExample)
	DBConnection = conn
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer DBConnection.Close(context.Background())
	http.ListenAndServe(":8080", handlers.CORS(headers, methods, origins)(router))
}
