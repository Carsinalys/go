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

var (
	urlExample = "postgres://cardinalys:cardinalys@localhost:5432/godb"
	mockAuthor = Author{
		FirstName: "xyz",
		LastName:  "pqr",
		UserName:  "kjhab",
		Password:  "1234567890",
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
	fmt.Printf("===>>>1 %v\n", newAuthor)
	mutation := `"query": "{mutation { updateAuthor(author: { id: "` + newAuthor.Id + `" firstname: "` + newAuthor.FirstName + `" lastname: "` + newAuthor.LastName + `" username: "` + newAuthor.UserName + `" password: "` + newAuthor.Password + `" }) { id, firstname, lastname, username, password }}}"`
	fmt.Printf("===>>>2 %v\n", bytes.NewBuffer([]byte(mutation)))
	req, err := http.NewRequest("POST", "/graphql", bytes.NewBuffer([]byte(mutation)))
	addCookie(req)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GraphqlHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	var data Author
	json.NewDecoder(rr.Body).Decode(&data)
	fmt.Printf("===>>>3 %v\n", data)
	defer DBConnection.Close(context.Background())
}

func TestClearTable(t *testing.T) {
	connectDB()

	query := fmt.Sprintf("delete from authors where id='%v'", dbAuthor.Id)
	_, err := DBConnection.Exec(context.Background(), query)
	if err != nil {
		t.Errorf("Unable to delete record from database: %v\n", err)
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
