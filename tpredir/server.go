package tpredir

import (
	"encoding/binary"
	"log"
	"net"
	"tp/proxy"
)

type Server struct {
	ProxyDial func() (net.Conn, error)
}

func (s *Server) Serve(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go s.serveConn(withOriginalDst(conn))
	}
}

func (s *Server) serveConn(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr()
	log.Println("redir:", remoteAddr)
	backendConn, err := s.ProxyDial()
	if err != nil {
		log.Println("dial proxy:", err)
		return
	}
	defer backendConn.Close()

	remoteAddrBytes := ([]byte)(remoteAddr.String())
	err = binary.Write(backendConn, binary.BigEndian, uint32(len(remoteAddrBytes)))
	if err != nil {
		return
	}
	_, err = backendConn.Write(remoteAddrBytes)
	if err != nil {
		return
	}

	proxy.Tunnel(conn, backendConn)
}
