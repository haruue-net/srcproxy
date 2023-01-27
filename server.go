package srcproxy

import (
	"fmt"
	"github.com/haruue-net/srcproxy/utils"
	"github.com/lunixbochs/struc"
	"net"
	"time"
)

type Server struct {
	CheckDstFunc      CheckAddrFunc
	CheckSrcFunc      CheckAddrFunc
	Logger            ServerLogger
	Timeout           time.Duration
	DialerCreatedFunc func(dialer *net.Dialer)
}

type CheckAddrFunc = func(origAddr *net.TCPAddr, auth []byte) (modifiedAddr *net.TCPAddr, err error)

var sockopts = utils.LinuxSockopts{FreeBind: true}

type ServerLogger interface {
	LogConn(reqAddr net.Addr, src, dst *net.TCPAddr)
	LogError(reqAddr net.Addr, src, dst *net.TCPAddr, err error)
}

func (s *Server) logConn(conn net.Conn, src, dst *net.TCPAddr) {
	if s.Logger != nil {
		s.Logger.LogConn(conn.RemoteAddr(), src, dst)
	}
}

func (s *Server) logError(conn net.Conn, src, dst *net.TCPAddr, err error) {
	if s.Logger != nil {
		s.Logger.LogError(conn.RemoteAddr(), src, dst, err)
	}
}

func (s *Server) checkSrc(origSrc *net.TCPAddr, auth []byte) (modifiedSrc *net.TCPAddr, err error) {
	if s.CheckSrcFunc != nil {
		return s.CheckSrcFunc(origSrc, auth)
	}
	modifiedSrc = origSrc
	return
}

func (s *Server) checkDst(origDst *net.TCPAddr, auth []byte) (modifiedDst *net.TCPAddr, err error) {
	if s.CheckDstFunc != nil {
		return s.CheckDstFunc(origDst, auth)
	}
	modifiedDst = origDst
	return
}

func (s *Server) ServeListener(listener utils.TCPListener) error {
	for {
		c, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			defer c.Close()
			s.handleConn(c)
		}()
	}
}

func (s *Server) handleConn(conn net.Conn) {
	var err error
	defer func() {
		if err != nil {
			s.logError(conn, nil, nil, err)
		}
	}()

	var req RelayRequest
	err = struc.Unpack(conn, &req)
	if err != nil {
		return
	}

	if req.Version != Version {
		err = fmt.Errorf("unsupported version: %d", req.Version)
		return
	}

	switch req.OpCode {
	case OpCodeRelayRequest:
		s.handleRelayRequest(conn, &req)
	default:
		err = fmt.Errorf("unsupported opcode: %d", req.OpCode)
	}
}

func (s *Server) handleRelayRequest(conn net.Conn, req *RelayRequest) {
	var err error
	var src, dst *net.TCPAddr

	defer func() {
		if err != nil {
			s.logError(conn, src, dst, err)
		}
	}()

	if req.SrcAddrLen != net.IPv4len && req.SrcAddrLen != net.IPv6len {
		err = fmt.Errorf("invalid src addr len: %d", req.SrcAddrLen)
		return
	}
	if req.DstAddrLen != net.IPv4len && req.DstAddrLen != net.IPv6len {
		err = fmt.Errorf("invalid dst addr len: %d", req.DstAddrLen)
		return
	}

	src = &net.TCPAddr{
		IP:   req.SrcAddr,
		Port: int(req.SrcPort),
	}
	msrc, err := s.checkSrc(src, req.Auth)
	if err != nil {
		err = fmt.Errorf("rejected by src checker: %w", err)
		return
	}

	dst = &net.TCPAddr{
		IP:   req.DstAddr,
		Port: int(req.DstPort),
	}
	mdst, err := s.checkDst(dst, req.Auth)
	if err != nil {
		err = fmt.Errorf("rejected by dst checker: %w", err)
		return
	}

	src = msrc
	dst = mdst

	dialer := &net.Dialer{
		LocalAddr: src,
		Timeout:   s.Timeout,
		Control:   sockopts.Control,
	}
	if s.DialerCreatedFunc != nil {
		s.DialerCreatedFunc(dialer)
	}

	dstConn, err := dialer.Dial("tcp", dst.String())
	if err != nil {
		return
	}
	defer dstConn.Close()

	s.logConn(conn, src, dst)

	err = utils.PipePairWithTimeout(conn, dstConn, s.Timeout)
}
