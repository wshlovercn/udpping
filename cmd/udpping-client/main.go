package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wshlovercn/udpping/pkg/protocol"
	"github.com/wshlovercn/udpping/pkg/stats"
)

var (
	serverAddr = flag.String("server", "localhost:9999", "UDP ping server address (host:port)")
	rate       = flag.Duration("rate", 1*time.Second, "Probe packet send rate")
	timeout    = flag.Duration("timeout", 5*time.Second, "Read timeout for feedback packets")
)

func main() {
	flag.Parse()

	udpAddr, err := net.ResolveUDPAddr("udp", *serverAddr)
	if err != nil {
		log.Fatalf("Failed to resolve server address %s: %v", *serverAddr, err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", *serverAddr, err)
	}
	defer conn.Close()

	log.Printf("UDP Ping Client started, sending probes to %s at rate %v", *serverAddr, *rate)

	statistics := stats.NewStatistics()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go receiveLoop(conn, statistics)

	ticker := time.NewTicker(*rate)
	defer ticker.Stop()

	var sequence uint64 = 0

	for {
		select {
		case <-ticker.C:
			sequence++
			packet := protocol.NewProbePacket(sequence)
			
			_, err := conn.Write(packet.Marshal())
			if err != nil {
				log.Printf("Error sending probe: %v", err)
			} else {
				statistics.IncrementSent()
				log.Printf("Sent probe #%d", sequence)
			}

		case <-sigChan:
			log.Println("\nShutting down...")
			log.Println("Final Statistics:")
			log.Println(statistics.Report())
			os.Exit(0)
		}
	}
}

func receiveLoop(conn *net.UDPConn, statistics *stats.Statistics) {
	buffer := make([]byte, 1024)

	for {
		conn.SetReadDeadline(time.Now().Add(*timeout))
		
		n, err := conn.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Printf("Error reading feedback: %v", err)
			continue
		}

		packet, err := protocol.Unmarshal(buffer[:n])
		if err != nil {
			log.Printf("Error unmarshaling feedback packet: %v", err)
			continue
		}

		if packet.Type == protocol.PacketTypeFeedback {
			sent, received, lossRate := statistics.GetStats()
			log.Printf("Received feedback from server: seq=%d | %s | Server stats: sent=%d, received=%d, loss=%.2f%%", 
				packet.Sequence, statistics.Report(), sent, received, lossRate)
		}
	}
}
