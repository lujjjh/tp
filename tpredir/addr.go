package tpredir

import (
	"encoding/binary"
	"net"
	"syscall"
)

type wrappedConn struct {
	net.Conn

	remoteAddr net.Addr
}

func (c *wrappedConn) RemoteAddr() net.Addr { return c.remoteAddr }

func withOriginalDst(conn net.Conn) net.Conn {
	defer conn.Close()

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return conn
	}
	f, err := tcpConn.File()
	if err != nil {
		return conn
	}
	defer f.Close()

	req, err := syscall.GetsockoptIPv6Mreq(int(f.Fd()), syscall.IPPROTO_IP, 80)
	if err != nil {
		return conn
	}
	port := int(binary.BigEndian.Uint16(req.Multiaddr[2:4]))
	// TODO byte order?
	ip := net.IP(req.Multiaddr[4:8])

	newConn, err := net.FileConn(f)
	if err != nil {
		return conn
	}

	return &wrappedConn{
		Conn:       newConn,
		remoteAddr: &net.TCPAddr{IP: ip, Port: port},
	}
}
