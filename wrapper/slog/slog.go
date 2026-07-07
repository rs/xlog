package slog

import (
	"context"
	"log/slog"
	"strings"

	"github.com/Ak-Army/xlog"
)

// Wrapper implements slog.Handler by delegating to an xlog.Logger.
type Wrapper struct {
	logger xlog.Logger
	attrs  []slog.Attr
	groups []string
}

func NewSlogger(logger xlog.Logger) *slog.Logger {
	return slog.New(&Wrapper{logger: logger})
}

func NewWrapper(logger xlog.Logger) *Wrapper {
	return &Wrapper{logger: logger}
}

func (h *Wrapper) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *Wrapper) Handle(_ context.Context, r slog.Record) error {
	fields := xlog.F{}
	for _, a := range h.attrs {
		fields[strings.Join(append(h.groups, a.Key), ".")] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})
	switch r.Level {
	case slog.LevelError:
		h.logger.Error(r.Message, fields)
	case slog.LevelWarn:
		h.logger.Warn(r.Message, fields)
	case slog.LevelDebug:
		h.logger.Debug(r.Message, fields)
	default:
		h.logger.Info(r.Message, fields)
	}
	return nil
}

func (h *Wrapper) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &Wrapper{logger: h.logger, attrs: newAttrs, groups: h.groups}
}

func (h *Wrapper) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	return &Wrapper{logger: h.logger, attrs: h.attrs, groups: newGroups}
}
