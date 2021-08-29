package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/volam1999/gomail/internal/app/entity"
	"github.com/volam1999/gomail/internal/pkg/db/mysqldb"
	"github.com/volam1999/gomail/internal/pkg/email"

	"github.com/volam1999/gomail/internal/pkg/log"

	"github.com/gorilla/mux"
)

func GetAllEmail(w http.ResponseWriter, r *http.Request) {

	db := mysqldb.New()
	var emails []entity.Email
	db.Find(&emails)
	json, _ := json.Marshal(emails)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func GetEmailById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	db := mysqldb.New()

	var email entity.Email
	err := db.First(&email, id).Error
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	json, _ := json.Marshal(email)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func SendEmail(w http.ResponseWriter, r *http.Request) {

	// Parse our multipart form, 5 << 20 specifies a maximum
	// upload of 5 MB files.
	r.ParseMultipartForm(5 << 20)
	files := r.MultipartForm.File["attachments"]

	attachments := []string{}
	for _, file := range files {
		if file.Size > (5 << 20) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "file {%v} has passed the maximum allowed size of attachments is 5MB", file.Filename)
			return
		}
		ext := file.Filename[len(file.Filename)-4:]

		f, _ := file.Open()
		tempFile, err := ioutil.TempFile("uploads", "upload-*"+ext)
		if err != nil {
			fmt.Println(err)
			log.Error(err)
		}
		defer tempFile.Close()

		// read all of the contents of our uploaded file into a
		// byte array
		fileBytes, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Println(err)
		}
		// write this byte array to our temporary file
		tempFile.Write(fileBytes)
		attachments = append(attachments, tempFile.Name())
	}

	r.ParseForm() // Parses the request body
	from := os.Getenv("SMTP_DEFAULT_EMAIL")
	to := r.Form.Get("to")
	if to == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "who do i email?")
		return
	}
	cc := r.Form.Get("cc")
	subject := r.Form.Get("subject")
	body := r.Form.Get("body")
	schedule := r.Form.Get("schedule")
	scheduleTime := time.Time{}
	db := mysqldb.New()

	if schedule != "" {
		layout := "02-01-2006 15:04"
		var err error
		scheduleTime, err = time.Parse(layout, schedule)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "schedule time format expected: dd-MM-yyyy hh:mm")
			log.Error("schedule time wrong format: " + schedule)
			return
		}
		db.Create(&entity.Email{From: from, To: to, Subject: subject, CC: cc, Body: body, Attachment: convertArrayToString(attachments), ScheduleSentTime: scheduleTime.Add(-time.Hour * 7), Status: "PENDING"})
		w.WriteHeader(http.StatusOK)
		log.Info("the email has been saved and will be sent automatically.")
		fmt.Fprintf(w, "the email has been saved and will be sent automatically.")
		return
	}

	if !email.Send(email.Email{From: from, To: strings.Split(to, ";"), CC: strings.Split(cc, ";"), Subject: subject, Body: body, Attachments: attachments}) {
		db.Create(&entity.Email{From: from, To: to, CC: cc, Subject: subject, Body: body, Attachment: convertArrayToString(attachments), Status: "ERROR"})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	db.Create(&entity.Email{From: from, To: to, CC: cc, Subject: subject, Body: body, Attachment: convertArrayToString(attachments), SentTime: time.Now(), Status: "SENT"})
	log.Info("the email has been sent")
	fmt.Fprintf(w, "the email has been sent successfully!")
}

func SendScheduleEmail() {
	log.Warn("automatic check and send schedule email in the database every 20s.")
	for {
		db := mysqldb.New()
		var emails []entity.Email
		db.Where("status = ?", "PENDING").Find(&emails)
		log.Infof("there are [%v] schedule email in the database", len(emails))
		for _, mail := range emails {
			if mail.ScheduleSentTime.Before(time.Now()) {

				if email.Send(email.Email{From: mail.From, To: strings.Split(mail.To, ";"), CC: strings.Split(mail.CC, ";"), Subject: mail.Subject, Body: mail.Body, Attachments: strings.Split(mail.Attachment, ";")}) {
					db.Model(&mail).Updates(entity.Email{SentTime: time.Now(), Status: "SENT"})
					log.Infof("the scheduled email [%v] has been sent!", mail.Id)
				} else {
					db.Model(&mail).Updates(entity.Email{Status: "ERROR"})
					log.Warnf("the scheduled email [%v] could not be delivered", mail.Id)
				}
			}
		}
		time.Sleep(time.Second * 20)
	}
}
