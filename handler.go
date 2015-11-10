package xlog

import (
	"net"
	"net/http"

	"github.com/rs/xhandler"
	"golang.org/x/net/context"
)

type key int

const (
	logKey key = iota
	idKey
)

// FromContext gets the logger out of the context.
// If not logger is stored in the context, a NopLogger is returned
func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return NopLogger
	}
	l, ok := ctx.Value(logKey).(Logger)
	if !ok {
		return NopLogger
	}
	return l
}

// NewContext returns a copy of the parent context and associates it with passed logger.
func NewContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, logKey, l)
}

// IDFromContext returns a uniq id associated to the request if any
func IDFromContext(ctx context.Context) (ID, bool) {
	id, ok := ctx.Value(idKey).(ID)
	return id, ok
}

// NewHandler instanciates a new xlog.Handler.
//
// By default, the output is set to ConsoleOutput(os.Stderr).
func NewHandler(c Config) func(xhandler.HandlerC) xhandler.HandlerC {
	if c.Output == nil {
		c.Output = NewOutputChannel(NewConsoleOutput())
	}
	return func(next xhandler.HandlerC) xhandler.HandlerC {
		return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			l := New(c)
			ctx = NewContext(ctx, l)
			next.ServeHTTPC(ctx, w, r)
			if l, ok := l.(*logger); ok {
				l.close()
			}
		})
	}
}

// RemoteAddrHandler returns a handler setting the request's remote address as a field
// to the current context's logger.
func RemoteAddrHandler(name string) func(next xhandler.HandlerC) xhandler.HandlerC {
	return func(next xhandler.HandlerC) xhandler.HandlerC {
		return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				FromContext(ctx).SetField(name, host)
			}
			next.ServeHTTPC(ctx, w, r)
		})
	}
}

// UserAgentHandler returns a handler setting the request's client's user-agent as
// a field to the current context's logger.
func UserAgentHandler(name string) func(next xhandler.HandlerC) xhandler.HandlerC {
	return func(next xhandler.HandlerC) xhandler.HandlerC {
		return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if ua := r.Header.Get("User-Agent"); ua != "" {
				FromContext(ctx).SetField(name, ua)
			}
			next.ServeHTTPC(ctx, w, r)
		})
	}
}

// RefererHandler returns a handler setting the request's referer header as
// a field to the current context's logger.
func RefererHandler(name string) func(next xhandler.HandlerC) xhandler.HandlerC {
	return func(next xhandler.HandlerC) xhandler.HandlerC {
		return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if ref := r.Header.Get("Referer"); ref != "" {
				FromContext(ctx).SetField(name, ref)
			}
			next.ServeHTTPC(ctx, w, r)
		})
	}
}

// RequestIDHandler returns a handler setting a unique id to the request which can
// be gathered using IDFromContext(). This generated id is added as a field to the
// logger and as a response header if the headerName is not an empty string.
//
// The generated id is a URL safe base64 encoded mongo object-id-like unique id.
// Mongo unique id generation algorithm has been selected as a trade-off between
// size and ease of use: UUID is less space efficient and snowflake requires machine
// configuration.
func RequestIDHandler(name, headerName string) func(next xhandler.HandlerC) xhandler.HandlerC {
	return func(next xhandler.HandlerC) xhandler.HandlerC {
		return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			id, ok := IDFromContext(ctx)
			if !ok {
				id = NewID()
				ctx = context.WithValue(ctx, idKey, id)
			}
			if name != "" {
				FromContext(ctx).SetField(name, id)
			}
			if headerName != "" {
				w.Header().Set(headerName, id.String())
			}
			next.ServeHTTPC(ctx, w, r)
		})
	}
}
