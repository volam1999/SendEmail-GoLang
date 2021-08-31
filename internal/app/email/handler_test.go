package email_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/volam1999/gomail/internal/app/email"
	"github.com/volam1999/gomail/internal/app/types"
)

var email1 = &types.Email{
	Id:   1,
	From: "jack",
}

//serverErr := errors.New("internal server error")
type expect struct {
	code int
	body string
}

func TestHandlerFindAll(t *testing.T) {
	srv := NewMockservice(gomock.NewController(t))
	handler := email.NewHandler(srv)

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
				body: `{"Id": 1,"From":"jack"}`,
			},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			test.tearDown()
			w := httptest.NewRecorder()

			r, err := http.NewRequest(http.MethodGet, "/?id="+strconv.Itoa(test.input), nil)
			if err != nil {
				t.Error(err)
			}
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
