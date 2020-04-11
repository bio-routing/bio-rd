package server

import (
	"net"
	"os"
	"strings"
	"syscall"
)

func extractFileAndFamilyFromTCPListener(l *net.TCPListener) (*os.File, int, error) {
	// Note #1: TCPListener.File() has the unexpected side-effect of putting
	// the original socket into blocking mode. See Note #2.
	fi, err := l.File()
	if err != nil {
		return nil, 0, err
	}

	// Note #2: Call net.FileListener() to put the original socket back into
	// non-blocking mode.
	fl, err := net.FileListener(fi)
	if err != nil {
		fi.Close()
		return nil, 0, err
	}
	fl.Close()

	return fi, getAFIFromAddr(l.Addr().String()), nil
}

func extractFileAndFamilyFromTCPConn(c *net.TCPConn) (*os.File, int, error) {
	// Note #1: TCPListener.File() has the unexpected side-effect of putting
	// the original socket into blocking mode. See Note #2.
	fi, err := c.File()
	if err != nil {
		return nil, 0, err
	}

	// Note #2: Call net.FileListener() to put the original socket back into
	// non-blocking mode.
	fl, err := net.FileListener(fi)
	if err != nil {
		fi.Close()
		return nil, 0, err
	}
	fl.Close()

	return fi, getAFIFromAddr(c.LocalAddr().String()), nil
}

func getAFIFromAddr(addr string) int {
	if strings.Contains(addr, "[") {
		return syscall.AF_INET6
	}

	return syscall.AF_INET
}
