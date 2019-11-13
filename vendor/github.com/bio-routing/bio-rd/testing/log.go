package testing

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// LogFormatter provides a log formatter for unit tests free of timestamps
type LogFormatter struct{}

// NewLogFormatter creates a new log formatter
func NewLogFormatter() *LogFormatter {
	return &LogFormatter{}
}

// Format formats a log entry
func (l *LogFormatter) Format(e *logrus.Entry) ([]byte, error) {
	var res string
	if len(e.Data) == 0 {
		res = fmt.Sprintf("level=%s msg=%q", e.Level, e.Message)
	} else {
		res = fmt.Sprintf("level=%s msg=%q fields=%v", e.Level, e.Message, e.Data)
	}
	return []byte(res), nil
}
