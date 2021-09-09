package mysqldb

import (
	"fmt"

	"github.com/volam1999/gomail/internal/app/types"
	"github.com/volam1999/gomail/internal/pkg/config/envconfig"
	"github.com/volam1999/gomail/internal/pkg/log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type (
	// Config hold MongoDB configuration information
	Config struct {
		Addrs    string `envconfig:"MYSQL_ADDRS" default:"127.0.0.1:3306"`
		Database string `envconfig:"MYSQL_DATABASE" default:"goway"`
		Username string `envconfig:"MYSQL_USERNAME"`
		Password string `envconfig:"MYSQL_PASSWORD"`
	}
)

// LoadConfigFromEnv load mongodb configurations from environments
func LoadConfigFromEnv() *Config {
	var conf Config
	envconfig.Load("", &conf)
	return &conf
}

func MustNew(config *Config) *gorm.DB {
	log.Infof("dialing to target MySqlDb at: %v, database: %v", config.Addrs, config.Database)
	dsn := fmt.Sprintf("%v:%v@%v/%v?charset=utf8mb4&parseTime=True&loc=Local", config.Username, config.Password, config.Addrs, config.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
		panic("failed to connect database")
	}
	// Migrate the schema
	db.AutoMigrate(&types.Email{})
	return db
}
