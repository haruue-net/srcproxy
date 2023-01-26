package srcproxy

import (
	"github.com/lunixbochs/struc"
	"net"
)

type Client struct {
	DialServerFunc func() (*net.TCPConn, error)
	Auth           []byte
}

func (c *Client) DialTCPWithSource(dst, src *net.TCPAddr) (conn *net.TCPConn, err error) {
	conn, err = c.DialServerFunc()
	if err != nil {
		return
	}
	err = conn.SetKeepAlive(true)

	err = struc.Pack(conn, &RelayRequest{
		Version: Version,
		OpCode:  OpCodeRelayRequest,
		Auth:    c.Auth,
		DstAddr: dst.IP,
		DstPort: uint16(dst.Port),
		SrcAddr: src.IP,
		SrcPort: uint16(src.Port),
	})

	if err != nil {
		conn.Close()
		return
	}

	return
}
