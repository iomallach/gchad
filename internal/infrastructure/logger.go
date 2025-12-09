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

func (l *ZeroLogLogger) Debug(msg string, fields map[string]any) {
	event := l.logger.Debug()

	for k, v := range fields {
		event.Interface(k, v)
	}

	event.Msg(msg)
}

func (l *ZeroLogLogger) Info(msg string, fields map[string]any) {
	event := l.logger.Info()

	for k, v := range fields {
		event.Interface(k, v)
	}

	event.Msg(msg)
}

func (l *ZeroLogLogger) Error(msg string, fields map[string]any) {
	event := l.logger.Error()

	for k, v := range fields {
		event.Interface(k, v)
	}

	event.Msg(msg)
}
