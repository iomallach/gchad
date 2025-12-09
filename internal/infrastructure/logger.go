package infrastructure

import "github.com/rs/zerolog"

type ZeroLogLogger struct {
	logger zerolog.Logger
}

func NewZeroLogLogger(logger zerolog.Logger) *ZeroLogLogger {
	return &ZeroLogLogger{
		logger,
	}
}

func (l *ZeroLogLogger) Debug() *zerolog.Event {
	return l.logger.Debug()
}

func (l *ZeroLogLogger) Info() *zerolog.Event {
	return l.logger.Info()
}

func (l *ZeroLogLogger) Error() *zerolog.Event {
	return l.logger.Error()
}
