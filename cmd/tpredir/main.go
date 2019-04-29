package main

import (
	"flag"
	"log"
	"net"
	"tp/tpredir"
)

var (
	proxyServer = flag.String("proxy-server", "", "proxy server address")
)

func main() {
	flag.Parse()

	l, err := net.Listen("tcp", "127.0.0.1:3000")
	if err != nil {
		log.Fatal(err)
	}
	server := &tpredir.Server{
		ProxyDial: func() (net.Conn, error) {
			return net.Dial("tcp", *proxyServer)
		},
	}
	server.Serve(l)
}
