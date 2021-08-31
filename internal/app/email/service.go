package email

import (
	"github.com/volam1999/gomail/internal/app/types"
)

type (
	Repository interface {
		Create(email *types.Email) (string, error)
		FindAll() (*[]types.Email, error)
		FindByEmailId(emailId string) (*types.Email, error)
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
