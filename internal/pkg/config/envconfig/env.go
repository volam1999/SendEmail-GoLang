package envconfig

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	Username = os.Getenv("MYSQL_USERNAME")
	Password = os.Getenv("MYSQL_PASSWORD")
	Database = os.Getenv("MYSQL_DATABASE")
)

func Error(filenames ...string) error {
	return godotenv.Load(filenames...)
}
