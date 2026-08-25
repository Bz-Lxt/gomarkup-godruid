package tcp_test

import (
	"net"
	"testing"

	"godruid/internal/adapter"
	"godruid/internal/adapter/tcp"
)

func TestTCPContract(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				buf := make([]byte, 16)
				for {
					if _, err := conn.Read(buf); err != nil {
						_ = conn.Close()
						return
					}
				}
			}(c)
		}
	}()
	adapter.CheckConnector(t, tcp.New(ln.Addr().String()))
}
