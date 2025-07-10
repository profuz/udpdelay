package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/enriquebris/goconcurrentqueue"
)

type element struct {
	Buffer    []byte
	Timestamp time.Time
}

var queue = goconcurrentqueue.NewFIFO()

const maxDatagramSize = 8192

func getAddress(address *string) *net.UDPAddr {
	if *address == "" {
		return nil
	}

	result, err := net.ResolveUDPAddr("udp", *address)
	if err != nil {
		log.Fatal(err)
	}
	return result
}

func write(output, outputInterface *string, delay float64) {
	connection, err := net.DialUDP("udp", getAddress(outputInterface), getAddress(output))
	if err != nil {
		log.Fatal(err)
	}

	duration, err := time.ParseDuration(fmt.Sprintf("%fs", delay))
	if err != nil {
		log.Fatal(err)
	}

	for {
		item, err := queue.DequeueOrWaitForNextElement()
		if err != nil {
			log.Println(err)
			continue
		}

		element := item.(element)
		time.Sleep(time.Until(element.Timestamp.Add(duration))) // this is OK to be negative
		connection.Write(element.Buffer)
	}
}

// udpdelay -i 224.0.0.1:1234 [-i_interface 192.168.0.100] -o 224.0.0.3:1235 [-o_interface 192.168.1.200] -delay 20.240
func main() {
	input := flag.String("i", "", "Input address and port (required)")
	inputInterface := flag.String("i_interface", "", "Interface for input stream")
	output := flag.String("o", "", "Output address and port (required)")
	outputInterface := flag.String("o_interface", "", "Interface for output stream")
	delay := flag.Float64("delay", 20, "Delay in seconds")
	flag.Parse()

	if *input == "" || *output == "" {
		fmt.Println("Error: Missing required parameters")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	go write(output, outputInterface, *delay)

	remoteAddress := getAddress(input)

	var server *net.UDPConn
	var err error
	if remoteAddress.IP.IsMulticast() {
		var localInterace *net.Interface
		if inputInterface != nil {
			localInterace, err = net.InterfaceByName(*inputInterface)
			if err != nil {
				localAddress := net.ParseIP(*inputInterface)

				interfaces, err := net.Interfaces()
				if err != nil {
					log.Fatal(err)
				}

				for _, iface := range interfaces {
					addresses, err := iface.Addrs()
					if err != nil {
						log.Fatal(err)
					}

					for _, address := range addresses {
						ip, _, err := net.ParseCIDR(address.String())
						if err != nil {
							log.Fatal(err)
						}

						if ip.String() == localAddress.String() {
							localInterace = &iface
							break
						}
					}
				}
			}
		}

		server, err = net.ListenMulticastUDP("udp", localInterace, remoteAddress)
	} else {
		server, err = net.ListenUDP("udp", remoteAddress)
	}
	if err != nil {
		log.Fatal(err)
	}

	server.SetReadBuffer(maxDatagramSize)
	buffer := make([]byte, maxDatagramSize)
	for {
		n, _, err := server.ReadFromUDP(buffer)
		if err != nil {
			log.Println(err)
			continue
		}

		queueBuffer := make([]byte, n)
		n = copy(queueBuffer, buffer)
		if n != len(queueBuffer) {
			log.Println("Didn't copy correctly")
			continue
		}

		queue.Enqueue(element{Buffer: queueBuffer, Timestamp: time.Now()})
	}
}
