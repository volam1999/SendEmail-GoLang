package email_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/volam1999/gomail/internal/app/email"
	"github.com/volam1999/gomail/internal/app/types"
)

var (
	email1 = &types.Email{
		Id:   1,
		To:   "admin@gmail.com",
		CC:   "manager@gmail.com",
		Body: "Hello world",
	}

	emails = []types.Email{
		{
			Id: 1,
		},
		{
			Id: 2,
		},
	}
)

//serverErr := errors.New("internal server error")
type expect struct {
	code int
	body string
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
	srv := NewMockservice(gomock.NewController(t))
	handler := email.NewHandler(srv)
	testCases := []struct {
		name     string
		tearDown func()
		input    *types.Email
		expect   expect
	}{
		{
			name:  "send email success",
			input: email1,
			expect: expect{
				code: http.StatusOK,
			},
		},
		{
			name:  "get email data in the database failed no record was found",
			input: email1,
			expect: expect{
				code: http.StatusNotFound,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			var inputBody bytes.Buffer
			if err := json.NewEncoder(&inputBody).Encode(test.input); err != nil {
				t.Error(err)
			}
			r, err := http.NewRequest(http.MethodPost, "", &inputBody)
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
