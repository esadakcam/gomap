package gomap

import (
	"fmt"
	"math"
	"math/bits"
	"net"
	"time"
)

type App struct {
	IpRange      *net.IPNet
	HostInfoList []HostInfo
}

type HostInfo struct {
	Address      string //IPV4 Address
	Reachable    bool
	OpenTcpPorts []uint16
}

func (host *HostInfo) Scan() {
	host.Reachable = true // For now
	host.OpenTcpPorts = scanTcpForIp(host.Address)
}

func WellKnownPorts() []uint16 {
	return []uint16{22, 80, 443, 53}
}

func Scan(cidr string) *App {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}
	hostInfoList := make([]HostInfo, 0)
	ch := make(chan HostInfo)
	numIps := countNumberOfIps(network.Mask)

	for ip := network.IP.Mask(network.Mask); network.Contains(ip); inc(ip) {
		go func(ip string, ch chan HostInfo) {
			hostInfo := HostInfo{Address: ip}
			hostInfo.Scan()
			ch <- hostInfo
		}(ip.String(), ch)
	}

	for i := 0; i < numIps; i++ {
		hostInfo := <-ch
		hostInfoList = append(hostInfoList, hostInfo)
	}

	return &App{IpRange: network, HostInfoList: hostInfoList}
}

func scanTcpForIp(ip string) (result []uint16) {
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

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func countNumberOfIps(mask net.IPMask) int {
	numOfOnes := 0
	for _, byt := range mask {
		numOfOnes += bits.OnesCount8(uint8(byt))
	}
	return int(math.Pow(2, float64(32-numOfOnes)))
}
