package types

import (
	"time"

	"gorm.io/gorm"
)

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
