package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type ClientConfig struct {
	Inbound struct {
		Mode   ClientMode `json:"mode"`
		Listen string     `json:"listen"`
	} `json:"inbound"`

	Outbound struct {
		Server       string     `json:"server"`
		Auth         string     `json:"auth"`
		LocalAddr    *LocalAddr `json:"local_addr"`
		FirewallMark *uint32    `json:"fwmark"`
	} `json:"outbound"`

	Timeout  *Timeout `json:"timeout"`
	LogLevel LogLevel `json:"log_level"`
}

type ServerConfig struct {
	Listen   string               `json:"listen"`
	ACL      []AccessControlEntry `json:"acl"`
	Timeout  *Timeout             `json:"timeout"`
	LogLevel LogLevel             `json:"log_level"`
}

type ClientMode string

const (
	ClientModeDefault  ClientMode = ""
	ClientModeRedirect ClientMode = "redirect"
)

func (c *ClientMode) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(string(*c))), nil
}

func (c *ClientMode) UnmarshalJSON(bs []byte) (err error) {
	s, err := strconv.Unquote(string(bs))
	if err != nil {
		err = fmt.Errorf("failed to parse client mode: %w", err)
		return
	}
	cs := ClientMode(s)
	switch cs {
	case ClientModeDefault:
		break
	case ClientModeRedirect:
		break
	default:
		err = errors.New("invalid client mode")
		return
	}
	*c = cs
	return
}

type LocalAddr struct {
	IP     net.IP
	Device string
}

func (l *LocalAddr) String() string {
	sb := strings.Builder{}
	if l.IP != nil {
		sb.WriteString(l.IP.String())
	}
	if l.Device != "" {
		sb.WriteString("%")
		sb.WriteString(l.Device)
	}
	return sb.String()
}

func (l *LocalAddr) Parse(s string) (err error) {
	tokens := strings.SplitN(s, "%", 2)
	addrString := strings.TrimSpace(tokens[0])
	if addrString != "" {
		l.IP = net.ParseIP(addrString)
		if l.IP == nil {
			err = errors.New("invalid IP address")
			return
		}
	}
	if len(tokens) > 1 {
		l.Device = strings.TrimSpace(tokens[1])
	}
	return
}

func (l *LocalAddr) MarshalJSON() (bs []byte, err error) {
	bs = []byte(strconv.Quote(l.String()))
	return
}

func (l *LocalAddr) UnmarshalJSON(bs []byte) (err error) {
	s, err := strconv.Unquote(string(bs))
	if err != nil {
		err = fmt.Errorf("failed to parse local address: %w", err)
		return
	}
	err = l.Parse(s)
	return
}

type Timeout time.Duration

const defaultTimeout = Timeout(300 * time.Second)

func (t *Timeout) MarshalJSON() (bs []byte, err error) {
	bs = []byte(strconv.Quote(time.Duration(*t).String()))
	return
}

func (t *Timeout) UnmarshalJSON(bs []byte) (err error) {
	s, uerr := strconv.Unquote(string(bs))
	if uerr != nil {
		// it also might be a number
		i, nerr := strconv.Atoi(string(bs))
		if nerr != nil {
			err = fmt.Errorf("failed to parse timeout: %w, %w", uerr, nerr)
			return
		}
		*t = Timeout(time.Duration(i) * time.Second)
		return
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return
	}
	*t = Timeout(d)
	return
}

func (t *Timeout) Duration() time.Duration {
	if t == nil {
		return time.Duration(defaultTimeout)
	}
	return time.Duration(*t)
}

type Prefix struct {
	*net.IPNet
}

func (p *Prefix) MarshalJSON() (bs []byte, err error) {
	s := p.String()
	bs = []byte(strconv.Quote(s))
	return
}

func (p *Prefix) UnmarshalJSON(bs []byte) (err error) {
	s, err := strconv.Unquote(string(bs))
	if err != nil {
		err = fmt.Errorf("failed to parse cidr prefix: %w", err)
		return
	}
	_, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return
	}
	p.IPNet = ipnet
	return
}

type AccessControlEntry struct {
	Auth          string   `json:"auth"`
	AllowedSrcIPs []Prefix `json:"allowed_src_ips"`
}

type LogLevel int

const (
	LogLevelInfo  LogLevel = 3
	LogLevelError LogLevel = 5
)

func (l *LogLevel) String() string {
	switch *l {
	case LogLevelInfo:
		return "info"
	case LogLevelError:
		return "error"
	default:
		return "unknown"
	}
}

func (l *LogLevel) Parse(s string) (err error) {
	switch s {
	case "":
		fallthrough
	case "info":
		*l = LogLevelInfo
	case "error":
		*l = LogLevelError
	default:
		err = fmt.Errorf("invalid log level: %s", s)
	}
	return
}

func (l *LogLevel) MarshalJSON() (bs []byte, err error) {
	bs = []byte(strconv.Quote(l.String()))
	return
}

func (l *LogLevel) UnmarshalJSON(bs []byte) (err error) {
	s, err := strconv.Unquote(string(bs))
	if err != nil {
		err = fmt.Errorf("failed to parse log level: %w", err)
		return
	}
	err = l.Parse(s)
	return
}
