package redis_test

import (
	"bufio"
	"net"
	"testing"

	"godruid/internal/adapter"
	"godruid/internal/adapter/redis"
)

func TestRedisContract(t *testing.T) {
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
				defer conn.Close()
				br := bufio.NewReader(conn)
				for {
					if _, err := br.ReadString('\n'); err != nil {
						return
					}
					_, _ = conn.Write([]byte("+PONG\r\n"))
				}
			}(c)
		}
	}()
	adapter.CheckConnector(t, redis.New(ln.Addr().String()))
}
