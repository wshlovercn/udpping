package main

import (
	"flag"
	"fmt"
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
	port           = flag.Int("port", 9999, "UDP port to listen on")
	reportInterval = flag.Duration("report", 10*time.Second, "Statistics report interval")
)

func main() {
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}
	defer conn.Close()

	log.Printf("UDP Ping Server listening on %s", addr)

	statistics := stats.NewStatistics()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		ticker := time.NewTicker(*reportInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println(statistics.Report())
			case <-sigChan:
				log.Println("\nShutting down...")
				log.Println("Final Statistics:")
				log.Println(statistics.Report())
				os.Exit(0)
			}
		}
	}()

	buffer := make([]byte, 1024)
	clientAddrs := make(map[string]*net.UDPAddr)
	lastFeedbackTime := time.Now()

	for {
		n, clientAddr, err := conn.ReadFrom(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		packet, err := protocol.Unmarshal(buffer[:n])
		if err != nil {
			log.Printf("Error unmarshaling packet: %v", err)
			continue
		}

		if packet.Type == protocol.PacketTypeProbe {
			statistics.IncrementReceived(packet.Sequence)

			udpAddr, ok := clientAddr.(*net.UDPAddr)
			if ok {
				clientAddrs[clientAddr.String()] = udpAddr
			}

			if time.Since(lastFeedbackTime) >= *reportInterval {
				for _, addr := range clientAddrs {
					sent, received, lossRate := statistics.GetStats()
					feedback := protocol.NewFeedbackPacket(received)
					
					_, err := conn.WriteTo(feedback.Marshal(), addr)
					if err != nil {
						log.Printf("Error sending feedback to %s: %v", addr, err)
					} else {
						log.Printf("Sent feedback to %s: sent=%d, received=%d, loss=%.2f%%", 
							addr, sent, received, lossRate)
					}
				}
				lastFeedbackTime = time.Now()
			}
		}
	}
}
