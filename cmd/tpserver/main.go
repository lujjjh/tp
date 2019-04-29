package main

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"sync"
	"tp/proxy"
)

var (
	// TODO add TTL for each connection.
	connPool   = make(map[string][]net.Conn)
	connPoolMu sync.Mutex
)

func main() {
	go listenAndServeProxy()
	go listenAndServeRegistry()
	select {}
}

func listenAndServeProxy() {
	l, err := net.Listen("tcp", ":3001")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go handleProxyConn(conn)
	}
}

func shiftConn(ip string) net.Conn {
	connPoolMu.Lock()
	defer connPoolMu.Unlock()
	conns := connPool[ip]
	if len(conns) == 0 {
		return nil
	}
	conn := conns[0]
	connPool[ip] = conns[1:]
	return conn
}

func handleProxyConn(conn net.Conn) {
	defer conn.Close()

	var addrLen uint32
	err := binary.Read(conn, binary.BigEndian, &addrLen)
	if err != nil {
		return
	}
	if addrLen > 0xFF {
		log.Println("invalid addrLen:", addrLen)
		return
	}
	addrBytes := make([]byte, addrLen)
	_, err = io.ReadFull(conn, addrBytes)
	if err != nil {
		return
	}
	addr, err := net.ResolveTCPAddr("tcp", string(addrBytes))
	if err != nil {
		log.Println("invalid addr:", addrBytes)
		return
	}

	ip := addr.IP.String()
	var backendConn net.Conn
	for {
		backendConn = shiftConn(ip)
		if backendConn == nil {
			log.Println("no available conn:", ip)
			return
		}
		log.Println("writing port to", backendConn.RemoteAddr())
		err := binary.Write(backendConn, binary.BigEndian, uint16(addr.Port))
		if err != nil {
			log.Println("try next registered server:", err)
			continue
		}
		break
	}
	defer backendConn.Close()

	proxy.Tunnel(conn, backendConn)
}

func listenAndServeRegistry() {
	l, err := net.Listen("tcp", ":3002")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go handleRegistryConn(conn)
	}
}

func handleRegistryConn(conn net.Conn) {
	remoteIP := conn.RemoteAddr().(*net.TCPAddr).IP.String()
	log.Println("register", remoteIP)
	connPoolMu.Lock()
	connPool[remoteIP] = append(connPool[remoteIP], conn)
	connPoolMu.Unlock()
}
