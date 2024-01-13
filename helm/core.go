/*
Copyright © 2022 Martin Marsh martin@marshtrio.com

*/

package helm

import (
	"fmt"
	"autohelm/io"
	"time"
	"os"
	"os/exec"

	"github.com/stianeikeland/go-rpio/v4"
)

type ConfigData struct {
	Index    map[string]([]string)
	TypeList map[string]([]string)
	Values   map[string]map[string]([]string)
}

var Motor *io.HelmCtrl

func Execute(config *ConfigData) {
	// wait for everything to connect on boot up
	controller_name := ""
	time.Sleep(5 * time.Second)
	
	channels := make(map[string](chan string))
	fmt.Println("Autohelm execute")
	
	for name, param := range config.Index {
		for _, value := range param {
			if value == "outputs" {
				for _, chanName := range config.Values[name][value] {
					if _, ok := channels[chanName]; !ok {
						channels[chanName] = make(chan string, 30)
					}
				}
			}
			if value == "input" {
				for _, chanName := range config.Values[name][value] {
					if _, ok := channels[chanName]; !ok {
						channels[chanName] = make(chan string, 30)
					}
				}
			}
		}
	}

	if err := rpio.Open(); err != nil {
		fmt.Println("RPIO - could not be openned")
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()

	Motor = io.Init()

	for processType, names := range config.TypeList {
		fmt.Println(processType, names)
		for _, name := range names {
			switch processType {
			case "controller":
				controller_name = name
			case "udp_client":
				udpClientProcess(name, config.Values[name], &channels)
			case "udp_listen":
				udpListenerProcess(name, config.Values[name], &channels)
			case "keyboard":
				keyBoardProcess(name, config.Values[name], &channels)
			case "helm":
				helmProcess(name, config.Values[name], &channels)
			case "course":
				courseProcess(name, config.Values[name], &channels)	
			}
		}
	}

	
	io.Beep("1s")

	//Now run controller process

	input := config.Values[controller_name]["input"][0]
	fmt.Printf("Controller in core.go wiating on channel: %s\n", input)
	for {
		str := <-(channels)[input]
		fmt.Printf("Got key request: %s", str)
		switch str[0:1] {
		case "*":
			process_commands(str[1:])
		case "+":
			{

			}
			
		case "-":
			{

			}
		case "/":
			{

			}
		}
	}
}
	
	

func process_commands(str string){
	fmt.Printf("Process command: %s", str)

	switch str{
	case "999\n":
		{
			fmt.Println("Shutting Down!")
			io.Beep("2l")
			Exit()
		}
	case "1\n": {
			fmt.Println("Motor Disabled")
			Motor.Enabled = false
			io.Beep("1s")
		}
	case "7\n":{
			fmt.Println("Motor Enabled")
	    	Motor.Enabled = true
			io.Beep("1s")
		}
	}
}


func Exit() {
	out, err := exec.Command("shutdown","-h","now").Output()

    // if there is an error with our execution
    // handle it here
    if err != nil {
        fmt.Printf("%s", err)
    }
    // as the out variable defined above is of type []byte we need to convert
    // this to a string or else we will see garbage printed out in our console
    // this is how we convert it to a string
    fmt.Println("Command Successfully Executed")
    output := string(out[:])
    fmt.Println(output)
}

