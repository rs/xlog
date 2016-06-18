// +build !go1.7

package xlog

import (
	"net"
	"net/http"

	"github.com/rs/xhandler"
	"github.com/rs/xid"
	"golang.org/x/net/context"
)

type key int

const (
	logKey key = iota
	idKey
)

// IDFromContext returns the unique id associated to the request if any.
func IDFromContext(ctx context.Context) (xid.ID, bool) {
	id, ok := ctx.Value(idKey).(xid.ID)
	return id, ok
}

// FromContext gets the logger out of the context.
// If not logger is stored in the context, a NopLogger is returned.
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

// NewContext returns a copy of the parent context and associates it with the provided logger.
func NewContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, logKey, l)
}

// NewHandler instanciates a new xlog HTTP handler.
//
// If not configured, the output is set to NewConsoleOutput() by default.
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

// URLHandler returns a handler setting the request's URL as a field
// to the current context's logger using the passed name as field name.
func URLHandler(name string) func(next xhandler.HandlerC) xhandler.HandlerC {
	return func(next xhandler.HandlerC) xhandler.HandlerC {
		return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			FromContext(ctx).SetField(name, r.URL.String())
			next.ServeHTTPC(ctx, w, r)
		})
	}
}

// MethodHandler returns a handler setting the request's method as a field
// to the current context's logger using the passed name as field name.
func MethodHandler(name string) func(next xhandler.HandlerC) xhandler.HandlerC {
	return func(next xhandler.HandlerC) xhandler.HandlerC {
		return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			FromContext(ctx).SetField(name, r.Method)
			next.ServeHTTPC(ctx, w, r)
		})
	}
}

// RequestHandler returns a handler setting the request's method and URL as a field
// to the current context's logger using the passed name as field name.
func RequestHandler(name string) func(next xhandler.HandlerC) xhandler.HandlerC {
	return func(next xhandler.HandlerC) xhandler.HandlerC {
		return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			FromContext(ctx).SetField(name, r.Method+" "+r.URL.String())
			next.ServeHTTPC(ctx, w, r)
		})
	}
}

// RemoteAddrHandler returns a handler setting the request's remote address as a field
// to the current context's logger using the passed name as field name.
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
// a field to the current context's logger using the passed name as field name.
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
// a field to the current context's logger using the passed name as field name.
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
// be gathered using IDFromContext(ctx). This generated id is added as a field to the
// logger using the passed name as field name. The id is also added as a response
// header if the headerName is not empty.
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
				id = xid.New()
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
