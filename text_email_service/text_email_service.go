// text_email_service.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/smtp"
	"os"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func sendTextAsEmail(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	text := r.URL.Query().Get("text")

	fileName := "file.txt"
	file, err := os.Create(fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	_, err = file.WriteString(text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	from := "your-email@example.com"
	password := "your-email-password"

	to := []string{email}
	smtpHost := "smtp.example.com"
	smtpPort := "587"

	message := []byte("Subject: Your Text File\r\n\r\n" + "Please find the attached text file.")

	auth := smtp.PlainAuth("", from, password, smtpHost)
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
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
	r.HandleFunc("/sendTextAsEmail", sendTextAsEmail).Methods("GET")

	http.Handle("/", r)
	http.ListenAndServe(":8083", nil)
}
