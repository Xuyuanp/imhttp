package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"

	"github.com/Xuyuanp/imhttp"
)

var flags = struct {
	frontend string
	backend  string
	certFile string
	keyFile  string
}{}

func init() {
	flag.StringVar(&flags.frontend, "fe", ":443", "server listening addr (<ip>:443 is highly recomended)")
	flag.StringVar(&flags.backend, "be", "127.0.0.1:8838", "shadowsocks-server backend")
	flag.StringVar(&flags.certFile, "cert", "/etc/certs/server.cert", "cert file")
	flag.StringVar(&flags.keyFile, "key", "/etc/certs/server.key", "key file")
	flag.Parse()
}

func main() {
	cert, err := tls.LoadX509KeyPair(flags.certFile, flags.keyFile)
	if err != nil {
		log.Fatalf("load key pair failed: %s", err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	scanner := imhttp.ScanTLS("tcp", flags.frontend, config)
	defer scanner.Close()
	for scanner.Scan() {
		conn := scanner.Conn()
		go handleConn(conn)
	}
	log.Fatal(scanner.Err())
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	data, err := ioutil.ReadAll(io.LimitReader(conn, int64(len(imhttp.FakeRequestRaw))))
	if err != nil {
		log.Printf("read fake request failed: %s", err)
		return
	}
	if bytes.Compare(data, []byte(imhttp.FakeRequestRaw)) != 0 {
		log.Printf("Fuck you: %s", conn.RemoteAddr())
		conn.Write([]byte(imhttp.FakeFuck))
		return
	}

	fmt.Fprint(conn, imhttp.FakeResponseRaw)

	ssConn, err := net.Dial("tcp", flags.backend)
	if err != nil {
		log.Printf("dial ssserver failed: %s", err)
		return
	}
	defer ssConn.Close()

	imhttp.Exchange(ssConn, conn)
}
