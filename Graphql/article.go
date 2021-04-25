package main

import (
	"context"
	"fmt"
	"os"

	"github.com/graphql-go/graphql"
)

type Article struct {
	Id      string `json:"id,omitempty" validate:"omitempty,uuid"`
	Author  string `json:"author,omitempty" validate:"isdefault"`
	Title   string `json:"title,omitempty" validate:"required"`
	Content string `json:"content,omitempty" validate:"required"`
}

var articleType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Article",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"author": &graphql.Field{
			Type: authorType,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				article := params.Source.(Article)
				var userDB Author
				query := fmt.Sprintf("select * from authors where id='%v'", article.Author)
				error := DBConnection.QueryRow(context.Background(), query).Scan(&userDB.Id, &userDB.FirstName, &userDB.LastName, &userDB.UserName, &userDB.Password)
				if error != nil {
					fmt.Fprintf(os.Stderr, "Unable to find user in database: %v\n", error)
					return nil, error
				}
				return userDB, nil
			},
		},
		"title": &graphql.Field{
			Type: graphql.String,
		},
		"content": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var articleInputType *graphql.InputObject = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "ArticleInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"title": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"content": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})
