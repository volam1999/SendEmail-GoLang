package envconfig

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	Error    = godotenv.Load("configs/config.env")
	Username = os.Getenv("MYSQL_USERNAME")
	Password = os.Getenv("MYSQL_PASSWORD")
	Database = os.Getenv("MYSQL_DATABASE")
)
