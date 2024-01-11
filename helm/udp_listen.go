/*
Copyright © 2024 Martin Marsh martin@marshtrio.com

*/

package helm

import (
	"fmt"
	"net"

)

func udpListenerProcess(name string, config map[string][]string, channels *map[string](chan string)) {
	// listens on a port and writes to output channels
	fmt.Println("started udp Listener " + name)
	server_port := config["port"][0]
	fmt.Printf("UDP listener on port %s and out to channels:", server_port)
	for _, out := range config["outputs"] {
		fmt.Printf(" %s", out)
	}
	fmt.Println()
	if len(config["outputs"]) > 0 {
		go udpListener(name, server_port, config["outputs"], channels)
	}
	
}

func udpListener(name string, server_port string, outputs []string, channels *map[string](chan string)) {
	const maxBufferSize = 1024
	pc, err := net.ListenPacket("udp", "0.0.0.0:"+server_port)
	if err != nil {
		fmt.Println("udp listen error - aborted")
		return
	}
	defer pc.Close()

	buffer := make([]byte, maxBufferSize)
	
	for{
		n, _, err := pc.ReadFrom(buffer)
		if err != nil {
			fmt.Printf("packet error")
	
		} else {
			for _, out := range outputs {
				(*channels)[out] <- string(buffer[:n])
			}
		}

	}


}
