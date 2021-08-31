package api

import (
	"github.com/volam1999/gomail/internal/app/email"
	"github.com/volam1999/gomail/internal/pkg/db/mysqldb"
)

func newEmailService() (*email.Service, error) {
	db := mysqldb.New()
	repo := email.NewMysqlDBRepository(db)
	return email.New(repo), nil
}

func newEmailHandler(srv *email.Service) *email.Handler {
	return email.NewHandler(srv)
}
