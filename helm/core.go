/*
Copyright © 2022 Martin Marsh martin@marshtrio.com

*/

package helm

import (
	"autohelm/io"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

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
			adjust_heading(str, +1)

		case "-":
			adjust_heading(str, -1)
		case "/":
			adjust_gain(str[1:])
		}
	}
}

func adjust_gain(str string){
	end_byte := len(str)
	if end_byte > 3 {
		p, e := strconv.ParseFloat(str[2:end_byte-1], 64)
		if str[end_byte-1] == '\n' && e == nil {
			switch str[2:3]{
			case "1*":
				Motor.Helm_gain = p
			case "2*":
				Motor.Compass_gain = p

			}
		}
	}	
}

func adjust_heading(str string, dir float64){
	end_byte := len(str)
	p, e := strconv.ParseFloat(str[1:end_byte-1], 64)
	if str[end_byte-1] == '\n' && e == nil {
		p = p * dir
		Motor.Set_heading = compass_direction(Motor.Set_heading + p)
		fmt.Printf("adjust by %.1f New Heading: %.1f", p, Motor.Set_heading)
	} else {
		fmt.Printf("Bad value for compass setting: %s value: '%s'", e, str)
	}

}

func compass_direction(compass float64) float64 {
	for compass < 0 || compass >= 360.0 {
		if compass >= 360.0 {
			compass -= 360.0
		} else if compass < 0 {
			compass += 360.0
		}
	}
    return compass
}


func process_commands(str string) {
	fmt.Printf("Process command: %s", str)

	switch str {
	case "999\n":
		{
			fmt.Println("Shutting Down!")
			io.Beep("2l")
			Exit()
		}
	case "1\n":
		{
			fmt.Println("Motor Disabled")
			Motor.Enabled = false

			io.Beep("1s")
		}
	case "7\n":
		{
			fmt.Println("Motor Enabled")
			Motor.Enabled = true
			io.Beep("1s")
		}
	}
}

func Exit() {
	out, err := exec.Command("shutdown", "-h", "now").Output()

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
