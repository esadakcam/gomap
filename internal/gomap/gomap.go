package gomap

import (
	"net"
	"sync"

	"github.com/esadakcam/gomap/internal/icmp"
	"github.com/esadakcam/gomap/internal/tcp"
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

	host.Reachable = icmp.Ping(host.Address)
	if host.Reachable {
		host.OpenTcpPorts = tcp.Scan(host.Address)
		return
	}
	host.OpenTcpPorts = make([]uint16, 0)
}

func Scan(cidr string) *App {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}
	numOfWorkers := 500

	jobs := make(chan string)
	results := make(chan HostInfo)

	var wg sync.WaitGroup

	for i := 0; i < numOfWorkers; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); scanWorker(jobs, results) }()
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	go func() {
		for ip := network.IP.Mask(network.Mask); network.Contains(ip); inc(ip) {
			jobs <- ip.String()
		}
		close(jobs)
	}()

	hostInfoList := make([]HostInfo, 0)
	for hostInfo := range results {
		hostInfoList = append(hostInfoList, hostInfo)
	}

	return &App{IpRange: network, HostInfoList: hostInfoList}
}

func scanWorker(jobs chan string, results chan HostInfo) {
	for ip := range jobs {
		hostInfo := HostInfo{Address: ip}
		hostInfo.Scan()
		results <- hostInfo
	}
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
