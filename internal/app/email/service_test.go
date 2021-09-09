package email_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/volam1999/gomail/internal/app/email"
	types "github.com/volam1999/gomail/internal/app/types"
	mail "github.com/volam1999/gomail/internal/pkg/email"
)

func TestCreateEmail(t *testing.T) {
	mockedRepo := NewMockRepository(gomock.NewController(t))
	service := email.New(mockedRepo, mail.Mailer{})

	email := &types.Email{From: "a", To: "b", CC: "c"}
	dbErr := errors.New("cannot connect to database")
	testCases := []struct {
		name     string
		tearDown func()
		input    *types.Email
		output   error
	}{
		{
			name:  "create email succeed",
			input: email,
			tearDown: func() {
				mockedRepo.EXPECT().Create(email).Times(1).Return(1, nil)
			},
			output: nil,
		},
		{
			name:  "create email failed because of database connection failed",
			input: email,
			tearDown: func() {
				mockedRepo.EXPECT().Create(email).Times(1).Return(-1, dbErr)
			},
			output: dbErr,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			test.tearDown()
			_, err := service.Create(test.input)
			if err != test.output {
				t.Errorf("got err = %v, expects err = %v", err, test.output)
			}
		})
	}

}
