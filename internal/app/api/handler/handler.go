package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"main/internal/app/entity"
	"main/internal/pkg/db/mysqldb"
	"main/internal/pkg/email"
	"net/http"
	"os"
	"strings"
	"time"

	"main/internal/pkg/log"

	"github.com/gorilla/mux"
)

func GetHandler(w http.ResponseWriter, r *http.Request) {

	db := mysqldb.New()
	var emails []entity.Email
	db.Find(&emails)
	json, _ := json.Marshal(emails)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func GetByIdHanler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	db := mysqldb.New()

	var email entity.Email
	db.First(&email, id)
	json, _ := json.Marshal(email)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func SendHandler(w http.ResponseWriter, r *http.Request) {

	// Parse our multipart form, 5 << 20 specifies a maximum
	// upload of 5 MB files.
	r.ParseMultipartForm(5 << 20)
	file, handler, err := r.FormFile("attactment")
	attachment := ""
	if err != nil {
		log.Info("this email has no acttachments")
	} else {
		defer file.Close()

		if handler.Size > (5 << 20) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "the maximum size of attachments is 5MB")
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
	from := os.Getenv("SMTP_DEFAULT_FROM")
	to := r.Form.Get("to") // x will be "" if parameter is not set
	cc := r.Form.Get("cc")
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
			log.Error("schedule time wrong format: " + schedule)
			return
		}
		mysqldb.Save(from, to, subject, body, attachment, "", time.Time{}, scheduleTime.Add(-time.Hour*7), "PENDING")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "The email has been successfully saved and will automatically be sent when the time comes.")
		return
	}
	//from, strings.Split(to, ";"), subject, body, attachment, ""
	if !email.Send(email.Email{From: from, To: strings.Split(to, ";"), CC: strings.Split(cc, ";"), Subject: subject, Body: body, Attachment: attachment}) {
		mysqldb.Save(from, to, subject, body, attachment, "", time.Time{}, time.Time{}, "ERROR")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		mysqldb.Save(from, to, subject, body, attachment, "", time.Now(), time.Time{}, "SENT")
	}
	log.Info("the email has been sent")
	fmt.Fprintf(w, "the email has been sent successfully!")
}

func SendScheduleEmail() {
	for {
		db := mysqldb.New()

		var emails []entity.Email
		db.Where("status = ?", "PENDING").Find(&emails)
		log.Info("there are %v schedule email in DB", len(emails))
		for _, mail := range emails {
			if mail.ScheduleSentTime.Before(time.Now()) {

				if email.Send(email.Email{From: mail.From, To: strings.Split(mail.To, ";"), CC: strings.Split(mail.CC, ";"), Subject: mail.Subject, Body: mail.Body, Attachment: mail.Attachment}) {
					db.Model(&mail).Updates(entity.Email{SentTime: time.Now(), Status: "SENT"})
					log.Info("schedule mail has been sent!")
				} else {
					db.Model(&mail).Updates(entity.Email{Status: "ERROR"})
					log.Warn("schedule mail sent failed!")
				}
			}
		}
		time.Sleep(time.Second * 20)
	}
}
