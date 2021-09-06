package email_test

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/volam1999/gomail/internal/app/email"
)

func TestRoutes(t *testing.T) {
	srv := NewMockservice(gomock.NewController(t))
	handler := email.NewHandler(srv)
	handler.Routes()
}
