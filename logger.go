package goEagi

import (
	"log"
	"os"
)

type Logger struct {
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

func NewLogger(filePath string) (*Logger, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	l := Logger{
		Info:    log.New(file, "INFO:\t", log.Ldate|log.Ltime|log.Lshortfile),
		Warning: log.New(file, "WARNING:\t", log.Ldate|log.Ltime|log.Lshortfile),
		Error:   log.New(file, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile),
	}

	return &l, nil
}
