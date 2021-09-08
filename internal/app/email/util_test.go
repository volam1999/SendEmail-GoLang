package email_test

import (
	"testing"

	"github.com/volam1999/gomail/internal/app/email"
)

func TestConvertArrayToString(t *testing.T) {
	email.ConvertArrayToString([]string{})
}
