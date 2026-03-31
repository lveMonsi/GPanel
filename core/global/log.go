package global

import (
	"io"
	"log"
	"os"
)

var logHandle *os.File

func InitLogger() error {
	if err := os.MkdirAll(LogDir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	logHandle = file
	log.SetOutput(io.MultiWriter(os.Stdout, file))
	return nil
}

func CloseLogger() {
	if logHandle != nil {
		_ = logHandle.Close()
		logHandle = nil
	}
}
