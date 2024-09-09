package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

var (
	delay  int
	size   int
	count  int
	target string
)

var rootCmd = &cobra.Command{
	Use:   "pingo",
	Short: "A ping tool written in Go",
	Run:   runPing,
}

func init() {
	rootCmd.Flags().IntVarP(&delay, "delay", "d", 1000, "Delay between each packet in milliseconds")
	rootCmd.Flags().IntVarP(&size, "size", "s", 56, "Size of the packet in bytes")
	rootCmd.Flags().IntVarP(&count, "count", "c", 4, "Number of packets to send")
	rootCmd.Flags().StringVarP(&target, "target", "t", "", "Target host to ping")
	rootCmd.MarkFlagRequired("target")
}

func runPing(cmd *cobra.Command, args []string) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer conn.Close()

	for i := 0; i < count; i++ {
		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{
				ID:   os.Getpid() & 0xffff,
				Seq:  i,
				Data: make([]byte, size),
			},
		}
		msgBytes, err := msg.Marshal(nil)
		if err != nil {
			fmt.Println("Error marshalling message:", err)
			continue
		}

		start := time.Now()
		_, err = conn.WriteTo(msgBytes, &net.IPAddr{IP: net.ParseIP(target)})
		if err != nil {
			fmt.Println("Error sending packet:", err)
			continue
		}

		reply := make([]byte, 1500)
		err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			fmt.Println("Error setting read deadline:", err)
			continue
		}

		n, peer, err := conn.ReadFrom(reply)
		if err != nil {
			fmt.Println("Error reading reply:", err)
			continue
		}

		duration := time.Since(start)
		rm, err := icmp.ParseMessage(1, reply[:n])
		if err != nil {
			fmt.Println("Error parsing reply:", err)
			continue
		}

		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v\n", n, peer, i, duration)
		default:
			fmt.Printf("Got %+v from %v\n", rm, peer)
		}

		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
