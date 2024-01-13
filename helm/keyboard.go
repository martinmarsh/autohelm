package helm

import (
	"bufio"
	"fmt"
	"os"
	//"strings"
	"time"
)

func keyBoardProcess(name string, config map[string][]string, channels *map[string](chan string)) {
	reader := bufio.NewReader(os.Stdin)
	if len(config["outputs"]) > 0 {
		go keyOutputs(name, reader, config["outputs"], channels)
	}
}

func keyOutputs(name string, reader *bufio.Reader, outputs []string, channels *map[string](chan string)) {
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("FATAL Error reading keys: " + name)
			time.Sleep(time.Minute)
		} else {
			if len(message) > 1 || message[0] == '*' || 
				message[0] == '+' || message[0] == '-' || message[0] == '/' {
					
				fmt.Printf("Keyboard message sent: %s\n", message)
				for _, out := range outputs {
					(*channels)[out] <- message
				}
				
			} else {
				fmt.Printf("bad key command must start with * + - or / got %s\n", message)
			}
		}

	}
}
