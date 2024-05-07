package utils

import "log/slog"

type logMsg struct {
	logger  *slog.Logger
	URL     string
	Method  string
	Message string
	Status  int
}

// Возвращает структуру, которая пишет логи с помощью logger.
// Остальные поля - информация, которая будет выводиться.
func NewLogMsg(logger *slog.Logger, url, method string) *logMsg {
	return &logMsg{
		logger: logger,
		URL:    url,
		Method: method,
	}
}

func (msg *logMsg) Set(message string, status int) {
	msg.Message = message
	msg.Status = status
}

func (msg *logMsg) Info() {
	msg.logger.Info(msg.Message, getArgs(msg)...)
}

func (msg *logMsg) Error() {
	msg.logger.Error(msg.Message, getArgs(msg)...)
}

func getArgs(msg *logMsg) []any {
	return []any{
		"status", msg.Status,
		"url", msg.URL,
		"method", msg.Method,
	}
}
