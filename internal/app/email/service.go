//go:generate mockgen -source service.go -destination service_mock_test.go -package email_test
package email

import (
	"strings"
	"time"

	"github.com/volam1999/gomail/internal/app/types"
	mail "github.com/volam1999/gomail/internal/pkg/email"
	"github.com/volam1999/gomail/internal/pkg/log"
)

type (
	Repository interface {
		Create(email *types.Email) (int, error)
		Update(emailId int, email *types.Email) error
		FindAll() (*[]types.Email, error)
		FindByEmailId(emailId int) (*types.Email, error)
		FindAllScheduleEmail() (*[]types.Email, error)
	}

	Service struct {
		repo   Repository
		mailer mail.Mailer
	}
)

func New(repo Repository, mailer mail.Mailer) *Service {
	return &Service{
		repo:   repo,
		mailer: mailer,
	}
}

func (s *Service) Create(email *types.Email) (int, error) {
	return s.repo.Create(email)
}

func (s *Service) FindAll() (*[]types.Email, error) {
	return s.repo.FindAll()
}

func (s *Service) FindByEmailId(emailId int) (*types.Email, error) {
	return s.repo.FindByEmailId(emailId)
}

func (s *Service) Update(emailId int, email *types.Email) error {
	return s.repo.Update(emailId, email)
}

func (s *Service) Send(email *mail.Email) error {
	return s.mailer.Send(email)
}

func (s *Service) SendScheduleEmail() {
	log.Warn("automatic check and send schedule email in the database every 20s.")
	var err error
	for {
		var emails *[]types.Email
		emails, err = s.repo.FindAllScheduleEmail()
		if err != nil {
			log.Errorf("read data in the database failed, retry after 1 minute")
			time.Sleep(time.Second * 60)
			continue
		}
		log.Infof("there are [%v] schedule email in the database", len(*emails))
		for _, email := range *emails {
			if email.ScheduleSentTime.Before(time.Now()) {

				if err := s.Send(&mail.Email{From: email.From, To: strings.Split(email.To, ";"), CC: strings.Split(email.CC, ";"), Subject: email.Subject, Body: email.Body, Attachments: strings.Split(email.Attachment, ";")}); err != nil {
					err := s.repo.Update(email.Id, &types.Email{SentTime: time.Now(), Status: "SENT"})
					if err != nil {
						log.Error("the scheduled email [%v] has been sent! but have an error when saving to the db", email.Id)
						continue
					}
					log.Infof("the scheduled email [%v] has been sent!", email.Id)
				} else {
					err := s.repo.Update(email.Id, &types.Email{Status: "ERROR"})
					if err != nil {
						log.Error("the scheduled email [%v] could not be delivered! and have an error when saving to the db", email.Id)
						continue
					}
					log.Warnf("the scheduled email [%v] could not be delivered", email.Id)
				}
			}
		}
		time.Sleep(time.Second * 20)
	}
}
