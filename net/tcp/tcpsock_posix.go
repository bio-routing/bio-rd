package tcp

import (
	"fmt"
	"net"
	"runtime"

	"golang.org/x/sys/unix"
)

func dialTCP(afi uint16, laddr, raddr *net.TCPAddr, ttl uint8, md5secret string, noRoute bool) (*Conn, error) {
	fd, err := unix.Socket(int(afi), unix.SOCK_STREAM, unix.IPPROTO_TCP)
	if err != nil {
		return nil, fmt.Errorf("socket() failed: %w", err)
	}

	c := &Conn{
		fd:    fd,
		laddr: laddr,
		raddr: raddr,
	}

	err = c.SetNoDelay()
	if err != nil {
		return nil, fmt.Errorf("unable to set TCP_NODELAY: %w", err)
	}

	if ttl != 0 {
		err = c.SetTTL(ttl)
		if err != nil {
			return nil, fmt.Errorf("unable to set IP_TTL: %w", err)
		}
	}

	if noRoute {
		err = c.SetDontRoute()
		if err != nil {
			return nil, fmt.Errorf("unable to set SO_DONTROUTE: %w", err)
		}
	}

	if laddr != nil && laddr.IP != nil {
		var bindSA unix.Sockaddr
		if laddr.IP.To4() != nil {
			la := ipv4AddrToArray(laddr.IP)
			bindSA = &unix.SockaddrInet4{
				Port: laddr.Port,
				Addr: la,
			}
		} else {
			la := ipv6AddrToArray(laddr.IP)
			bindSA = &unix.SockaddrInet6{
				Port: laddr.Port,
				Addr: la,
			}
		}

		err := unix.Bind(fd, bindSA)
		if err != nil {
			return nil, fmt.Errorf("bind() failed: %w", err)
		}
	}

	if md5secret != "" {
		if runtime.GOOS != "linux" {
			return nil, fmt.Errorf("TCP MD5 authentication is not supported on %s", runtime.GOOS)
		}
		err := setTCPMD5Option(fd, raddr.IP, md5secret)
		if err != nil {
			return nil, fmt.Errorf("unable to set TCP MD5 secret: %w", err)
		}
	}

	var connectSA unix.Sockaddr
	if raddr.IP.To4() != nil {
		connectSA = &unix.SockaddrInet4{
			Port: raddr.Port,
			Addr: ipv4AddrToArray(raddr.IP),
		}
	} else {
		connectSA = &unix.SockaddrInet6{
			Port: raddr.Port,
			Addr: ipv6AddrToArray(raddr.IP),
		}
	}

	err = unix.Connect(fd, connectSA)
	if err != nil {
		return nil, fmt.Errorf("connect() failed: %w", err)
	}

	return &Conn{
		fd:    fd,
		laddr: laddr,
		raddr: raddr,
	}, nil
}

func ipv6AddrToArray(x net.IP) [16]byte {
	return [16]byte{
		x[0], x[1], x[2], x[3], x[4], x[5], x[6], x[7],
		x[8], x[9], x[10], x[11], x[12], x[13], x[14], x[15],
	}
}

func ipv4AddrToArray(x net.IP) [4]byte {
	return [4]byte{
		x[0], x[1], x[2], x[3],
	}
}
