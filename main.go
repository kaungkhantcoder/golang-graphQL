package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

type Blog struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

var db *sql.DB

// establish a connection to the database

func initDB() {
	var err error
	db, err = sql.Open("mysql", "blog_user:db_password@tcp(127.0.0.1:3306)/my_graphql_blog_db?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}
}

// Create GraphQL types

func createBlogType() *graphql.Object {
	return graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Blog",
			Fields: graphql.Fields{
				"id": &graphql.Field{
					Type: graphql.Int,
				},
				"title": &graphql.Field{
					Type: graphql.String,
				},
				"content": &graphql.Field{
					Type: graphql.String,
				},
			},
		},
	)
}

// Define the query type for the GraphQL server

func queryType(blogType *graphql.Object) *graphql.Object {
	return graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"blogs": &graphql.Field{
					Type: graphql.NewList(blogType),
					Args: graphql.FieldConfigArgument{
						"limit": &graphql.ArgumentConfig{
							Type: graphql.Int,
						},
						"offset": &graphql.ArgumentConfig{
							Type: graphql.Int,
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						// Read limit
						limit, _ := p.Args["limit"].(int)
						if limit <= 0 || limit > 20{
							limit = 10
						}
						// Read offset
						offset, _ := p.Args["limit"].(int)
						if offset < 0{
							offset = 0
						}

						return getBlogs(limit, offset)
					},
				},
			},
		},
	)
}

func getBlogs(limit int, offset int) ([]Blog, error) {
	var blogs []Blog
	rows, err := db.Query("SELECT id, title, content FROM blogs limit " + strconv.Itoa(limit) + " offset " +strconv.Itoa(offset))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var b Blog
		if err := rows.Scan(&b.ID, &b.Title, &b.Content); err != nil {
			return nil, err
		}
		blogs = append(blogs, b)
	}
	return blogs, nil
}

func main() {
	initDB()

	blogType := createBlogType()

	schema, err := graphql.NewSchema(
		graphql.SchemaConfig{
			Query: queryType(blogType),
		},
	)

	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	handler := handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
	})

	http.Handle("/graphql", handler)

	const port = ":8080"
	fmt.Printf("Server is running on port %s...\n", port)

	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
