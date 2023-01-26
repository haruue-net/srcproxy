package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/haruue-net/srcproxy"
	"github.com/haruue-net/srcproxy/redirect"
	"github.com/haruue-net/srcproxy/utils"
	"io"
	"log"
	"net"
)

func runClient(config *ClientConfig) (err error) {
	serverDialer := &net.Dialer{
		Timeout: config.Timeout.Duration(),
	}
	if config.Outbound.LocalAddr != nil && config.Outbound.LocalAddr.IP != nil {
		serverDialer.LocalAddr = &net.TCPAddr{
			IP: config.Outbound.LocalAddr.IP,
		}
	}
	sockopts := utils.LinuxSockopts{}
	if config.Outbound.LocalAddr != nil && config.Outbound.LocalAddr.Device != "" {
		var iface *net.Interface
		iface, err = net.InterfaceByName(config.Outbound.LocalAddr.Device)
		if err != nil {
			err = fmt.Errorf("failed to find interface: %w", err)
			return
		}
		sockopts.BindIface = iface
	}
	if config.Outbound.FirewallMark != nil {
		sockopts.FirewallMark = config.Outbound.FirewallMark
	}
	serverDialer.Control = sockopts.Control
	dialServer := func() (conn *net.TCPConn, err error) {
		xconn, err := serverDialer.Dial("tcp", config.Outbound.Server)
		if err != nil {
			return
		}
		conn = xconn.(*net.TCPConn)
		return
	}

	auth := sha256.Sum256([]byte(config.Outbound.Auth))

	client := &srcproxy.Client{
		DialServerFunc: dialServer,
		Auth:           auth[:],
	}

	logger := &clientLogger{
		level: config.LogLevel,
	}

	switch config.Inbound.Mode {
	case ClientModeDefault:
		fallthrough
	case ClientModeRedirect:
		tcpRedirect := &redirect.TCPRedirect{
			Dialer:  client,
			Logger:  logger,
			Timeout: config.Timeout.Duration(),
		}

		var listener net.Listener
		listener, err = net.Listen("tcp", config.Inbound.Listen)
		if err != nil {
			return
		}

		log.Printf("[info] listening on %s\n", config.Inbound.Listen)

		err = tcpRedirect.ServeListener(listener)
	default:
		log.Panicf("[panic] unhandled inbound mode: %s", config.Inbound.Mode)
	}
	return
}

type clientLogger struct {
	level LogLevel
}

func (c *clientLogger) LogConn(addr, reqAddr net.Addr) {
	if c.level <= LogLevelInfo {
		log.Printf("[info] %s <-> %s: established\n", reqAddr, addr)
	}
}

func (c *clientLogger) LogError(addr, reqAddr net.Addr, err error) {
	if err == io.EOF {
		if c.level <= LogLevelInfo {
			log.Printf("[info] %s <-> %s: closed\n", reqAddr, addr)
		}
		return
	}
	if c.level <= LogLevelError {
		log.Printf("[error] %s <-> %s: %s\n", reqAddr, addr, err)
	}
}
