package mysqldb

import (
	"fmt"

	"github.com/volam1999/gomail/internal/app/types"
	"github.com/volam1999/gomail/internal/pkg/config/envconfig"
	"github.com/volam1999/gomail/internal/pkg/log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func New() *gorm.DB {
	log.Infof("dialing to target MySqlDb at: %v, database: %v", "localhost:3306", envconfig.Database)
	dsn := fmt.Sprintf("%v:%v@/%v?charset=utf8mb4&parseTime=True&loc=Local", envconfig.Username, envconfig.Password, envconfig.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
		panic("failed to connect database")
	}
	// Migrate the schema
	db.AutoMigrate(&types.Email{})
	return db
}

func Dial() *gorm.DB {
	log.Infof("dialing to target MySqlDb at: %v, database: %v", "localhost:3306", envconfig.Database)
	dsn := fmt.Sprintf("%v:%v@/%v?charset=utf8mb4&parseTime=True&loc=Local", "root", envconfig.Password, "testdb")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
		panic("failed to connect database")
	}
	// Migrate the schema
	db.AutoMigrate(&types.Email{})
	return db
}
