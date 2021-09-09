package mail

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/volam1999/gomail/internal/pkg/config/envconfig"
	"github.com/volam1999/gomail/internal/pkg/log"

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

	Mailer struct {
		dialer *gomail.Dialer
	}

	Config struct {
		Address     string `envconfig:"SMTP_ADDRESS"`
		Username    string `envconfig:"SMTP_USERNAME"`
		Password    string `envconfig:"SMTP_PASSWORD"`
		DefaultFrom string `envconfig:"SMTP_DEFAULT_FROM" default:"goway382@gmail.com"`
	}
)

func LoadConfigFromEnv() *Config {
	var conf Config
	envconfig.Load("", &conf)
	return &conf
}

func New(conf *Config) (*Mailer, error) {
	username := conf.Username
	password := conf.Password
	host, port, _ := net.SplitHostPort(conf.Address)
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("address must be in form of <host>:<port>: %w", err)
	}
	d := gomail.NewDialer(host, portInt, username, password)

	return &Mailer{
		dialer: d,
	}, nil
}

func (m *Mailer) Send(email *Email) error {

	from := os.Getenv("SMTP_DEFAULT_EMAIL")
	if email.From != "" {
		from = email.From
	}
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", email.To...)
	if email.CC[0] != "" {
		msg.SetHeader("Cc", email.CC...)
	}

	msg.SetHeader("Subject", email.Subject)
	msg.SetBody("text/html", email.Body)

	if email.Attachments[0] != "" {
		for _, atm := range email.Attachments {
			msg.Attach(atm)
		}
	}

	if err := m.dialer.DialAndSend(msg); err != nil {
		log.Error("failed to send mail: %w", err)
		return err
	}
	return nil
}
