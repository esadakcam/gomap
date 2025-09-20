package icmp

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	PING_ATTEMTPS = 3
)

func Ping(address string) bool {
	raddr, err := net.ResolveIPAddr("ip4", address)
	if err != nil {
		fmt.Println("failed to resolve target address: ", err)
		return false
	}

	conn, err := net.DialIP("ip4:icmp", nil, raddr)
	if err != nil {
		fmt.Println("connection failed: ", err)
		return false
	}
	defer conn.Close()
	data := []byte("hello from client")
	for i := 0; i < PING_ATTEMTPS; i++ {
		echoReq := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   os.Getpid() & 0xffff,
				Seq:  i,
				Data: data[:],
			},
		}
		msgBytes, err := echoReq.Marshal(nil)

		if err != nil {
			fmt.Println("failed to marshal ICMP message: ", err)
			continue
		}

		if err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
			fmt.Println("failed to set read deadline: ", err)
			continue
		}

		timeStart := time.Now()
		if _, err := conn.Write(msgBytes); err != nil {
			fmt.Println("failed to send ICMP message: ", err)
			continue
		}

		resp := make([]byte, 512)
		n, peer, err := conn.ReadFrom(resp)
		timeEnd := time.Now()
		if err != nil {
			fmt.Println("failed to read ICMP response: ", err)
			continue
		}

		parsedMsg, err := icmp.ParseMessage(1, resp[:n])
		if err != nil {
			fmt.Println("failed to parse ICMP message: ", err)
			continue
		}

		echoType := parsedMsg.Type
		body := parsedMsg.Body.(*icmp.Echo)
		proto := parsedMsg.Type.Protocol()

		switch parsedMsg.Type {
		case ipv4.ICMPTypeEchoReply:
			elapsed := timeEnd.Sub(timeStart)
			fmt.Printf("%d bytes from %s: pid=%d, icmp_type=%v, icmp_seq=%d, data=%s, time:%dμs\n", body.Len(proto), peer, body.ID, echoType, body.Seq, string(body.Data), elapsed.Microseconds())
			return true
		default:
			fmt.Printf("received unexpected message from %s: pid=%d, icmp_type=%v, icmp_seq=%d, data=%s\n", peer, body.ID, echoType, body.Seq, string(body.Data))
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false

}
