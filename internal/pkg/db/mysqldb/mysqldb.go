package mysqldb

import (
	"fmt"
	"log"
	"time"

	"main/internal/app/entity"
	"main/internal/pkg/config/envconfig"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Save(from string, to string, subject string, body string, atm string, template string, sentTime time.Time, scheduleSendTime time.Time, status string) {
	dsn := fmt.Sprintf("%v:%v@/%v?charset=utf8mb4&parseTime=True&loc=Local", envconfig.Username, envconfig.Password, envconfig.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
		panic("failed to connect database")
	}
	// Migrate the schema
	db.AutoMigrate(&entity.Email{})
	// Create
	db.Create(&entity.Email{From: from, To: to, Subject: subject, Body: body, Attachment: atm, Template: template, SentTime: sentTime, ScheduleSentTime: scheduleSendTime, Status: status})
}

func New() *gorm.DB {
	dsn := fmt.Sprintf("%v:%v@/%v?charset=utf8mb4&parseTime=True&loc=Local", envconfig.Username, envconfig.Password, envconfig.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
		panic("failed to connect database")
	}
	// Migrate the schema
	db.AutoMigrate(&entity.Email{})
	return db
}
