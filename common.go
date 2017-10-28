package imhttp

import (
	"crypto/tls"
	"io"
	"net"
)

const (
	FakeRequestRaw  = "POST /fuck HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n"
	FakeResponseRaw = "HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok"
	FakeFuck        = "HTTP/1.1 401 OK\r\nContent-Length: 12\r\n\r\nUnauthorized"
)

type ListenerScanner struct {
	ln   net.Listener
	err  error
	conn net.Conn
}

func (s *ListenerScanner) Scan() (ok bool) {
	if s.err != nil {
		return false
	}
	s.conn, s.err = s.ln.Accept()
	return s.err == nil
}

func (s *ListenerScanner) Conn() net.Conn {
	return s.conn
}

func (s *ListenerScanner) Err() error {
	return s.err
}

func (s *ListenerScanner) Close() error {
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}

func ScanNet(network, addr string) *ListenerScanner {
	scanner := &ListenerScanner{}
	scanner.ln, scanner.err = net.Listen(network, addr)
	return scanner
}

func ScanTLS(network, addr string, config *tls.Config) *ListenerScanner {
	scanner := &ListenerScanner{}
	scanner.ln, scanner.err = tls.Listen(network, addr, config)
	return scanner
}

func ProxyCopy(errc chan<- error, dst io.Writer, src io.Reader) {
	_, err := io.Copy(dst, src)
	errc <- err
}

func Exchange(rw1, rw2 io.ReadWriter) error {
	errc := make(chan error, 1)
	defer func() {
		go func() {
			<-errc
			close(errc)
		}()
	}()
	go ProxyCopy(errc, rw1, rw2)
	go ProxyCopy(errc, rw2, rw1)
	return <-errc
}
