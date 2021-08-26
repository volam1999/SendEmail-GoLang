package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	err      = godotenv.Load("config.env")
	dbUser   = os.Getenv("DB_USER_NAME")
	dbPass   = os.Getenv("DB_PASS_WORD")
	dbName   = os.Getenv("DB_NAME")
	username = os.Getenv("USER_NAME")
	password = os.Getenv("PASS_WORD")
)

type Email struct {
	gorm.Model
	Id               int
	From             string
	To               string
	Subject          string
	Body             string
	Attachment       string
	Template         string
	SentTime         time.Time
	ScheduleSentTime time.Time
	Status           string
}

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "02/01/2006 15:04",
		FullTimestamp:   true,
	})
	log.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
}

func main() {
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}
	go checkScheduleEmail()
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

	r.ParseForm() // Parses the request body
	from := username + "@gmail.com"
	to := r.Form.Get("to") // x will be "" if parameter is not set
	subject := r.Form.Get("subject")
	body := r.Form.Get("body")
	schedule := r.Form.Get("schedule")
	scheduleTime := time.Time{}
	if schedule != "" {
		layout := "02-01-2006 15:04"
		var err error
		scheduleTime, err = time.Parse(layout, schedule)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Schedule time format expected: dd-MM-yyyy hh:mm")
			log.Error("Schedule time wrong format: " + schedule)
			return
		}

		saveToDB(from, to, subject, body, attachment, "", time.Time{}, scheduleTime.Add(-time.Hour*7), "PENDING")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "The email has been successfully saved and will automatically be sent when the time comes.")
		return
	}

	if !sentEmail(from, strings.Split(to, ";"), subject, body, attachment, "") {
		saveToDB(from, to, subject, body, attachment, "", time.Time{}, time.Time{}, "ERROR")
	} else {
		saveToDB(from, to, subject, body, attachment, "", time.Now(), time.Time{}, "SENT")
	}

	fmt.Fprintf(w, "The email has been sent successfully!")
}

func sentEmail(from string, to []string, subject string, body string, atm string, template string) bool {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	if atm != "" {
		m.Attach(atm)
	}

	d := gomail.NewDialer("smtp.gmail.com", 587, username, password)

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		return false
	}
	return true
}

func saveToDB(from string, to string, subject string, body string, atm string, template string, sentTime time.Time, scheduleSendTime time.Time, status string) {
	dsn := "root:@/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database")
		panic("Failed to connect database")
	}
	// Migrate the schema
	db.AutoMigrate(&Email{})
	// Create
	db.Create(&Email{From: from, To: to, Subject: subject, Body: body, Attachment: atm, Template: template, SentTime: sentTime, ScheduleSentTime: scheduleSendTime, Status: status})
}

func convertArrayToString(emails []string) string {
	result := emails[0]
	for i := 1; i < len(emails); i++ {
		result += ";" + emails[i]
	}
	return result
}

func checkScheduleEmail() {
	for {
		dsn := "root:@/testdb?charset=utf8mb4&parseTime=True&loc=Local"
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect database")
			panic("Failed to connect database")
		}
		var emails []Email
		db.Where("status = ?", "PENDING").Find(&emails)
		for _, email := range emails {
			if email.ScheduleSentTime.Before(time.Now()) {
				if sentEmail(email.From, strings.Split(email.To, ";"), email.Subject, email.Body, email.Attachment, email.Template) {
					db.Model(&email).Updates(Email{SentTime: time.Now(), Status: "SENT"})
					log.Info("Schedule mail has been sent!")
				} else {
					db.Model(&email).Updates(Email{Status: "ERROR"})
					log.Warn("Schedule mail sent failed!")
				}
			}
		}
		time.Sleep(time.Second * 20)
	}
}
