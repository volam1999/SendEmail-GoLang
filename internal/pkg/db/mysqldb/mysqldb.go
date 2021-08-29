package mysqldb

import (
	"fmt"
	"log"

	"github.com/volam1999/gomail/internal/app/entity"
	"github.com/volam1999/gomail/internal/pkg/config/envconfig"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

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
