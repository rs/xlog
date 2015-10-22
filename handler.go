package xlog

import (
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/rs/xhandler"
	"golang.org/x/net/context"
)

// Handler injects a per request Logger in the net/context which can be retrived using xlog.FromContext(ctx)
type Handler struct {
	mu     sync.Mutex
	level  Level
	input  chan map[string]interface{}
	output Output
	next   xhandler.Handler
	stop   chan struct{}
	fields map[string]interface{}
}

type key int

const logKey key = 0

var loggerPool = sync.Pool{
	New: func() interface{} {
		return &logger{}
	},
}

// FromContext gets the logger out of the context.
// If not logger is stored in the context, a nopLogger is returned
func FromContext(ctx context.Context) Logger {
	l, ok := ctx.Value(logKey).(Logger)
	if !ok {
		l = nopLogger
	}
	return l
}

// newContext restores a new context storing the given logger
func newContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, logKey, l)
}

// NewHandler instanciates a new xlog.Handler.
//
// By default, the output is set to ConsoleOutput(os.Stderr), you may change that using SetOutput().
// The logger go routine is started automatically. You may start/stop this go routine
// using Start()/Stop() methods.
func NewHandler(level Level, next xhandler.Handler) *Handler {
	h := &Handler{
		level:  level,
		input:  make(chan map[string]interface{}, 100),
		output: NewConsoleOutput(),
		next:   next,
	}
	h.Start()
	return h
}

// SetFields sets fields to append to all messages
func (h *Handler) SetFields(f map[string]interface{}) {
	h.fields = f
}

// SetOutput sets the output destination for the logs
func (h *Handler) SetOutput(o Output) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.output = o
}

// Start starts the logger go routine
func (h *Handler) Start() {
	h.mu.Lock()
	if h.stop != nil {
		// Already started
		return
	}
	h.stop = make(chan struct{})
	h.mu.Unlock()
	go func() {
		for {
			select {
			case msg := <-h.input:
				if err := h.output.Write(msg); err != nil {
					log.Printf("xlog: cannot write log message: %v", err)
				}
			case <-h.stop:
				break
			}
		}
	}()
}

// Stop stops the logger go routine
func (h *Handler) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	close(h.stop)
	h.stop = nil
}

// NewLogger manually creates a logger.
// This method should only be used out of a request. Use FromContext in request.
func (h *Handler) NewLogger() Logger {
	l := loggerPool.Get().(*logger)
	l.level = h.level
	l.output = h.input
	for k, v := range h.fields {
		l.SetField(k, v)
	}
	return l
}

// Implements xhandler.Handler interface
func (h *Handler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	l := h.NewLogger()
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		l.SetField(KeyIP, host)
	}
	if ua := r.Header.Get("User-Agent"); ua != "" {
		l.SetField(KeyUserAgent, ua)
	}
	ctx = newContext(ctx, l)
	h.next.ServeHTTP(ctx, w, r)
	if l, ok := l.(*logger); ok {
		l.output = nil
		l.fields = nil
		loggerPool.Put(l)
	}
}
