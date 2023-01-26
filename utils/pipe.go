package utils

import (
	"io"
	"net"
	"time"
)

const PipeBufferSize = 32 * 1024

func PipePairWithTimeout(conn net.Conn, stream io.ReadWriteCloser, timeout time.Duration) error {
	errChan := make(chan error, 2)
	// TCP to stream
	go func() {
		buf := make([]byte, PipeBufferSize)
		for {
			if timeout != 0 {
				_ = conn.SetDeadline(time.Now().Add(timeout))
			}
			rn, err := conn.Read(buf)
			if rn > 0 {
				_, err := stream.Write(buf[:rn])
				if err != nil {
					errChan <- err
					return
				}
			}
			if err != nil {
				errChan <- err
				return
			}
		}
	}()
	// Stream to TCP
	go func() {
		buf := make([]byte, PipeBufferSize)
		for {
			rn, err := stream.Read(buf)
			if rn > 0 {
				_, err := conn.Write(buf[:rn])
				if err != nil {
					errChan <- err
					return
				}
				if timeout != 0 {
					_ = conn.SetDeadline(time.Now().Add(timeout))
				}
			}
			if err != nil {
				errChan <- err
				return
			}
		}
	}()
	return <-errChan
}
