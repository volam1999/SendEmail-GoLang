package email_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/volam1999/gomail/internal/app/email"
	"github.com/volam1999/gomail/internal/app/types"
	mail "github.com/volam1999/gomail/internal/pkg/email"
)

var (
	email1 = &types.Email{
		Id: 1,
	}

	emails = []types.Email{
		{
			Id: 1,
		},
		{
			Id: 2,
		},
	}

	requests = []request{
		{
			from:        "system@outlook.com",
			to:          "admin@gmail.com",
			subject:     "Unit Test",
			body:        "Learn Unit Test in goLang",
			attachments: "",
		},
		{
			from:    "system@outlook.com",
			subject: "Unit Test",
			body:    "Learn Unit Test in goLang",
		},
		{
			from:     "system@outlook.com",
			to:       "admin@gmail.com",
			subject:  "Unit Test",
			body:     "Learn Unit Test in goLang",
			schedule: "02-09-2021 11:11",
		},
		{
			from:     "system@outlook.com",
			to:       "admin@gmail.com",
			subject:  "Unit Test",
			body:     "Learn Unit Test in goLang",
			schedule: "11:11 20-21-2020",
		},
	}
)

//serverErr := errors.New("internal server error")
type expect struct {
	code int
	body string
}

type request struct {
	from        string
	to          string
	cc          string
	subject     string
	body        string
	schedule    string
	attachments string
}

func TestHandlerFindEmailById(t *testing.T) {
	srv := NewMockservice(gomock.NewController(t))
	handler := email.NewHandler(srv)
	err := errors.New("not found")
	testCases := []struct {
		name     string
		tearDown func()
		input    int
		expect   expect
	}{
		{
			name: "get email data in the database success",
			tearDown: func() {
				srv.EXPECT().FindByEmailId(strconv.Itoa(email1.Id)).Times(1).Return(email1, nil)
			},
			input: email1.Id,
			expect: expect{
				code: http.StatusOK,
				body: `{"ID":0,"CreatedAt":"0001-01-01T00:00:00Z","UpdatedAt":"0001-01-01T00:00:00Z","DeletedAt":null,"Id":1,"From":"","To":"","CC":"","Subject":"","Body":"","Attachment":"","Template":"","SentTime":"0001-01-01T00:00:00Z","ScheduleSentTime":"0001-01-01T00:00:00Z","Status":""}`,
			},
		},
		{
			name: "get email data in the database failed no record was found",
			tearDown: func() {
				srv.EXPECT().FindByEmailId(strconv.Itoa(email1.Id)).Times(1).Return(nil, err)
			},
			input: email1.Id,
			expect: expect{
				code: http.StatusNotFound,
				body: "",
			},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			test.tearDown()
			w := httptest.NewRecorder()
			r, err := http.NewRequest(http.MethodGet, "", nil)
			if err != nil {
				t.Error(err)
			}
			vars := map[string]string{
				"id": strconv.Itoa(test.input),
			}
			r = mux.SetURLVars(r, vars)

			handler.FindByEmailId(w, r)
			if w.Code != test.expect.code {
				t.Errorf("got code=%d, wants code=%d", w.Code, test.expect.code)
			}
			gotBody := strings.TrimSpace(w.Body.String())
			if gotBody != test.expect.body {
				t.Errorf("got body=%s, wants body=%s", gotBody, test.expect.body)
			}
		})
	}
}

func TestHandlerFindAll(t *testing.T) {
	srv := NewMockservice(gomock.NewController(t))
	handler := email.NewHandler(srv)
	err := errors.New("not found")
	testCases := []struct {
		name     string
		tearDown func()
		expect   expect
	}{
		{
			name: "get email data in the database success",
			tearDown: func() {
				srv.EXPECT().FindAll().Times(1).Return(&emails, nil)
			},
			expect: expect{
				code: http.StatusOK,
				body: `[{"ID":0,"CreatedAt":"0001-01-01T00:00:00Z","UpdatedAt":"0001-01-01T00:00:00Z","DeletedAt":null,"Id":1,"From":"","To":"","CC":"","Subject":"","Body":"","Attachment":"","Template":"","SentTime":"0001-01-01T00:00:00Z","ScheduleSentTime":"0001-01-01T00:00:00Z","Status":""},{"ID":0,"CreatedAt":"0001-01-01T00:00:00Z","UpdatedAt":"0001-01-01T00:00:00Z","DeletedAt":null,"Id":2,"From":"","To":"","CC":"","Subject":"","Body":"","Attachment":"","Template":"","SentTime":"0001-01-01T00:00:00Z","ScheduleSentTime":"0001-01-01T00:00:00Z","Status":""}]`,
			},
		},
		{
			name: "get email data in the database failed no record was found",
			tearDown: func() {
				srv.EXPECT().FindAll().Times(1).Return(nil, err)
			},
			expect: expect{
				code: http.StatusNotFound,
				body: "",
			},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			test.tearDown()
			w := httptest.NewRecorder()
			r, err := http.NewRequest(http.MethodGet, "", nil)
			if err != nil {
				t.Error(err)
			}

			handler.FindAll(w, r)
			if w.Code != test.expect.code {
				t.Errorf("got code=%d, wants code=%d", w.Code, test.expect.code)
			}
			gotBody := strings.TrimSpace(w.Body.String())
			if gotBody != test.expect.body {
				t.Errorf("got body=%s, wants body=%s", gotBody, test.expect.body)
			}
		})
	}
}

func TestHandlerSendEmail(t *testing.T) {

	emailToSend := mail.Email{
		From:        requests[0].from,
		To:          strings.Split(requests[0].to, ";"),
		Subject:     requests[0].subject,
		Body:        requests[0].body,
		CC:          strings.Split(requests[0].cc, ";"),
		Attachments: []string{},
	}

	emailToCreate := types.Email{
		From:       requests[0].from,
		To:         requests[0].to,
		Subject:    requests[0].subject,
		Body:       requests[0].body,
		CC:         requests[0].cc,
		Attachment: "",
		SentTime:   time.Now(),
		Status:     "SENT",
	}

	emailErrorToCreate := types.Email{
		From:       requests[0].from,
		To:         requests[0].to,
		Subject:    requests[0].subject,
		Body:       requests[0].body,
		CC:         requests[0].cc,
		Attachment: "",
		Status:     "ERROR",
	}

	layout := "02-01-2006 15:04"
	scheduleTime, _ := time.Parse(layout, requests[2].schedule)
	emailScheduleToCreate := types.Email{
		From:             requests[2].from,
		To:               requests[2].to,
		Subject:          requests[2].subject,
		Body:             requests[2].body,
		CC:               requests[2].cc,
		Attachment:       "",
		ScheduleSentTime: scheduleTime.Add(-time.Hour * 7),
		Status:           "PENDING",
	}

	srv := NewMockservice(gomock.NewController(t))
	handler := email.NewHandler(srv)
	testCases := []struct {
		name     string
		tearDown func()
		input    request
		expect   expect
	}{
		{
			name:  "send email success",
			input: requests[0],
			tearDown: func() {
				srv.EXPECT().Send(emailToSend).Times(1).Return(true)
				srv.EXPECT().Create(&emailToCreate).Times(1).Return("", nil)
			},
			expect: expect{
				code: http.StatusOK,
			},
		},
		{
			name:  "send email failed because the missing recipient",
			input: requests[1],
			tearDown: func() {

			},
			expect: expect{
				code: http.StatusBadRequest,
			},
		},
		{
			name:  "save the email with the schedule success",
			input: requests[2],
			tearDown: func() {
				srv.EXPECT().Create(&emailScheduleToCreate).Times(1).Return("", nil)
			},
			expect: expect{
				code: http.StatusOK,
			},
		},
		{
			name:  "send email failed because the email has the invalid schedule time format",
			input: requests[3],
			tearDown: func() {

			},
			expect: expect{
				code: http.StatusBadRequest,
			},
		},
		{
			name:  "send email failed because the mailler has failed to send",
			input: requests[0],
			tearDown: func() {
				srv.EXPECT().Send(emailToSend).Times(1).Return(false)
				srv.EXPECT().Create(&emailErrorToCreate).Times(1).Return("", nil)
			},
			expect: expect{
				code: http.StatusInternalServerError,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			test.tearDown()
			w := httptest.NewRecorder()
			form := fmt.Sprintf("from=%v&to=%v&cc=%v&subject=%v&body=%v&schedule=%v", test.input.from, test.input.to, test.input.cc, test.input.subject, test.input.body, test.input.schedule)
			r, err := http.NewRequest(http.MethodPost, "", strings.NewReader(form))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if err != nil {
				t.Error(err)
			}
			handler.SendEmail(w, r)
			if w.Code != test.expect.code {
				t.Errorf("got code=%d, wants code=%d", w.Code, test.expect.code)
			}
		})
	}
}

func TestHandlerSendScheduleEmail(t *testing.T) {

}
