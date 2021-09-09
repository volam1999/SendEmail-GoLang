package api

import (
	"github.com/volam1999/gomail/internal/app/email"
	"github.com/volam1999/gomail/internal/pkg/db/mysqldb"
	mail "github.com/volam1999/gomail/internal/pkg/email"
	"github.com/volam1999/gomail/internal/pkg/log"
)

func newEmailService() (*email.Service, error) {
	db := mysqldb.MustNew(mysqldb.LoadConfigFromEnv())
	repo := email.NewMysqlDBRepository(db)

	mailer, err := mail.New(mail.LoadConfigFromEnv())
	if err != nil {
		log.Error(err.Error())
	}
	return email.New(repo, *mailer), nil
}

func newEmailHandler(srv *email.Service) *email.Handler {
	return email.NewHandler(srv)
}
