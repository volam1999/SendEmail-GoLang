package log

import (
	"os"

	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "02/01/2006 15:04:05",
		FullTimestamp:   true,
	})
	log.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
}

// Infof print info with format.
func Infof(format string, v ...interface{}) {
	log.Infof(format, v...)
}

func Info(v ...interface{}) {
	log.Info(v...)
}

func Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}

func Fatal(v ...interface{}) {
	log.Fatal(v)
}

func Warnf(format string, v ...interface{}) {
	log.Warnf(format, v...)
}

func Warn(v ...interface{}) {
	log.Warn(v)
}

func Errorf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}

func Error(v ...interface{}) {
	log.Error(v)
}
