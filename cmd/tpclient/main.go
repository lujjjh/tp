package main

import (
	"encoding/binary"
	"flag"
	"log"
	"net"
	"strconv"
	"time"
	"tp/proxy"
)

var (
	serverAddr   = flag.String("server", "", "Server address.")
	connPoolSize = flag.Int("conn-pool-size", 2, "Connection pool size.")
)

func main() {
	flag.Parse()

	for i := 0; i < *connPoolSize; i++ {
		go register()
	}
	select {}
}

func register() {
	defer func() { go register() }()

	conn, err := net.Dial("tcp", *serverAddr)
	if err != nil {
		log.Println("dial server:", err)
		time.Sleep(5 * time.Second)
		return
	}

	log.Println("reading port")
	var port uint16
	err = binary.Read(conn, binary.BigEndian, &port)
	if err != nil {
		conn.Close()
		log.Println("read port:", err)
		return
	}

	log.Println("proxying", port)
	go proxyConn(conn, int(port))
}

func proxyConn(conn net.Conn, port int) {
	defer conn.Close()
	backendConn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		log.Println("dial local:", err)
		return
	}
	defer backendConn.Close()

	proxy.Tunnel(conn, backendConn)
}
