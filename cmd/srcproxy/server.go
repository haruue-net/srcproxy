package main

import (
	"github.com/haruue-net/srcproxy"
	"io"
	"log"
	"net"
)

func runServer(config *ServerConfig) (err error) {
	auther := NewAccessControl(config.ACL)
	logger := &serverLogger{
		level: config.LogLevel,
	}

	listener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		return
	}

	log.Printf("[info] listening on %s\n", config.Listen)

	server := &srcproxy.Server{
		CheckSrcFunc: auther.CheckSrc,
		Logger:       logger,
		Timeout:      config.Timeout.Duration(),
	}

	return server.ServeListener(listener)
}

type serverLogger struct {
	level LogLevel
}

func (s *serverLogger) LogConn(reqAddr net.Addr, src, dst *net.TCPAddr) {
	if s.level <= LogLevelInfo {
		log.Printf("[info] (%s) %s <-> %s: established\n", reqAddr, src, dst)
	}
}

func (s *serverLogger) LogError(reqAddr net.Addr, src, dst *net.TCPAddr, err error) {
	if err == io.EOF {
		if s.level <= LogLevelInfo {
			log.Printf("[info] (%s) %s <-> %s: closed\n", reqAddr, src, dst)
		}
		return
	}
	if s.level <= LogLevelError {
		log.Printf("[error] (%s) %s <-> %s: %s\n", reqAddr, src, dst, err)
	}
}
