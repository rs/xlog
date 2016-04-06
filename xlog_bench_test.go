package xlog

import "testing"

func BenchmarkSend(b *testing.B) {
	l := New(Config{Output: Discard, Fields: F{"a": "b"}}).(*logger)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.send(0, 0, "test", F{"foo": "bar", "bar": "baz"})
	}
}
