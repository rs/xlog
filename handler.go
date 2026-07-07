//go:build go1.7
// +build go1.7

package xlog

import (
	"context"
	"net"
	"net/http"

	"github.com/rs/xid"
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

// IDFromRequest returns the unique id accociated to the request if any.
func IDFromRequest(r *http.Request) (xid.ID, bool) {
	if r == nil {
		return xid.ID{}, false
	}
	return IDFromContext(r.Context())
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

// FromContextWithStd gets the logger out of the context.
// If not logger is stored in the context, a GetLogger is returned.
func FromContextWithStd(ctx context.Context) Logger {
	if ctx == nil {
		return GetLogger()
	}
	l, ok := ctx.Value(logKey).(Logger)
	if !ok {
		return GetLogger()
	}
	return l
}

// FromContextWithFields gets the logger out of the context.
// If not logger is stored in the context, a GetLogger is returned.
func FromContextWithFields(ctx context.Context, fields F) Logger {
	l := FromContextWithStd(ctx)
	l = Copy(l)
	l.SetFields(fields)

	return l
}

// FromRequest gets the logger in the request's context.
// This is a shortcut for xlog.FromContext(r.Context())
func FromRequest(r *http.Request) Logger {
	if r == nil {
		return NopLogger
	}
	return FromContext(r.Context())
}

// NewContext returns a copy of the parent context and associates it with the provided logger.
func NewContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, logKey, l)
}

// NewHandler instanciates a new xlog HTTP handler.
//
// If not configured, the output is set to NewConsoleOutput() by default.
func NewHandler(c Config) func(http.Handler) http.Handler {
	if c.Output == nil {
		c.Output = NewOutputChannel(NewConsoleOutput())
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var l Logger
			if r != nil {
				l = New(c)
				r = r.WithContext(NewContext(r.Context(), l))
			}
			next.ServeHTTP(w, r)
			if l, ok := l.(*logger); ok {
				l.close()
			}
		})
	}
}

// URLHandler returns a handler setting the request's URL as a field
// to the current context's logger using the passed name as field name.
func URLHandler(name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := FromContext(r.Context())
			l.SetField(name, r.URL.String())
			next.ServeHTTP(w, r)
		})
	}
}

// MethodHandler returns a handler setting the request's method as a field
// to the current context's logger using the passed name as field name.
func MethodHandler(name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := FromContext(r.Context())
			l.SetField(name, r.Method)
			next.ServeHTTP(w, r)
		})
	}
}

// RequestHandler returns a handler setting the request's method and URL as a field
// to the current context's logger using the passed name as field name.
func RequestHandler(name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := FromContext(r.Context())
			l.SetField(name, r.Method+" "+r.URL.String())
			next.ServeHTTP(w, r)
		})
	}
}

// RemoteAddrHandler returns a handler setting the request's remote address as a field
// to the current context's logger using the passed name as field name.
func RemoteAddrHandler(name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				l := FromContext(r.Context())
				l.SetField(name, host)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// UserAgentHandler returns a handler setting the request's client's user-agent as
// a field to the current context's logger using the passed name as field name.
func UserAgentHandler(name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ua := r.Header.Get("User-Agent"); ua != "" {
				l := FromContext(r.Context())
				l.SetField(name, ua)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RefererHandler returns a handler setting the request's referer header as
// a field to the current context's logger using the passed name as field name.
func RefererHandler(name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ref := r.Header.Get("Referer"); ref != "" {
				l := FromContext(r.Context())
				l.SetField(name, ref)
			}
			next.ServeHTTP(w, r)
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
func RequestIDHandler(name, headerName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			id, ok := IDFromContext(ctx)
			if !ok {
				id = xid.New()
				ctx = context.WithValue(ctx, idKey, id)
				r = r.WithContext(ctx)
			}
			if name != "" {
				FromContext(ctx).SetField(name, id)
			}
			if headerName != "" {
				w.Header().Set(headerName, id.String())
			}
			next.ServeHTTP(w, r)
		})
	}
}
