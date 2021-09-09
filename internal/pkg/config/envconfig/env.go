package envconfig

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/volam1999/gomail/internal/pkg/log"
)

// Load loads the environment variables into the provided struct
func Load(envPrefix string, t interface{}) {
	if err := envconfig.Process(envPrefix, t); err != nil {
		log.Errorf("config: unable to load config for %T: %s", t, err)
	}
}

// SetEnvFromFile load environments from file
func SetEnvFromFile(f string) error {
	err := godotenv.Load(f)
	if err != nil {
		return err
	}
	return nil
}
