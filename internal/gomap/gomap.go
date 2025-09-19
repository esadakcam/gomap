package gomap

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type App struct {
	IpRange      *net.IPNet
	HostInfoList []HostInfo
}

type HostInfo struct {
	Address      net.IP
	Reachable    bool
	OpenTcpPorts []uint16
}

func (host *HostInfo) Scan() {
	var wg sync.WaitGroup
	host.Reachable = true // For now
	wg.Add(1)
	host.OpenTcpPorts = scanTcpForIp(host.Address.String(), &wg)
	wg.Wait()
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
	var wg sync.WaitGroup
	var mtx sync.Mutex

	for ip := network.IP.Mask(network.Mask); network.Contains(ip); inc(ip) {
		wg.Add(1)
		ipCopy := make(net.IP, len(ip))
		copy(ipCopy, ip)
		go func(hostInfoList *[]HostInfo, ip net.IP, wg *sync.WaitGroup, mtx *sync.Mutex) {
			defer wg.Done()
			hostInfo := HostInfo{Address: ip}
			hostInfo.Scan()
			mtx.Lock()
			*hostInfoList = append(*hostInfoList, hostInfo)
			mtx.Unlock()

		}(&hostInfoList, ipCopy, &wg, &mtx)
	}
	wg.Wait()
	return &App{IpRange: network, HostInfoList: hostInfoList}
}

func scanTcpForIp(ip string, wg *sync.WaitGroup) (result []uint16) {
	defer wg.Done()
	result = make([]uint16, 0)
	timeout := 200 * time.Millisecond
	for _, port := range WellKnownPorts() {
		address := ip + ":" + fmt.Sprintf("%d", port)
		fmt.Println(ip, port)
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
