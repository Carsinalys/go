package main

import (
	"os"
	"reflect"
	"testing"

	"test/quotes"
)

func TestApp_createQuote(t *testing.T) {
	tests := []struct {
		name    string
		quotes  quotes.Quote
		wantErr bool
	}{
		{"Alfred", quotes.Quote{Author: "Alfred E. Neuman", Text: "What, me worry?", Source: "MAD Magazine"}, false},
		{"Alfred again", quotes.Quote{Author: "Alfred E. Neuman again", Text: "What, me worry?", Source: "MAD Magazine"}, false},
	}
	db, err := quotes.Open("testdb")
	if err != nil {
		t.Fatalf("Cannot create test DB")
	}
	app := &App{db: *db}
	defer func() {
		app.db.Close()
		os.Remove("testdb")
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := app.db.Create(&tt.quotes); (err != nil) != tt.wantErr {
				t.Errorf("app.db.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	data, err := app.db.List()

	if err != nil {
		t.Errorf("app.db.List() error = %v", err)
	}

	if len(data) != 2 {
		t.Errorf("app.db.List() len() is not eqyal 2, data is = %v", data)
	}
}

func TestApp_getQuote(t *testing.T) {
	tests := []struct {
		name    string
		author  string
		want    *quotes.Quote
		wantErr bool
	}{
		{"Alfred", "Alfred E. Neuman", &quotes.Quote{
			Author: "Alfred E. Neuman",
			Text:   "What, me worry?",
			Source: "MAD Magazine",
		},
			false},
	}
	db, err := quotes.Open("testdb")
	if err != nil {
		t.Fatalf("Cannot create test DB")
	}
	app := &App{db: *db}
	defer func() {
		app.db.Close()
		os.Remove("testdb")
	}()
	app.db.Create(
		&quotes.Quote{
			Author: "Alfred E. Neuman",
			Text:   "What, me worry?",
			Source: "MAD Magazine",
		})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := app.db.Get(tt.author)
			if (err != nil) != tt.wantErr {
				t.Errorf("app.db.Get() error = %s, wantErr %t", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("app.db.Get() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
