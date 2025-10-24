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
	port             = flag.Int("port", 9999, "UDP port to listen on")
	reportInterval   = flag.Duration("report", 10*time.Second, "Statistics report interval")
	feedbackInterval = flag.Duration("feedback", 1*time.Minute, "Per-client feedback interval")
	inactiveTimeout  = flag.Duration("timeout", 5*time.Minute, "Client inactive timeout")
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
	log.Printf("Per-client feedback interval: %v", *feedbackInterval)
	log.Printf("Client inactive timeout: %v", *inactiveTimeout)

	clientManager := stats.NewClientManager(*feedbackInterval, *inactiveTimeout)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		reportTicker := time.NewTicker(*reportInterval)
		defer reportTicker.Stop()

		cleanupTicker := time.NewTicker(1 * time.Minute)
		defer cleanupTicker.Stop()

		for {
			select {
			case <-reportTicker.C:
				log.Println(clientManager.GetReport())
			case <-cleanupTicker.C:
				if removed := clientManager.CleanupInactiveClients(); removed > 0 {
					log.Printf("Cleaned up %d inactive client(s)", removed)
				}
			case <-sigChan:
				log.Println("\nShutting down...")
				log.Println("Final Statistics:")
				log.Println(clientManager.GetReport())
				os.Exit(0)
			}
		}
	}()

	buffer := make([]byte, 1024)

	go func() {
		feedbackTicker := time.NewTicker(10 * time.Second)
		defer feedbackTicker.Stop()

		for range feedbackTicker.C {
			for _, client := range clientManager.GetAllClients() {
				if client.ShouldSendFeedback(*feedbackInterval) {
					sent, received, lossRate := client.Statistics.GetStats()
					feedback := protocol.NewFeedbackPacket(received)

					udpAddr, err := net.ResolveUDPAddr("udp", client.Address)
					if err != nil {
						log.Printf("Error resolving address %s: %v", client.Address, err)
						continue
					}

					_, err = conn.WriteTo(feedback.Marshal(), udpAddr)
					if err != nil {
						log.Printf("Error sending feedback to %s: %v", client.Address, err)
					} else {
						log.Printf("Sent feedback to %s: sent=%d, received=%d, loss=%.2f%%",
							client.Address, sent, received, lossRate)
						client.UpdateFeedbackTime()
					}
				}
			}
		}
	}()

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
			clientAddrStr := clientAddr.String()
			client := clientManager.GetOrCreateClient(clientAddrStr)
			client.Statistics.IncrementReceived(packet.Sequence)
			client.UpdateLastPacketTime()
		}
	}
}
