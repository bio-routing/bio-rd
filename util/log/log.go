package log

var logger LoggerInterface

// LoggerInterface is the interface used to abstract logging.
type LoggerInterface interface {
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Error(msg string)
	Info(msg string)
	Debug(msg string)
	WithFields(fields Fields) LoggerInterface
	WithError(err error) LoggerInterface
}

// Fields type, used to pass to `WithFields`.
type Fields map[string]interface{}

// SetLogger changes the global logger to the one provided.
func SetLogger(l LoggerInterface) {
	logger = l
}

// Module-passthrough functions

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}
func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}
func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}
func Error(msg string) {
	logger.Error(msg)
}
func Info(msg string) {
	logger.Info(msg)
}
func Debug(msg string) {
	logger.Debug(msg)
}
func WithFields(fields Fields) LoggerInterface {
	return logger.WithFields(fields)
}
func WithError(err error) LoggerInterface {
	return logger.WithError(err)
}
