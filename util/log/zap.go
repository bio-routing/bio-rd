package log

import (
	"go.uber.org/zap"
)

func init() {
	logger = NewZapWrapper(zap.L())
}

type zapWrapper struct {
	logger *zap.SugaredLogger
}

func NewZapWrapper(l *zap.Logger) LoggerInterface {
	return zapWrapper{logger: l.Sugar()}
}

func (zw zapWrapper) Errorf(format string, args ...interface{}) {
	zw.logger.Errorf(format, args...)
}
func (zw zapWrapper) Infof(format string, args ...interface{}) {
	zw.logger.Infof(format, args...)
}
func (zw zapWrapper) Debugf(format string, args ...interface{}) {
	zw.logger.Debugf(format, args...)
}
func (zw zapWrapper) Error(msg string) {
	zw.logger.Error(msg)
}
func (zw zapWrapper) Info(msg string) {
	zw.logger.Info(msg)
}
func (zw zapWrapper) Debug(msg string) {
	zw.logger.Debug(msg)
}
func (zw zapWrapper) WithFields(fields Fields) LoggerInterface {
	l := zw.logger
	for k, v := range fields {
		l = l.WithOptions(zap.Fields(zap.Any(k, v)))
	}
	return zapWrapper{logger: l}
}
func (zw zapWrapper) WithError(err error) LoggerInterface {
	zw.logger.Error(err)
	return zapWrapper{logger: zw.logger}
}
