package utils

import "net"

type TCPDialer interface {
	DialTCPWithSource(dst, src *net.TCPAddr) (*net.TCPConn, error)
}

type TCPListener interface {
	Accept() (net.Conn, error)
}
