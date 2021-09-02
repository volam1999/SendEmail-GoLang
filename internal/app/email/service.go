//go:generate mockgen -source service.go -destination email_mock_test.go -package email_test
package email

import (
	"strconv"
	"strings"
	"time"

	"github.com/volam1999/gomail/internal/app/types"
	mail "github.com/volam1999/gomail/internal/pkg/email"
	"github.com/volam1999/gomail/internal/pkg/log"
)

type (
	Repository interface {
		Create(email *types.Email) (string, error)
		Update(emailId string, email *types.Email) error
		FindAll() (*[]types.Email, error)
		FindByEmailId(emailId string) (*types.Email, error)
		FindAllScheduleEmail() (*[]types.Email, error)
	}

	Service struct {
		repo Repository
	}
)

func New(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(email *types.Email) (string, error) {
	return s.repo.Create(email)
}

func (s *Service) FindAll() (*[]types.Email, error) {
	return s.repo.FindAll()
}

func (s *Service) FindByEmailId(emailId string) (*types.Email, error) {
	return s.repo.FindByEmailId(emailId)
}

func (s *Service) Update(emailId string, email *types.Email) error {
	return s.repo.Update(emailId, email)
}

func (s *Service) Send(email mail.Email) bool {
	return mail.Send(email)
}

func (s *Service) SendScheduleEmail() {
	log.Warn("automatic check and send schedule email in the database every 20s.")
	for {
		var emails *[]types.Email
		emails, _ = s.repo.FindAllScheduleEmail()
		log.Infof("there are [%v] schedule email in the database", len(*emails))
		for _, email := range *emails {
			if email.ScheduleSentTime.Before(time.Now()) {

				if s.Send(mail.Email{From: email.From, To: strings.Split(email.To, ";"), CC: strings.Split(email.CC, ";"), Subject: email.Subject, Body: email.Body, Attachments: strings.Split(email.Attachment, ";")}) {
					//db.Model(&email).Updates(types.Email{SentTime: time.Now(), Status: "SENT"})
					s.repo.Update(strconv.Itoa(email.Id), &types.Email{SentTime: time.Now(), Status: "SENT"})
					log.Infof("the scheduled email [%v] has been sent!", email.Id)
				} else {
					//db.Model(&email).Updates(types.Email{Status: "ERROR"})
					s.repo.Update(strconv.Itoa(email.Id), &types.Email{Status: "ERROR"})
					log.Warnf("the scheduled email [%v] could not be delivered", email.Id)
				}
			}
		}
		time.Sleep(time.Second * 20)
	}
}
