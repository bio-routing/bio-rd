package tcp

import (
	"fmt"
	"net"
	"runtime"

	"golang.org/x/sys/unix"
)

func dialTCP(afi uint16, laddr, raddr *net.TCPAddr, ttl uint8, md5secret string, noRoute bool, bindDev string) (*Conn, error) {
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

	if bindDev != "" {
		err = c.SetBindToDev(bindDev)
		if err != nil {
			return nil, fmt.Errorf("unable to set SO_BINDTODEV: %w", err)
		}
	}

	if laddr != nil && laddr.IP != nil {
		err := unix.Bind(fd, netTCPAddrToSockAddr(laddr))
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

	err = unix.Connect(fd, netTCPAddrToSockAddr(raddr))
	if err != nil {
		return nil, fmt.Errorf("connect() failed: %w", err)
	}

	return &Conn{
		fd:    fd,
		laddr: laddr,
		raddr: raddr,
	}, nil
}

func netTCPAddrToSockAddr(tcpAddr *net.TCPAddr) unix.Sockaddr {
	ip := tcpAddr.IP
	ip4 := ip.To4()
	if ip4 != nil {
		return &unix.SockaddrInet4{
			Port: tcpAddr.Port,
			Addr: [4]byte{
				ip4[0], ip4[1], ip4[2], ip4[3],
			},
		}
	}

	return &unix.SockaddrInet6{
		Port: tcpAddr.Port,
		Addr: [16]byte{
			ip[0], ip[1], ip[2], ip[3], ip[4], ip[5], ip[6], ip[7],
			ip[8], ip[9], ip[10], ip[11], ip[12], ip[13], ip[14], ip[15],
		},
	}
}
