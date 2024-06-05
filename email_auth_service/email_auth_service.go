package main

import (
	"context"
	"fmt"
	"net/http"
	"net/smtp"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
)

var client *mongo.Client

func sendAuthCode(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	code := r.URL.Query().Get("code")

	from := "esimgalikhamitov2005@gmail.com"
	password := "oauc fsxn vnxd paxx"
	SMTPHost := "smtp.gmail.com"
	port := 587
	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: Verification Code\n\nYour verification code is: %s", from, email, code)

	auth := smtp.PlainAuth("", from, password, SMTPHost)

	errLast := smtp.SendMail(fmt.Sprintf("%s:%d", SMTPHost, port), auth, from, []string{email}, []byte(msg))
	if errLast != nil {
		http.Error(w, errLast.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Email Sent Successfully!")
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
	r.HandleFunc("/sendAuthCode", sendAuthCode).Methods("GET")

	http.Handle("/", r)
	http.ListenAndServe(":8082", nil)
}
