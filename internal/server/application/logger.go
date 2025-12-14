package application

type Logger interface {
	Debug(msg string, fields map[string]any)
	Info(msg string, fields map[string]any)
	Error(msg string, fields map[string]any)
}
