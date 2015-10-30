package xlog

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/rs/xhandler"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestFromContext(t *testing.T) {
	assert.Equal(t, NopLogger, FromContext(nil))
	assert.Equal(t, NopLogger, FromContext(context.Background()))
	l := &logger{}
	ctx := NewContext(context.Background(), l)
	assert.Equal(t, l, FromContext(ctx))
}

func TestNewHandler(t *testing.T) {
	h := NewHandler(LevelDebug)
	defer h.Stop()
	assert.Equal(t, LevelDebug, h.level)
	assert.Equal(t, 100, cap(h.input))
}

func TestStop(t *testing.T) {
	h := NewHandler(LevelDebug)
	assert.NotNil(t, h.stop)
	h.Stop()
	assert.Nil(t, h.stop)
}

func TestSetFields(t *testing.T) {
	h := NewHandler(0)
	defer h.Stop()
	h.SetFields(F{"foo": "bar", "bar": "baz"})
	assert.Equal(t, F{"foo": "bar", "bar": "baz"}, F(h.fields))
}

func TestOutput(t *testing.T) {
	h := NewHandler(0)
	defer h.Stop()
	h.SetOutput(Discard)
	assert.Equal(t, Discard, h.output)
}

func TestChannel(t *testing.T) {
	h := NewHandler(0)
	defer h.Stop()
	o := &testOutput{}
	h.SetOutput(o)
	h.input <- F{"foo": "bar"}
	assert.Nil(t, o.last)
	runtime.Gosched()
	assert.Equal(t, F{"foo": "bar"}, F(o.last))

	// Trigger error path
	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)
	o.err = errors.New("some error")
	h.input <- F{"foo": "bar"}
	// Wait for log output to go through
	runtime.Gosched()
	for i := 0; i < 10 && buf.Len() == 0; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	assert.Contains(t, buf.String(), "xlog: cannot write log message: some error")

	// trigger already started path
	h.Start()
}

func TestNewLogger(t *testing.T) {
	h := NewHandler(LevelInfo)
	h.SetFields(F{"foo": "bar"})
	l, ok := h.NewLogger().(*logger)
	if assert.True(t, ok) {
		assert.Equal(t, LevelInfo, l.level)
		assert.Equal(t, h.input, l.output)
		assert.Equal(t, F{"foo": "bar"}, F(l.fields))
		// Ensure l.fields is a clone
		h.fields["bar"] = "baz"
		assert.Equal(t, F{"foo": "bar"}, F(l.fields))
	}
}

func TestServeHTTPC(t *testing.T) {
	lh := NewHandler(LevelInfo)
	lh.SetFields(F{"foo": "bar"})
	h := lh.Handle(xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		l := FromContext(ctx)
		assert.NotNil(t, l)
		assert.NotEqual(t, NopLogger, l)
		if l, ok := l.(*logger); assert.True(t, ok) {
			assert.Equal(t, LevelInfo, l.level)
			assert.Equal(t, lh.input, l.output)
			assert.Equal(t, F{"foo": "bar"}, F(l.fields))
		}
	}))
	h.ServeHTTPC(context.Background(), nil, nil)
}

func TestRemoteAddrHandler(t *testing.T) {
	r := &http.Request{
		RemoteAddr: "1.2.3.4:1234",
	}
	h := RemoteAddrHandler("ip")(xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		l := FromContext(ctx).(*logger)
		assert.Equal(t, F{"ip": "1.2.3.4"}, F(l.fields))
	}))
	h = NewHandler(LevelInfo).Handle(h)
	h.ServeHTTPC(context.Background(), nil, r)
}

func TestRemoteAddrHandlerIPv6(t *testing.T) {
	r := &http.Request{
		RemoteAddr: "[2001:db8:a0b:12f0::1]:1234",
	}
	h := RemoteAddrHandler("ip")(xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		l := FromContext(ctx).(*logger)
		assert.Equal(t, F{"ip": "2001:db8:a0b:12f0::1"}, F(l.fields))
	}))
	h = NewHandler(LevelInfo).Handle(h)
	h.ServeHTTPC(context.Background(), nil, r)
}

func TestUserAgentHandler(t *testing.T) {
	r := &http.Request{
		Header: http.Header{
			"User-Agent": []string{"some user agent string"},
		},
	}
	h := UserAgentHandler("ua")(xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		l := FromContext(ctx).(*logger)
		assert.Equal(t, F{"ua": "some user agent string"}, F(l.fields))
	}))
	h = NewHandler(LevelInfo).Handle(h)
	h.ServeHTTPC(context.Background(), nil, r)
}

func TestRefererHandler(t *testing.T) {
	r := &http.Request{
		Header: http.Header{
			"Referer": []string{"http://foo.com/bar"},
		},
	}
	h := RefererHandler("ua")(xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		l := FromContext(ctx).(*logger)
		assert.Equal(t, F{"ua": "http://foo.com/bar"}, F(l.fields))
	}))
	h = NewHandler(LevelInfo).Handle(h)
	h.ServeHTTPC(context.Background(), nil, r)
}
