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
				rows, err := DBConnection.Query(context.Background(), "select * from articles")
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to get articles from database: %v\n", err)
					return err, nil
				}
				defer rows.Close()
				var result []Article
				for rows.Next() {
					var r Article
					err = rows.Scan(&r.Id, &r.Author, &r.Title, &r.Content)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Unable to scan %v\n", err)
						return err, nil
					}
					result = append(result, r)
				}
				return result, nil
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
				var result Article
				query := fmt.Sprintf("select * from articles where id='%v'", id)
				err := DBConnection.QueryRow(context.Background(), query).Scan(&result.Id, &result.Author, &result.Title, &result.Content)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to find article in database: %v\n", err)
					return nil, err
				}
				return result, nil
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
				_, err := ValidateJWT(params.Context.Value("token").(string))
				if err != nil {
					return nil, err
				}
				id := params.Args["id"].(string)
				query := fmt.Sprintf("delete from authors where id='%v'", id)
				_, err = DBConnection.Exec(context.Background(), query)
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
				decoded, err := ValidateJWT(params.Context.Value("token").(string))
				if err != nil {
					return nil, err
				}
				var changes Author
				mapstructure.Decode(params.Args["author"], &changes)
				validate := validator.New()
				var dbAuthor Author
				query := fmt.Sprintf("select * from authors where id='%v'", decoded.(CustomJWTClaims).Id)
				err = DBConnection.QueryRow(context.Background(), query).Scan(&dbAuthor.Id, &dbAuthor.FirstName, &dbAuthor.LastName, &dbAuthor.UserName, &dbAuthor.Password)
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
				_, err := ValidateJWT(params.Context.Value("token").(string))
				if err != nil {
					return nil, err
				}
				id := params.Args["id"].(string)
				query := fmt.Sprintf("delete from articles where id='%v'", id)
				_, err = DBConnection.Exec(context.Background(), query)
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
				_, err := ValidateJWT(params.Context.Value("token").(string))
				if err != nil {
					return nil, err
				}
				var changes Article
				mapstructure.Decode(params.Args["article"], &changes)
				var dbArticle Article
				query := fmt.Sprintf("select * from articles where id='%v'", changes.Id)
				err = DBConnection.QueryRow(context.Background(), query).Scan(&dbArticle.Id, &dbArticle.Author, &dbArticle.Title, &dbArticle.Content)
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
	router.HandleFunc("/graphql", GraphqlHandler).Methods("POST")
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

func GraphqlHandler(res http.ResponseWriter, req *http.Request) {
	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
	fmt.Printf("===========1 %v\n", schema)
	res.Header().Set("content-type", "application/json")
	var payload GraphQLPayload
	fmt.Printf("===========2-1 %v\n", req.Body)
	json.NewDecoder(req.Body).Decode(&payload)
	fmt.Printf("===========2 %v\n", payload)
	result := graphql.Do(graphql.Params{
		Schema:         schema,
		RequestString:  payload.Query,
		VariableValues: payload.Variables,
		Context:        context.WithValue(context.Background(), "token", getToken(req.Cookies())),
	})
	fmt.Printf("===========3 %v\n", result)
	json.NewEncoder(res).Encode(result)
}

func getToken(cookies []*http.Cookie) string {
	token := ""
	for _, c := range cookies {
		if c.Name == "jwt" {
			token = c.Value
		}
	}

	return token
}
