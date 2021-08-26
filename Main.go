package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

var (
	err          = godotenv.Load("config.env")
	DB_USER_NAME = os.Getenv("DB_USER_NAME")
	DB_PASS_WORD = os.Getenv("DB_PASS_WORD")
	username     = os.Getenv("USER_NAME")
	password     = os.Getenv("PASS_WORD")
)

type Email struct {
	Id               int       `json:"id"`
	From             string    `json:"sender"`
	To               string    `json:"recipient"`
	Subject          string    `json:"subject"`
	Body             string    `json:"body"`
	Attachment       string    `json:"attachment_path"`
	Template         string    `json:"template_path"`
	SentTime         time.Time `json:"sent_time"`
	ScheduleSentTime time.Time `json:"schedule_sent_time"`
	Status           string    `json:"status"`
}

func main() {
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := mux.NewRouter()
	r.HandleFunc("/send", SendHandler)
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

func SendHandler(w http.ResponseWriter, r *http.Request) {

	// Parse our multipart form, 5 << 20 specifies a maximum
	// upload of 5 MB files.
	r.ParseMultipartForm(5 << 20)
	file, handler, err := r.FormFile("attactment")
	attachment := ""
	if err != nil {
		fmt.Println("This email has no Acttachments")
	} else {
		defer file.Close()

		if handler.Size > (5 << 20) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "The maximum size of attachments is 5MB")
			return
		}

		ext := handler.Filename[len(handler.Filename)-4:]
		// Create a temporary file within our uploads directory that follows
		// a particular naming pattern
		tempFile, err := ioutil.TempFile("uploads", "upload-*"+ext)
		if err != nil {
			fmt.Println(err)
		}
		defer tempFile.Close()

		// read all of the contents of our uploaded file into a
		// byte array
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println(err)
		}
		// write this byte array to our temporary file
		tempFile.Write(fileBytes)
		// return that we have successfully uploaded our file!
		attachment = tempFile.Name()
	}

	r.ParseForm()          // Parses the request body
	to := r.Form.Get("to") // x will be "" if parameter is not set
	subject := r.Form.Get("subject")
	body := r.Form.Get("body")
	from := username + "@gmail.com"
	sentEmail(from, strings.Split(to, ";"), subject, body, attachment, "")
	fmt.Fprintf(w, "The email has been sent successfully!")
}

func sentEmail(from string, to []string, subject string, body string, attachmentPath string, template string) {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	if attachmentPath != "" {
		m.Attach(attachmentPath)
	}

	d := gomail.NewDialer("smtp.gmail.com", 587, username, password)

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		saveToDB(from, to, subject, body, attachmentPath, template, time.Now(), "", "ERROR")
		panic(err)
	}
	saveToDB(from, to, subject, body, attachmentPath, template, time.Now(), "", "SENT")
}

func saveToDB(from string, to []string, subject string, body string, attachmentPath string, template string, sentTime time.Time, scheduleSendTime string, status string) {
	db, err := sql.Open("mysql", DB_USER_NAME+":"+DB_PASS_WORD+"@/testdb?parseTime=true")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	insert, err := db.Query(fmt.Sprintf("INSERT INTO emails VALUES ( 0,  '%v', '%v', '%v', '%v', '%v', '%v', '%v', '%v', '%v');", from, convertArrayToString(to), subject, body, attachmentPath, template, sentTime, scheduleSendTime, status))
	if err != nil {
		panic(err.Error())
	}
	// be careful deferring Queries if you are using transactions
	defer insert.Close()
}

func getAll() []Email {
	db, err := sql.Open("mysql", DB_USER_NAME+":"+DB_PASS_WORD+"@/testdb?parseTime=true")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	emails := []Email{}
	results, err := db.Query("SELECT * FROM emails")
	if err != nil {
		panic(err.Error())
	}
	for results.Next() {
		var email Email
		err = results.Scan(&email.Id, &email.From, &email.To, &email.Subject, &email.Body, &email.Attachment, &email.Template, &email.SentTime, &email.ScheduleSentTime, &email.Status)
		if err != nil {
			panic(err.Error())
		}
		emails = append(emails, email)
	}
	return emails
}

func convertArrayToString(emails []string) string {
	result := emails[0]
	for i := 1; i < len(emails); i++ {
		result += ";" + emails[i]
	}
	return result
}
