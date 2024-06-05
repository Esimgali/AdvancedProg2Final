package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/mux"
)

var client *mongo.Client

func generateCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func login(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	password := r.URL.Query().Get("password")

	collection := client.Database("bookstore").Collection("clients")
	var user bson.M
	err := collection.FindOne(context.Background(), bson.M{"name": name}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user["password"].(string)), []byte(password))
	if err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	code := generateCode()
	tokenForCode := generateCode()

	_, err = collection.UpdateOne(context.Background(), bson.M{"name": name}, bson.M{
		"$set": bson.M{
			"code":           code,
			"token_for_code": tokenForCode,
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.Get(fmt.Sprintf("http://localhost:8082/sendAuthCode?email=%s&code=%s", user["mail"], code))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	fmt.Fprint(w, "Auth code sent to email")
}

func checkCode(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	tokenForCode := r.URL.Query().Get("token_for_code")

	collection := client.Database("bookstore").Collection("clients")
	var user bson.M
	err := collection.FindOne(context.Background(), bson.M{"name": name, "token_for_code": tokenForCode}).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid token for code", http.StatusUnauthorized)
		return
	}

	token := generateCode()

	_, err = collection.UpdateOne(context.Background(), bson.M{"name": name}, bson.M{
		"$set": bson.M{
			"token": token,
		},
		"$unset": bson.M{
			"code":           "",
			"token_for_code": "",
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, token)
}

func toEmail(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")

	collection := client.Database("bookstore").Collection("clients")
	var user bson.M
	err := collection.FindOne(context.Background(), bson.M{"token": token}).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	resp, err := http.Get("http://localhost:8081/books")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var books string
	fmt.Fscan(resp.Body, &books)

	resp, err = http.Get(fmt.Sprintf("http://localhost:8083/sendTextAsEmail?email=%s&text=%s", user["mail"], books))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	fmt.Fprint(w, "Books sent to email")
}

func main() {
	var err error
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		fmt.Println(err)
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/login", login).Methods("GET")
	r.HandleFunc("/checkCode", checkCode).Methods("GET")
	r.HandleFunc("/toEmail", toEmail).Methods("GET")

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
