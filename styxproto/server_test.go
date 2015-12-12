package styxproto

import (
	"io"
	"os"
	"sync"
	"testing"

	"golang.org/x/net/context"
)

func mkQid(qtype uint8, version uint32, path uint64) Qid {
	var buf [QidLen]byte
	q, _, err := NewQid(buf[:], qtype, version, path)
	if err != nil {
		panic(err)
	}
	return q
}

type logger interface {
	Logf(format string, v ...interface{})
}

type logWriter struct {
	once sync.Once
	w    io.Writer
	t    *testing.T
}

func (lw *logWriter) init() {
	lw.once.Do(func() {
		r, w := io.Pipe()
		lw.w = w
		go func() {
			d := NewDecoder(r)
			for d.Next() {
				for _, msg := range d.Messages() {
					lw.t.Logf("→ %d %s", msg.Tag(), msg)
				}
				if d.Err() != nil {
					lw.t.Error(d.Err())
				}
			}
		}()
	})
}

func (lw *logWriter) Write(p []byte) (int, error) {
	lw.init()
	return lw.w.Write(p)
}

// An echoServer just logs its requests
type echoServer struct {
	logger
}

func (e echoServer) Attach(w *ResponseWriter, m Tattach) {
	e.Logf("← %d %s", m.Tag(), m)
	w.Rattach(m.Tag(), mkQid(0, 3, 1830))
	w.Flush()
	w.Close()
}
func (e echoServer) Auth(w *ResponseWriter, m Tauth) {
	e.Logf("← %d %s", m.Tag(), m)
	w.Rauth(m.Tag(), mkQid(0, 18, 3458))
	w.Flush()
	w.Close()
}
func (e echoServer) Clunk(w *ResponseWriter, m Tclunk) {
	e.Logf("← %d %s", m.Tag(), m)
	w.Rclunk(m.Tag())
	w.Flush()
	w.Close()
}
func (e echoServer) Create(w *ResponseWriter, m Tcreate) {
	e.Logf("← %d %s", m.Tag(), m)
	w.Rcreate(m.Tag(), mkQid(0, 45, 381), 0)
	w.Flush()
	w.Close()
}
func (e echoServer) Open(w *ResponseWriter, m Topen) {
	e.Logf("← %d %s", m.Tag(), m)
	w.Ropen(m.Tag(), mkQid(0, 12, 4), 0)
	w.Flush()
	w.Close()
}
func (e echoServer) Read(w *ResponseWriter, m Tread) {
	e.Logf("← %d %s", m.Tag(), m)
	w.Rread(m.Tag(), []byte("hello"))
	w.Flush()
	w.Close()
}
func (e echoServer) Remove(w *ResponseWriter, m Tremove) {
	e.Logf("← %d %s", m.Tag(), m)
	w.Rremove(m.Tag())
	w.Flush()
	w.Close()
}
func (e echoServer) Stat(w *ResponseWriter, m Tstat) {
	e.Logf("← %d %s", m.Tag(), m)
	w.Rerror(m.Tag(), "no such file")
	w.Flush()
	w.Close()
}
func (e echoServer) Walk(w *ResponseWriter, m Twalk) {
	e.Logf("← %d %s", m.Tag(), m)
	w.Rerror(m.Tag(), "too lazy to mock Rwalk")
	w.Flush()
	w.Close()
}
func (e echoServer) Write(w *ResponseWriter, m Twrite) {
	e.Logf("← %d %s", m.Tag(), m)
	w.Rwrite(m.Tag(), 10)
	w.Flush()
	w.Close()
}
func (e echoServer) Wstat(w *ResponseWriter, m Twstat) {
	e.Logf("← %d %s", m.Tag(), m)
	w.Rwstat(m.Tag())
	w.Flush()
	w.Close()
}

func TestServer(t *testing.T) {
	srv := echoServer{t}
	file, err := os.Open("testdata/sample.client.9p")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	pipe := struct {
		io.Writer
		io.ReadCloser
	}{
		Writer:     &logWriter{t: t},
		ReadCloser: file,
	}
	c := NewConn(pipe, DefaultMaxSize)
	if err := Serve(c, context.Background(), srv); err != nil {
		t.Error(err)
	}
}
