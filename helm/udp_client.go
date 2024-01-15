/*
Copyright © 2024 Martin Marsh martin@marshtrio.com

*/

package helm

import (
	"fmt"
	"net"

)

func udpClientProcess(name string, config map[string][]string, channels *map[string](chan string)) {
	Monitor("Process; started udp " + name, true, true)
	server_addr := config["server_address"][0]
	input_channel := config["input"][0]
	fmt.Println(server_addr, input_channel)
	go udpWriter(name, server_addr, input_channel, channels)

}

func udpWriter(name string, server_addr string, input string, channels *map[string](chan string)) {
	connection := false
	for{
		RemoteAddr, _ := net.ResolveUDPAddr("udp", server_addr)
		conn, err := net.DialUDP("udp", nil, RemoteAddr)
		
		if err != nil {
			Monitor(fmt.Sprintf("Error; Could not open udp server %s error: %s", name, err.Error()), true, true)
			//ensure channel is cleared then retry
			connection = false
			for i :=0 ; i > 100; i++ {
				<-(*channels)[input]
			}
		} else {
			defer conn.Close()
			Monitor(fmt.Sprintf("Udp_client; Established connection to %s", server_addr), true, true)
			// Monitor(fmt.Sprintf("Udp_client; Remote UDP address : %s", conn.RemoteAddr().String()),true, true)
			// Monitor(fmt.Sprintf("Udp_client; Local UDP client address : %s", conn.LocalAddr().String()), true, true)
			connection = true
		}
		if connection {
			for {
				str := <-(*channels)[input]
				_, err := conn.Write([]byte(str))
				if err != nil {
					Monitor("Error; Udp_client; FATAL Error on UDP connection" + name, true, true)
				}
				
			}
		}
	}
}
