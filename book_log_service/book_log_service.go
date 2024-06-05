// book_log_service.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
)

var client *mongo.Client

func getBooks(w http.ResponseWriter, r *http.Request) {
	collection := client.Database("bookstore").Collection("books")
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var books []bson.M
	if err = cursor.All(context.Background(), &books); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, books)
}

func getLogs(w http.ResponseWriter, r *http.Request) {
	collection := client.Database("bookstore").Collection("logs")
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var logs []bson.M
	if err = cursor.All(context.Background(), &logs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, logs)
}

func main() {
	var err error
	clientOptions := options.Client().ApplyURI("mongodb+srv://Esimgali:kuxeP8FmpY80Sj9g@cluster0.7lkmz1b.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0")
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	r := mux.NewRouter()
	r.HandleFunc("/books", getBooks).Methods("GET")
	r.HandleFunc("/logs", getLogs).Methods("GET")

	http.Handle("/", r)
	http.ListenAndServe(":8081", nil)
}
