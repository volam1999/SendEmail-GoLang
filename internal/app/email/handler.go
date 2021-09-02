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
	mail "github.com/volam1999/gomail/internal/pkg/email"

	"github.com/volam1999/gomail/internal/pkg/log"

	"github.com/gorilla/mux"
)

type (
	service interface {
		Create(email *types.Email) (string, error)
		FindAll() (*[]types.Email, error)
		FindByEmailId(emailId string) (*types.Email, error)
		Send(email mail.Email) bool
		SendScheduleEmail()
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
	attachments := []string{}
	if r.MultipartForm != nil {
		files := r.MultipartForm.File["attachments"]
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
	}

	r.ParseForm() // Parses the request body
	from := r.Form.Get("from")
	if from == "" {
		from = os.Getenv("SMTP_DEFAULT_EMAIL")
	}
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
	fmt.Println(from, to, cc, subject, body)
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

		h.srv.Create(&types.Email{From: from, To: to, Subject: subject, CC: cc, Body: body, Attachment: ConvertArrayToString(attachments), ScheduleSentTime: scheduleTime.Add(-time.Hour * 7), Status: "PENDING"})
		w.WriteHeader(http.StatusOK)
		log.Info("the email has been saved and will be sent automatically.")
		fmt.Fprintf(w, "the email has been saved and will be sent automatically.")
		return
	}
	//mail.Email{From: from, To: strings.Split(to, ";"), CC: strings.Split(cc, ";"), Subject: subject, Body: body, Attachments: attachments}
	if !h.srv.Send(mail.Email{From: from, To: strings.Split(to, ";"), CC: strings.Split(cc, ";"), Subject: subject, Body: body, Attachments: attachments}) {

		h.srv.Create(&types.Email{From: from, To: to, CC: cc, Subject: subject, Body: body, Attachment: ConvertArrayToString(attachments), Status: "ERROR"})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.srv.Create(&types.Email{From: from, To: to, CC: cc, Subject: subject, Body: body, Attachment: ConvertArrayToString(attachments), SentTime: time.Now(), Status: "SENT"})
	log.Info("the email has been sent")
	fmt.Fprintf(w, "the email has been sent successfully!")
}

func (h *Handler) SendScheduleEmail() {
	h.srv.SendScheduleEmail()
}
