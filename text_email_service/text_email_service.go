// text_email_service.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/gomail.v2"
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

	m := gomail.NewMessage()
	m.SetHeader("From", "esimgalikhamitov2005@gmail.com")
	m.SetHeader("To", string(email))
	m.SetHeader("Subject", "Hello!")
	m.SetBody("text/plain", "This is the plain text body of the email.")
	m.Attach("./file.txt")
	d := gomail.NewDialer("smtp.gmail.com", 587, "esimgalikhamitov2005@gmail.com", "ydbq okrv hvpq tcqc")
	if err := d.DialAndSend(m); err != nil {
		panic(err)
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
	r.HandleFunc("/sendTextAsEmail", sendTextAsEmail).Methods("GET")

	http.Handle("/", r)
	http.ListenAndServe(":8083", nil)
}
