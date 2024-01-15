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
			Monitor(fmt.Sprintf("Error; Keypad; FATAL Error reading keys: " + name), true, true)
			time.Sleep(time.Minute)
		} else {
			if len(message) > 1 || message[0] == '*' || 
				message[0] == '+' || message[0] == '-' || message[0] == '/' {
					// Monitor(fmt.Sprintf("Keypad; Keyboard message sent: %s", message), true, true)
				for _, out := range outputs {
					(*channels)[out] <- message
				}
				
			} else {
				Monitor(fmt.Sprintf("Error; Keypad; bad key command must start with * + - or / got %s", message), true, true)
			}
		}

	}
}
