//go:generate mockgen -source handler.go -destination handler_mock_test.go -package email_test
package email

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/volam1999/gomail/internal/app/types"
	"github.com/volam1999/gomail/internal/pkg/db/mysqldb"
	mail "github.com/volam1999/gomail/internal/pkg/email"

	"github.com/volam1999/gomail/internal/pkg/log"

	"github.com/gorilla/mux"
)

type (
	service interface {
		Create(email *types.Email) (string, error)
		FindAll() (*[]types.Email, error)
		FindByEmailId(emailId string) (*types.Email, error)
	}
	Handler struct {
		srv service
	}
)

func NewHandler(srv service) *Handler {
	return &Handler{
		srv: srv,
	}
}

func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {

	emails, err := h.srv.FindAll()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fmt.Println(emails)
	json, _ := json.Marshal(*emails)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func (h *Handler) FindByEmailId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	email, err := h.srv.FindByEmailId(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json, _ := json.Marshal(email)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func (h *Handler) SendEmail(w http.ResponseWriter, r *http.Request) {

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

		//db.Create(&entity.Email{From: from, To: to, Subject: subject, CC: cc, Body: body, Attachment: ConvertArrayToString(attachments), ScheduleSentTime: scheduleTime.Add(-time.Hour * 7), Status: "PENDING"})
		h.srv.Create(&types.Email{From: from, To: to, Subject: subject, CC: cc, Body: body, Attachment: ConvertArrayToString(attachments), ScheduleSentTime: scheduleTime.Add(-time.Hour * 7), Status: "PENDING"})
		w.WriteHeader(http.StatusOK)
		log.Info("the email has been saved and will be sent automatically.")
		fmt.Fprintf(w, "the email has been saved and will be sent automatically.")
		return
	}

	if !mail.Send(mail.Email{From: from, To: strings.Split(to, ";"), CC: strings.Split(cc, ";"), Subject: subject, Body: body, Attachments: attachments}) {
		//db.Create(&entity.Email{From: from, To: to, CC: cc, Subject: subject, Body: body, Attachment: ConvertArrayToString(attachments), Status: "ERROR"})
		h.srv.Create(&types.Email{From: from, To: to, CC: cc, Subject: subject, Body: body, Attachment: ConvertArrayToString(attachments), Status: "ERROR"})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.srv.Create(&types.Email{From: from, To: to, CC: cc, Subject: subject, Body: body, Attachment: ConvertArrayToString(attachments), SentTime: time.Now(), Status: "SENT"})
	//db.Create(&entity.Email{From: from, To: to, CC: cc, Subject: subject, Body: body, Attachment: ConvertArrayToString(attachments), SentTime: time.Now(), Status: "SENT"})
	log.Info("the email has been sent")
	fmt.Fprintf(w, "the email has been sent successfully!")
}

func (h *Handler) SendScheduleEmail() {
	log.Warn("automatic check and send schedule email in the database every 20s.")
	for {
		db := mysqldb.New()
		var emails []types.Email
		db.Where("status = ?", "PENDING").Find(&emails)
		log.Infof("there are [%v] schedule email in the database", len(emails))
		for _, email := range emails {
			if email.ScheduleSentTime.Before(time.Now()) {

				if mail.Send(mail.Email{From: email.From, To: strings.Split(email.To, ";"), CC: strings.Split(email.CC, ";"), Subject: email.Subject, Body: email.Body, Attachments: strings.Split(email.Attachment, ";")}) {
					db.Model(&email).Updates(types.Email{SentTime: time.Now(), Status: "SENT"})
					log.Infof("the scheduled email [%v] has been sent!", email.Id)
				} else {
					db.Model(&email).Updates(types.Email{Status: "ERROR"})
					log.Warnf("the scheduled email [%v] could not be delivered", email.Id)
				}
			}
		}
		time.Sleep(time.Second * 20)
	}
}
