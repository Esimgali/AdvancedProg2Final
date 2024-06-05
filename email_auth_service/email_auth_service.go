// email_auth_service.go
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

	from := "your-email@example.com"
	password := "your-email-password"

	to := []string{email}
	smtpHost := "smtp.example.com"
	smtpPort := "587"

	message := []byte("Subject: Authentication Code\r\n\r\n" + "Your authentication code is: " + code)

	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Email Sent Successfully!")
}

func main() {
	var err error
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
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
