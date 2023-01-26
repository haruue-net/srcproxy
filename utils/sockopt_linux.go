package utils

import (
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

type LinuxSockopts struct {
	BindIface    *net.Interface
	FirewallMark *uint32
}

func (o *LinuxSockopts) Control(network, address string, c syscall.RawConn) (err error) {
	return o.bindRawConn(c)
}

func (o *LinuxSockopts) bindRawConn(c syscall.RawConn) (err error) {
	var innerErr error
	outerErr := c.Control(func(fd uintptr) {
		if o.BindIface != nil {
			innerErr = unix.BindToDevice(int(fd), o.BindIface.Name)
			if innerErr != nil {
				return
			}
		}
		if o.FirewallMark != nil {
			innerErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_MARK, int(*o.FirewallMark))
			if innerErr != nil {
				return
			}
		}
	})
	if outerErr != nil {
		return outerErr
	} else {
		return innerErr
	}
}
