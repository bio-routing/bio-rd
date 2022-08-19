package log

import "github.com/sirupsen/logrus"

func init() {
	logger = NewLogrusWrapper(logrus.New())
}

type logrusWrapper struct {
	logger *logrus.Entry
}

func NewLogrusWrapper(l *logrus.Logger) LoggerInterface {
	return logrusWrapper{logrus.NewEntry(l)}
}

func (lw logrusWrapper) Errorf(format string, args ...interface{}) {
	lw.logger.Errorf(format, args...)
}
func (lw logrusWrapper) Infof(format string, args ...interface{}) {
	lw.logger.Infof(format, args...)
}
func (lw logrusWrapper) Debugf(format string, args ...interface{}) {
	lw.logger.Debugf(format, args...)
}
func (lw logrusWrapper) Error(msg string) {
	lw.logger.Error(msg)
}
func (lw logrusWrapper) Info(msg string) {
	lw.logger.Info(msg)
}
func (lw logrusWrapper) Debug(msg string) {
	lw.logger.Debug(msg)
}
func (lw logrusWrapper) WithFields(fields Fields) LoggerInterface {
	return logrusWrapper{lw.logger.WithFields(logrus.Fields(fields))}
}
func (lw logrusWrapper) WithError(err error) LoggerInterface {
	return logrusWrapper{lw.logger.WithError(err)}
}
