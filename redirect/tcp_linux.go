package redirect

import (
	"encoding/binary"
	"errors"
	"github.com/haruue-net/srcproxy/utils"
	"net"
	"syscall"
	"time"
)

type Logger interface {
	LogConn(addr, reqAddr net.Addr)
	LogError(addr, reqAddr net.Addr, err error)
}

type TCPRedirect struct {
	Dialer utils.TCPDialer
	Logger Logger

	Timeout time.Duration
}

func (r *TCPRedirect) logConn(addr, reqAddr net.Addr) {
	if r.Logger != nil {
		r.Logger.LogConn(addr, reqAddr)
	}
}

func (r *TCPRedirect) logError(addr, reqAddr net.Addr, err error) {
	if r.Logger != nil {
		r.Logger.LogError(addr, reqAddr, err)
	}
}

func (r *TCPRedirect) ServeListener(listener utils.TCPListener) error {
	for {
		c, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			defer c.Close()
			dest, err := getDestAddr(c.(*net.TCPConn))
			if err != nil || dest.IP.IsLoopback() {
				// Silently drop the connection if we failed to get the destination address,
				// or if it's a loopback address (not a redirected connection).
				return
			}
			r.logConn(c.RemoteAddr(), dest)
			rc, err := r.Dialer.DialTCPWithSource(dest, c.RemoteAddr().(*net.TCPAddr))
			if err != nil {
				r.logError(c.RemoteAddr(), dest, err)
				return
			}
			defer rc.Close()
			err = utils.PipePairWithTimeout(c, rc, r.Timeout)
			r.logError(c.RemoteAddr(), dest, err)
		}()
	}
}

func getDestAddr(conn *net.TCPConn) (*net.TCPAddr, error) {
	rc, err := conn.SyscallConn()
	if err != nil {
		return nil, err
	}
	var addr *sockAddr
	var err2 error
	err = rc.Control(func(fd uintptr) {
		addr, err2 = getOrigDst(fd)
	})
	if err != nil {
		return nil, err
	}
	if err2 != nil {
		return nil, err2
	}
	switch addr.family {
	case syscall.AF_INET:
		return &net.TCPAddr{IP: addr.data[:4], Port: int(binary.BigEndian.Uint16(addr.port[:]))}, nil
	case syscall.AF_INET6:
		return &net.TCPAddr{IP: addr.data[4:20], Port: int(binary.BigEndian.Uint16(addr.port[:]))}, nil
	default:
		return nil, errors.New("unknown address family")
	}
}
