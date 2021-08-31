package types

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type (
	repository interface {
		Get(ctx context.Context, id string) (*Email, error)
		Create(ctx context.Context, user *Email) error
		Update(ctx context.Context, user *Email) error
	}

	Service struct {
		repo repository
	}
)

func NewService(repo repository) *Service {
	return &Service{
		repo: repo,
	}
}

type Email struct {
	gorm.Model
	Id               int
	From             string
	To               string
	CC               string
	Subject          string
	Body             string
	Attachment       string
	Template         string
	SentTime         time.Time
	ScheduleSentTime time.Time
	Status           string
}
