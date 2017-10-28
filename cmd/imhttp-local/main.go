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
	"os"

	"github.com/Xuyuanp/imhttp"
)

var flags = struct {
	frontend string
	backend  string
	failFast bool
}{}

func init() {
	flag.StringVar(&flags.frontend, "fe", ":9093", "frontend")
	flag.StringVar(&flags.backend, "be", "", "backend")
	flag.Parse()

	if flags.backend == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	scanner := imhttp.ScanNet("tcp", ":9091")
	defer scanner.Close()
	for scanner.Scan() {
		conn := scanner.Conn()
		go handleConn(conn)
	}
	log.Fatal(scanner.Err())
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	remote, err := tls.Dial("tcp", flags.backend, nil)
	if err != nil {
		log.Printf("dial remote server failed: %s", err)
		return
	}
	defer remote.Close()

	fmt.Fprint(remote, imhttp.FakeRequestRaw)

	data, err := ioutil.ReadAll(io.LimitReader(remote, int64(len(imhttp.FakeResponseRaw))))
	if err != nil {
		log.Printf("read fake response failed: %s", err)
		return
	}
	if bytes.Compare(data, []byte(imhttp.FakeResponseRaw)) != 0 {
		log.Printf("invalid remote server")
		return
	}

	imhttp.Exchange(remote, conn)
}
