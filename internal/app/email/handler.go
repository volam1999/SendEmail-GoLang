//go:generate mockgen -source handler.go -destination handler_mock_test.go -package email_test
package email

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/volam1999/gomail/internal/app/types"
	mail "github.com/volam1999/gomail/internal/pkg/email"

	"github.com/volam1999/gomail/internal/pkg/log"

	"github.com/gorilla/mux"
)

type (
	service interface {
		Create(email *types.Email) (int, error)
		FindAll() (*[]types.Email, error)
		FindByEmailId(emailId int) (*types.Email, error)
		Send(email *mail.Email) bool
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
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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
	err := r.ParseMultipartForm(5 << 20)
	attachments := []string{}

	if err != nil && strings.Contains(err.Error(), "isn't multipart/form-data") {
		// email has no attachment
		log.Error(err.Error())
	} else {
		if r.MultipartForm != nil {
			files := r.MultipartForm.File["attachments"]
			for _, file := range files {

				if file.Size > (5 << 20) {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprintf(w, "file {%v} has passed the maximum allowed size of attachments is 5MB", file.Filename)
					return
				}

				ext := file.Filename[len(file.Filename)-4:]

				f, err := file.Open()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Read attachment failed!\nFiles: %v ", file.Filename)
					log.Errorf("Read attachment failed!\nFiles: %v ", file.Filename)
					return
				}

				tempFile, err := ioutil.TempFile("uploads", "upload-*"+ext)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Write attachment failed! \nFiles: %v ", file.Filename)
					log.Errorf("Write attachment failed!\nFiles: %v ", file.Filename)
					return
				}
				defer tempFile.Close()

				// read all of the contents of our uploaded file into a
				// byte array
				fileBytes, err := ioutil.ReadAll(f)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Write attachment failed!\nFiles: %v ", file.Filename)
					log.Errorf("Write attachment failed!\nFiles: %v ", file.Filename)
					return
				}
				// write this byte array to our temporary file
				tempFile.Write(fileBytes)
				attachments = append(attachments, tempFile.Name())
			}
		}
	}

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

		_, err = h.srv.Create(&types.Email{From: from, To: to, Subject: subject, CC: cc, Body: body, Attachment: ConvertArrayToString(attachments), ScheduleSentTime: scheduleTime.Add(-time.Hour * 7), Status: "PENDING"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("the email has not been saved.", err.Error())
			fmt.Fprintf(w, "error saving the email in the database, the email will not be delivered")
			return
		}

		w.WriteHeader(http.StatusOK)
		log.Info("the email has been saved and will be sent automatically.")
		fmt.Fprintf(w, "the email has been saved and will be sent automatically.")
		return
	}

	if !h.srv.Send(&mail.Email{From: from, To: strings.Split(to, ";"), CC: strings.Split(cc, ";"), Subject: subject, Body: body, Attachments: attachments}) {
		_, err := h.srv.Create(&types.Email{From: from, To: to, CC: cc, Subject: subject, Body: body, Attachment: ConvertArrayToString(attachments), Status: "ERROR"})
		if err != nil {
			log.Error("email sending failed and saving data to the database failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "email sending failed")
		return
	}

	log.Info("the email has been sent")
	_, err = h.srv.Create(&types.Email{From: from, To: to, CC: cc, Subject: subject, Body: body, Attachment: ConvertArrayToString(attachments), SentTime: time.Now(), Status: "SENT"})
	if err != nil {
		log.Error("email sent successfully but saving data to the database failed!")
		fmt.Fprintf(w, "temail sent successfully but saving data to the database failed!")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "the email has been sent successfully!")
}

func (h *Handler) SendScheduleEmail() {
	h.srv.SendScheduleEmail()
}
