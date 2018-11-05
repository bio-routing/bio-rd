package grpchelper

import (
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
)

// CodeToLogrusLevel is the default 0-stor implementation
// of gRPC return codes to log(rus) levels for server side.
func codeToLogrusLevel(code codes.Code) log.Level {
	if level, ok := _GRPCCodeToLogrusLevelMapping[code]; ok {
		return level
	}
	return log.ErrorLevel
}

var _GRPCCodeToLogrusLevelMapping = map[codes.Code]log.Level{
	codes.OK:                 log.DebugLevel,
	codes.Canceled:           log.DebugLevel,
	codes.InvalidArgument:    log.DebugLevel,
	codes.NotFound:           log.DebugLevel,
	codes.AlreadyExists:      log.DebugLevel,
	codes.Unauthenticated:    log.InfoLevel,
	codes.PermissionDenied:   log.InfoLevel,
	codes.DeadlineExceeded:   log.WarnLevel,
	codes.ResourceExhausted:  log.WarnLevel,
	codes.FailedPrecondition: log.WarnLevel,
	codes.Aborted:            log.WarnLevel,
}
