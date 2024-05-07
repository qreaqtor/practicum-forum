package utils

import (
	"context"
	"log/slog"
)

/*
Пришлось сделать так, потому что я не нашел как указать io.Discard в качестве вывода дефолтному логеру из slog.Logger,
а переделывать на другой логгер как-то не особо хотелось, зато теперь буду знать про этот нюанс.
*/
type DummyLogger struct{}

func (h DummyLogger) Enabled(context.Context, slog.Level) bool {
	return false
}

func (h DummyLogger) Handle(context.Context, slog.Record) error {
	return nil
}

func (h DummyLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h DummyLogger) WithGroup(name string) slog.Handler {
	return h
}
