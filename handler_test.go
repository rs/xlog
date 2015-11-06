package xlog

import (
	"net/http"
	"testing"

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
	c := Config{
		Level:  LevelInfo,
		Fields: F{"foo": "bar"},
		Output: NewOutputChannel(&testOutput{}),
	}
	lh := NewHandler(c)
	h := lh(xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		l := FromContext(ctx)
		assert.NotNil(t, l)
		assert.NotEqual(t, NopLogger, l)
		if l, ok := l.(*logger); assert.True(t, ok) {
			assert.Equal(t, LevelInfo, l.level)
			assert.Equal(t, c.Output, l.output)
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
	h = NewHandler(Config{})(h)
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
	h = NewHandler(Config{})(h)
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
	h = NewHandler(Config{})(h)
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
	h = NewHandler(Config{})(h)
	h.ServeHTTPC(context.Background(), nil, r)
}
