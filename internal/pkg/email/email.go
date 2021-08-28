package email

import (
	"net"
	"os"
	"strconv"

	"main/internal/pkg/log"

	"gopkg.in/gomail.v2"
)

type (
	Email struct {
		From        string
		To          []string
		CC          []string
		Subject     string
		Body        string
		Attachments []string
		Template    string
	}
)

func Send(email Email) bool {
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	host, port, _ := net.SplitHostPort(os.Getenv("SMTP_ADDRESS"))
	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Error("address must be in form of <host>:<port>: %w", err)
		return false
	}
	d := gomail.NewDialer(host, portInt, username, password)

	from := os.Getenv("SMTP_DEFAULT_EMAIL")
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", email.To...)
	if len(email.CC) > 1 {
		msg.SetHeader("Cc", email.CC...)
	}

	msg.SetHeader("Subject", email.Subject)
	msg.SetBody("text/html", email.Body)

	for _, atm := range email.Attachments {
		msg.Attach(atm)
	}

	if err := d.DialAndSend(msg); err != nil {
		log.Error("failed to send mail: %w", err)
		return false
	}
	return true
}
