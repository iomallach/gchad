package infrastructure_test

type LogCall struct {
	msg    string
	fields map[string]any
	level  string
}

type SpyLogger struct {
	calls []LogCall
}

func NewSpyLogger() *SpyLogger {
	return &SpyLogger{calls: make([]LogCall, 0)}
}

func (l *SpyLogger) Error(msg string, fields map[string]any) {
	l.calls = append(l.calls, LogCall{msg: msg, fields: fields, level: "ERROR"})
}
func (l *SpyLogger) Info(msg string, fields map[string]any) {
	l.calls = append(l.calls, LogCall{msg: msg, fields: fields, level: "INFO"})
}
func (l *SpyLogger) Debug(msg string, fields map[string]any) {
	l.calls = append(l.calls, LogCall{msg: msg, fields: fields, level: "DEBUG"})
}

func (l *SpyLogger) Errors() []LogCall {
	errors := make([]LogCall, 0)

	for _, call := range l.calls {
		if call.level == "ERROR" {
			errors = append(errors, call)
		}
	}

	return errors
}

func (l *SpyLogger) Debugs() []LogCall {
	errors := make([]LogCall, 0)

	for _, call := range l.calls {
		if call.level == "DEBUG" {
			errors = append(errors, call)
		}
	}

	return errors
}
