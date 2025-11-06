package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

func NewLogger() *logrus.Logger {
	log := logrus.New()

	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.JSONFormatter{})
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	return log
}
