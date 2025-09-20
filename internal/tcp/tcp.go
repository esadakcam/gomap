package tcp

import (
	"fmt"
	"net"
	"time"
)

func WellKnownPorts() []uint16 {
	return []uint16{22, 80, 443, 53}
}

func Scan(ip string) (result []uint16) {
	result = make([]uint16, 0)
	timeout := 200 * time.Millisecond
	for _, port := range WellKnownPorts() {
		address := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
		conn, err := net.DialTimeout("tcp", address, timeout)
		if err == nil {
			result = append(result, port)
			conn.Close()
		}
	}
	return
}
