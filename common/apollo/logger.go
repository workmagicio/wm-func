package apollo

import "log"

type logger struct {
	log *log.Logger
}

func (l *logger) Infof(format string, args ...interface{}) {
	//l.log.Printf("[INFO] "+format, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.log.Printf("[ERROR] "+format, args...)
}
