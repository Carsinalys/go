package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgx/v4"
)

// this test testing all api flow of CRUD for author and article
type GraphqlUpdateAuthor struct {
	Data struct {
		UpdateAuthor struct {
			Author
		} `json:"updateAuthor"`
	} `json:"data"`
}

type GraphqlDeleteAuthor struct {
	Data struct {
		DeleteAuthor struct {
			Author
		} `json:"deleteAuthor"`
	} `json:"data"`
}

type GraphqlCreateArticle struct {
	Data struct {
		CreateArticle struct {
			Author struct{ Author }
			Article
		} `json:"createArticle"`
	} `json:"data"`
}

type GraphqlGetArticle struct {
	Data struct {
		Article struct {
			Author struct{ Author }
			Article
		} `json:"article"`
	} `json:"data"`
}

type GraphqlGetAuthor struct {
	Data struct {
		Author struct {
			Author
		} `json:"author"`
	} `json:"data"`
}

type GraphqlGetArticles struct {
	Data struct {
		Articles []struct {
			Author struct{ Author }
			Article
		} `json:"articles"`
	} `json:"data"`
}

type GraphqlGetAuthors struct {
	Data struct {
		Authors []struct {
			Author
		} `json:"authors"`
	} `json:"data"`
}

type GraphqlUpdateArticle struct {
	Data struct {
		UpdateArticle struct {
			Author struct{ Author }
			Article
		} `json:"updateArticle"`
	} `json:"data"`
}

var (
	urlExample = "postgres://cardinalys:cardinalys@localhost:5432/godb"
	mockAuthor = Author{
		FirstName: "xyz",
		LastName:  "pqr",
		UserName:  "kjhab",
		Password:  "1234567890",
	}
	mockArticle = Article{
		Title:   "test title",
		Content: "test content",
	}
	dbAuthor Author
	token    string
)

func connectDB() {
	conn, err := pgx.Connect(context.Background(), urlExample)
	DBConnection = conn
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
}
func TestJWTValidation(t *testing.T) {
	claims := CustomJWTClaims{
		Id: "c5eecc30-084d-47f1-99a4-bf5dd1b6498f",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour).Unix(),
			Issuer:    "something you can indetify",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JWT_SECRET)
	if err != nil {
		t.Fatalf("unable sing token string: %v", err)
	}

	_, error := ValidateJWT(tokenString)
	if error != nil {
		t.Fatalf("token is invalid: %v", error)
	}
}

func TestRegister(t *testing.T) {
	connectDB()
	buffer, err := json.Marshal(mockAuthor)
	if err != nil {
		t.Error(err)
		return
	}
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(buffer))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(RegisterEndpoint)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	var data Author
	json.NewDecoder(rr.Body).Decode(&data)
	dbAuthor = data
	if data.FirstName != "xyz" || data.LastName != "pqr" || data.UserName != "kjhab" {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), buffer)
	}

	defer DBConnection.Close(context.Background())
}

func TestLogin(t *testing.T) {
	connectDB()
	b, err := json.Marshal(mockAuthor)
	if err != nil {
		t.Error(err)
		return
	}
	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(b))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(LoginEndpoint)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	cookie := rr.Header().Get("Set-Cookie")
	if !strings.Contains(cookie, "jwt") {
		t.Errorf("handler returned unexpected cookie: got %v",
			cookie)
	}

	c := strings.Split(cookie, "=")
	dirtyToken := strings.Split(c[1], ";")
	token = dirtyToken[0]

	defer DBConnection.Close(context.Background())
}

func TestUserUpdate(t *testing.T) {
	connectDB()
	newAuthor := dbAuthor
	newAuthor.FirstName = "John Weak"
	postBody, error := json.Marshal(map[string]string{
		"query":         `mutation { updateAuthor(author: { id: "` + newAuthor.Id + `" firstname: "` + newAuthor.FirstName + `" lastname: "` + newAuthor.LastName + `" username: "` + newAuthor.UserName + `" password: "` + newAuthor.Password + `" }) { id, firstname, lastname, username, password }}`,
		"operationName": "updateAuthor",
	})
	if error != nil {
		t.Fatal(error)
	}
	body := bytes.NewBuffer(postBody)
	req, err := http.NewRequest("POST", "/graphql", body)
	req.Header.Add("Content-Type", "application/json")
	addCookie(req)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GraphqlHandler)
	handler.ServeHTTP(rr, req)
	var data GraphqlUpdateAuthor
	json.NewDecoder(rr.Body).Decode(&data)
	if data.Data.UpdateAuthor.FirstName != "John Weak" {
		t.Fatal("Unable to update author.")
	}
	defer DBConnection.Close(context.Background())
}

func TestGetUser(t *testing.T) {
	connectDB()
	postBody, error := json.Marshal(map[string]string{
		"query":         `query {author(id: "` + dbAuthor.Id + `") { firstname }}`,
		"operationName": "author",
	})
	if error != nil {
		t.Fatal(error)
	}
	body := bytes.NewBuffer(postBody)
	req, err := http.NewRequest("POST", "/graphql", body)
	req.Header.Add("Content-Type", "application/json")
	addCookie(req)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GraphqlHandler)
	handler.ServeHTTP(rr, req)
	var data GraphqlGetAuthor
	json.NewDecoder(rr.Body).Decode(&data)
	if data.Data.Author.FirstName != "John Weak" {
		t.Fatal("Unable to get user.")
	}
	defer DBConnection.Close(context.Background())
}

func TestGetUsers(t *testing.T) {
	connectDB()
	postBody, error := json.Marshal(map[string]string{
		"query":         `query {authors { id }}`,
		"operationName": "authors",
	})
	if error != nil {
		t.Fatal(error)
	}
	body := bytes.NewBuffer(postBody)
	req, err := http.NewRequest("POST", "/graphql", body)
	req.Header.Add("Content-Type", "application/json")
	addCookie(req)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GraphqlHandler)
	handler.ServeHTTP(rr, req)
	var data GraphqlGetAuthors
	json.NewDecoder(rr.Body).Decode(&data)
	if len(data.Data.Authors) != 1 {
		t.Fatal("Get all authors error expect 1!")
	}
	defer DBConnection.Close(context.Background())
}

func TestCreateArticle(t *testing.T) {
	connectDB()
	postBody, error := json.Marshal(map[string]string{
		"query":         `mutation { createArticle(article: { title: "` + mockArticle.Title + `" content: "` + mockArticle.Content + `"}) { id, title, content, author { id, firstname, lastname, username, password } }}`,
		"operationName": "createArticle",
	})
	if error != nil {
		t.Fatal(error)
	}
	body := bytes.NewBuffer(postBody)
	req, err := http.NewRequest("POST", "/graphql", body)
	req.Header.Add("Content-Type", "application/json")
	addCookie(req)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GraphqlHandler)
	handler.ServeHTTP(rr, req)
	var data GraphqlCreateArticle
	json.NewDecoder(rr.Body).Decode(&data)
	if data.Data.CreateArticle.Id == "" && data.Data.CreateArticle.Author.Id == "" {
		t.Fatal("Unable to create article.")
	}
	mockArticle = data.Data.CreateArticle.Article
	defer DBConnection.Close(context.Background())
}

func TestGetArticle(t *testing.T) {
	connectDB()
	postBody, error := json.Marshal(map[string]string{
		"query":         `query {article(id: "` + mockArticle.Id + `") { id, title, content, author { id, firstname, lastname, username, password }}}`,
		"operationName": "article",
	})
	if error != nil {
		t.Fatal(error)
	}
	body := bytes.NewBuffer(postBody)
	req, err := http.NewRequest("POST", "/graphql", body)
	req.Header.Add("Content-Type", "application/json")
	addCookie(req)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GraphqlHandler)
	handler.ServeHTTP(rr, req)
	var data GraphqlGetArticle
	json.NewDecoder(rr.Body).Decode(&data)
	if data.Data.Article.Title != "test title" || data.Data.Article.Content != "test content" {
		t.Fatal("Unable to get article.")
	}
	defer DBConnection.Close(context.Background())
}

func TestGetArticles(t *testing.T) {
	connectDB()
	postBody, error := json.Marshal(map[string]string{
		"query":         `query {articles { id, title, content, author { id, firstname, lastname, username, password }}}`,
		"operationName": "article",
	})
	if error != nil {
		t.Fatal(error)
	}
	body := bytes.NewBuffer(postBody)
	req, err := http.NewRequest("POST", "/graphql", body)
	req.Header.Add("Content-Type", "application/json")
	addCookie(req)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GraphqlHandler)
	handler.ServeHTTP(rr, req)
	var data GraphqlGetArticles
	json.NewDecoder(rr.Body).Decode(&data)
	if len(data.Data.Articles) != 1 {
		t.Fatal("Get all articles error expect 1!")
	}
	defer DBConnection.Close(context.Background())
}

func TestArticleUpdate(t *testing.T) {
	connectDB()
	newArticle := mockArticle
	newArticle.Title = "New Title!"
	postBody, error := json.Marshal(map[string]string{
		"query":         `mutation { updateArticle(article: { id: "` + newArticle.Id + `" title: "` + newArticle.Title + `" content: "` + newArticle.Content + `" }) { id, title, content, author { id, firstname, lastname, username, password } }}`,
		"operationName": "updateArticle",
	})
	if error != nil {
		t.Fatal(error)
	}
	body := bytes.NewBuffer(postBody)
	req, err := http.NewRequest("POST", "/graphql", body)
	req.Header.Add("Content-Type", "application/json")
	addCookie(req)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GraphqlHandler)
	handler.ServeHTTP(rr, req)
	var data GraphqlUpdateArticle
	json.NewDecoder(rr.Body).Decode(&data)
	if data.Data.UpdateArticle.Title != "New Title!" {
		t.Fatal("Unable to update article.")
	}
	defer DBConnection.Close(context.Background())
}

func TestArticleDelete(t *testing.T) {
	connectDB()
	postBody, error := json.Marshal(map[string]string{
		"query":         `mutation { deleteArticle(id: "` + mockArticle.Id + `") { id }}`,
		"operationName": "deleteArticle",
	})
	if error != nil {
		t.Fatal(error)
	}
	body := bytes.NewBuffer(postBody)
	req, err := http.NewRequest("POST", "/graphql", body)
	req.Header.Add("Content-Type", "application/json")
	addCookie(req)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GraphqlHandler)
	handler.ServeHTTP(rr, req)
	var result Article
	query := fmt.Sprintf("delete from articles where id='%v'", mockArticle.Id)
	err = DBConnection.QueryRow(context.Background(), query).Scan(&result.Id)
	if err == nil {
		t.Fatal(err)
	}
	defer DBConnection.Close(context.Background())
}

func TestUserDelete(t *testing.T) {
	connectDB()
	postBody, error := json.Marshal(map[string]string{
		"query":         `mutation { deleteAuthor(id: "` + dbAuthor.Id + `") { id }}`,
		"operationName": "deleteAuthor",
	})
	if error != nil {
		t.Fatal(error)
	}
	body := bytes.NewBuffer(postBody)
	req, err := http.NewRequest("POST", "/graphql", body)
	req.Header.Add("Content-Type", "application/json")
	addCookie(req)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GraphqlHandler)
	handler.ServeHTTP(rr, req)
	var result Author
	query := fmt.Sprintf("delete from authors where username='%v'", dbAuthor.UserName)
	err = DBConnection.QueryRow(context.Background(), query).Scan(&result.Id, &result.UserName, &result.FirstName, &result.LastName, &result.Password)
	if err == nil {
		t.Fatal(err)
	}
	defer DBConnection.Close(context.Background())
}

func addCookie(req *http.Request) {
	cookie := &http.Cookie{
		Name:  "jwt",
		Value: token,
		// Secure:  true,
		Expires: time.Now().Add(24 * time.Hour),
		// HttpOnly: true,
	}
	req.AddCookie(cookie)
}
