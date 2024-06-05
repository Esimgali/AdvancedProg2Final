package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// Function to generate a random code
func generateCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Handler for user registration
func register(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		password := r.URL.Query().Get("password")
		email := r.URL.Query().Get("email")
		admin := r.URL.Query().Get("admin")

		if name == "" || password == "" || email == "" || admin == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		adminStatus := false
		if admin == "true" {
			adminStatus = true
		}

		collection := client.Database("bookstore").Collection("clients")
		_, err = collection.InsertOne(context.Background(), bson.M{
			"name":           name,
			"password":       string(hashedPassword),
			"mail":           email,
			"admin":          adminStatus,
			"token":          "",
			"token_for_code": "",
			"code":           "",
		})
		if err != nil {
			http.Error(w, "Failed to register user", http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, "User registered successfully")
	}
}

// Handler for user login
func login(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		fmt.Fprint(w, bson.M{
			"text":           "Auth code sent to email",
			"token_for_code": tokenForCode,
		})
	}
}

// Handler for checking the authentication code
func checkCode(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		tokenForCode := r.URL.Query().Get("token_for_code")

		collection := client.Database("bookstore").Collection("clients")
		var user bson.M
		err := collection.FindOne(context.Background(), bson.M{"code": code, "token_for_code": tokenForCode}).Decode(&user)
		if err != nil {
			http.Error(w, "Invalid token for code", http.StatusUnauthorized)
			return
		}

		token := generateCode()

		_, err = collection.UpdateOne(context.Background(), bson.M{"code": code}, bson.M{
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
}

// Handler for sending book list to email
func toEmail(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
}
func getBooks(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get("http://localhost:8081/books")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}
}

func main() {
	clientOptions := options.Client().ApplyURI("mongodb+srv://Esimgali:kuxeP8FmpY80Sj9g@cluster0.7lkmz1b.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0")
	//mongodb+srv://Esimgali:<password>@cluster0.7lkmz1b.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		fmt.Println(err)
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/register", register(client)).Methods("POST")
	r.HandleFunc("/login", login(client)).Methods("GET")
	r.HandleFunc("/checkCode", checkCode(client)).Methods("GET")
	r.HandleFunc("/toEmail", toEmail(client)).Methods("GET")
	r.HandleFunc("/books", getBooks(client)).Methods("GET")

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
