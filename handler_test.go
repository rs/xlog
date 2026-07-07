//go:build go1.7
// +build go1.7

package xlog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromContext(t *testing.T) {
	assert.Equal(t, NopLogger, FromContext(nil))
	assert.Equal(t, NopLogger, FromContext(context.Background()))
	l := &logger{}
	ctx := NewContext(context.Background(), l)
	assert.Equal(t, l, FromContext(ctx))
}

func TestFromContextWithStd(t *testing.T) {
	assert.Equal(t, GetLogger(), FromContextWithStd(nil))
	assert.Equal(t, GetLogger(), FromContextWithStd(context.Background()))
	l := &logger{}
	ctx := NewContext(context.Background(), l)
	assert.Equal(t, l, FromContext(ctx))
}

func TestFromContextWithFields(t *testing.T) {
	fields := F{"a": "a"}
	l := GetLogger()
	expected := Copy(l)
	expected.SetFields(fields)
	assert.Equal(t, expected, FromContextWithFields(nil, fields))
	assert.Equal(t, expected, FromContextWithFields(context.Background(), fields))
	nl := &logger{}
	ctx := NewContext(context.Background(), nl)

	assert.Equal(t, &logger{fields: fields}, FromContextWithFields(ctx, fields))
}

func TestNewHandler(t *testing.T) {
	c := Config{
		Level:  LevelInfo,
		Fields: F{"foo": "bar"},
		Output: NewOutputChannel(&testOutput{}),
	}
	lh := NewHandler(c)
	h := lh(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := FromRequest(r)
		assert.NotNil(t, l)
		assert.NotEqual(t, NopLogger, l)
		if l, ok := l.(*logger); assert.True(t, ok) {
			assert.Equal(t, LevelInfo, l.level)
			assert.Equal(t, c.Output, l.output)
			assert.Equal(t, F{"foo": "bar"}, F(l.fields))
		}
	}))
	h.ServeHTTP(nil, &http.Request{})
}

func TestURLHandler(t *testing.T) {
	r := &http.Request{
		URL: &url.URL{Path: "/path", RawQuery: "foo=bar"},
	}
	h := URLHandler("url")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := FromRequest(r).(*logger)
		assert.Equal(t, F{"url": "/path?foo=bar"}, F(l.fields))
	}))
	h = NewHandler(Config{})(h)
	h.ServeHTTP(nil, r)
}

func TestMethodHandler(t *testing.T) {
	r := &http.Request{
		Method: "POST",
	}
	h := MethodHandler("method")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := FromRequest(r).(*logger)
		assert.Equal(t, F{"method": "POST"}, F(l.fields))
	}))
	h = NewHandler(Config{})(h)
	h.ServeHTTP(nil, r)
}

func TestRequestHandler(t *testing.T) {
	r := &http.Request{
		Method: "POST",
		URL:    &url.URL{Path: "/path", RawQuery: "foo=bar"},
	}
	h := RequestHandler("request")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := FromRequest(r).(*logger)
		assert.Equal(t, F{"request": "POST /path?foo=bar"}, F(l.fields))
	}))
	h = NewHandler(Config{})(h)
	h.ServeHTTP(nil, r)
}

func TestRemoteAddrHandler(t *testing.T) {
	r := &http.Request{
		RemoteAddr: "1.2.3.4:1234",
	}
	h := RemoteAddrHandler("ip")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := FromRequest(r).(*logger)
		assert.Equal(t, F{"ip": "1.2.3.4"}, F(l.fields))
	}))
	h = NewHandler(Config{})(h)
	h.ServeHTTP(nil, r)
}

func TestRemoteAddrHandlerIPv6(t *testing.T) {
	r := &http.Request{
		RemoteAddr: "[2001:db8:a0b:12f0::1]:1234",
	}
	h := RemoteAddrHandler("ip")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := FromRequest(r).(*logger)
		assert.Equal(t, F{"ip": "2001:db8:a0b:12f0::1"}, F(l.fields))
	}))
	h = NewHandler(Config{})(h)
	h.ServeHTTP(nil, r)
}

func TestUserAgentHandler(t *testing.T) {
	r := &http.Request{
		Header: http.Header{
			"User-Agent": []string{"some user agent string"},
		},
	}
	h := UserAgentHandler("ua")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := FromRequest(r).(*logger)
		assert.Equal(t, F{"ua": "some user agent string"}, F(l.fields))
	}))
	h = NewHandler(Config{})(h)
	h.ServeHTTP(nil, r)
}

func TestRefererHandler(t *testing.T) {
	r := &http.Request{
		Header: http.Header{
			"Referer": []string{"http://foo.com/bar"},
		},
	}
	h := RefererHandler("ua")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := FromRequest(r).(*logger)
		assert.Equal(t, F{"ua": "http://foo.com/bar"}, F(l.fields))
	}))
	h = NewHandler(Config{})(h)
	h.ServeHTTP(nil, r)
}

func TestRequestIDHandler(t *testing.T) {
	r := &http.Request{}
	h := RequestIDHandler("id", "Request-Id")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := FromRequest(r).(*logger)
		if id, ok := IDFromRequest(r); assert.True(t, ok) {
			assert.Equal(t, l.fields["id"], id)
			assert.Len(t, id.String(), 20)
			assert.Equal(t, id.String(), w.Header().Get("Request-Id"))
		}
		assert.Len(t, l.fields["id"], 12)
	}))
	h = NewHandler(Config{})(h)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
}
