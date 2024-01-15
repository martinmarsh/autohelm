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
var Monitor_channel chan string
var Udp_monitor_active bool

func Monitor(str string, print bool, udp bool){
	if udp && Udp_monitor_active {
		Monitor_channel <- str
	}
	if print {
		fmt.Println(str)
	}
}

func Execute(config *ConfigData) {
	// wait for everything to connect on boot up
	controller_name := ""
	Udp_monitor_active = false
	
	time.Sleep(5 * time.Second)
	
	channels := make(map[string](chan string))
	fmt.Println("Parsing Config. in Execute function")

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

	for processType, names := range config.TypeList {
		for _, name := range names {
			if processType == "udp_monitor" {
				Monitor_channel = channels[config.Values[name]["input"][0]]
				Udp_monitor_active = true
				udpClientProcess(name, config.Values[name], &channels)
			}
		}
	}
	time.Sleep(2 * time.Second)
	Monitor(fmt.Sprintf("Autohelm; Monitoring: %t, Openning RPIO", Udp_monitor_active), true, true)

	if err := rpio.Open(); err != nil {
		Monitor(fmt.Sprintf("Error; RPIO - could not be openned err: %s", err.Error()), true, true)
		os.Exit(1)
	}
	defer rpio.Close()
	Monitor("RPIO; Openned", true, true)

	Motor = io.Init()
	Monitor("Autohelm; Motor has initialised - running up configured processes", true, true)

	for processType, names := range config.TypeList {
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

	Monitor(fmt.Sprintf("Controller; in core.go waiting on channel: %s", input), true, true)
	for {
		str := <-(channels)[input]
		Monitor(fmt.Sprintf("Controller; Got key request: %s", str), true, true)
		switch str[0:1] {
		case "*":
			process_commands(str[1:])
		case "+":
			adjust_heading(str, +1)

		case "-":
			adjust_heading(str, -1)
		case ".":
			adjust_ks(str[1:])
		}
	}
}

func adjust_ks(str string){
	Monitor(fmt.Sprintf("Controller; adjust ks command %s", str), true, true)
	end_byte := len(str)
	if end_byte > 3 {
		p, e := strconv.ParseFloat(str[2:end_byte-1], 64)

		if str[end_byte-1] == '\n' && e == nil {
			switch str[2:3]{
			case "0/":
				Motor.Compass_gain = p
				Monitor(fmt.Sprintf("Controller; compass gain: %.2f", Motor.Compass_gain), true, true)
			case "0.":
				Motor.Compass_kd = p
				Monitor(fmt.Sprintf("Controller; compass kd: %.2f", Motor.Compass_kd), true, true)
			case "0*":
				Motor.Compass_ki = p
				Monitor(fmt.Sprintf("Controller; compass ki: %.2f", Motor.Compass_ki), true, true)
			case "1/":
				Motor.Helm_gain = p
				Monitor(fmt.Sprintf("Controller; helm gain: %.2f", Motor.Helm_gain), true, true)
			case "1*":
				Motor.Helm_ki = p
				Monitor(fmt.Sprintf("Controller; helm ki: %.2f", Motor.Helm_ki), true, true)
			case "1.":
				Motor.Helm_kd = p
				Monitor(fmt.Sprintf("Controller; helm kd: %.2f",Motor.Helm_kd), true, true)
			}
		}else{
			Monitor(fmt.Sprintf("Error; Controller; adjusting PID values error: %s", e.Error()), true, true)
		}
	}else {
		Monitor(fmt.Sprintf("Error; Controller; adjusting PID wrong length command must be >3 got: %d", end_byte), true, true)
	}	
}

func adjust_heading(str string, dir float64){
	end_byte := len(str)
	p, e := strconv.ParseFloat(str[1:end_byte-1], 64)
	if str[end_byte-1] == '\n' && e == nil {
		p = p * dir
		Motor.Set_heading = compass_direction(Motor.Set_heading + p)
		Monitor(fmt.Sprintf("Controller; adjusted by %.1f New Heading: %.1f", p, Motor.Set_heading), true, true)
	} else {
		Monitor(fmt.Sprintf("Error; In Controller Bad value for compass setting: %s value: '%s'", e, str), true, true)
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
	Monitor(fmt.Sprintf("Controller; Process command: %s", str), true, true)
	end_byte := len(str)
	if str[0:1] == "7" && end_byte > 2{
		p, e := strconv.ParseFloat(str[1:end_byte-1], 64)
		if str[end_byte-1] == '\n' && e == nil && p<= 360 && p>= 0 {
			Motor.Set_heading = compass_direction(p)
			Monitor(fmt.Sprintf("Controller; motor: on, - set_course: %.1f\n", Motor.Set_heading), true, true)
			Motor.Enabled = true
			io.Beep("1s")
		}

	} else {

		switch str {
		case "999\n":
			{
				Monitor("Controller; Shutting Down!", true, true)
				io.Beep("2l")
				Exit()
			}
		case "1\n":
			{
				Monitor("Controller; motor: off", true, true)
				Motor.Enabled = false
				io.Beep("1s")
			}
		case "7\n":
			{
				Monitor("Controller; motor: on", true, true)
				Motor.Enabled = true
				io.Beep("1s")
			}
		case "0\n":
			{
				rep := fmt.Sprintf("Monitor; power: %d, set_rudder: %.0f, rudder: %.0f, set_heading: %.1f, heading: %.1f, enabled: %t, in_range: %t, compass_gain: %.1f, helm_gain: %.1f", 
					Motor.Power, Motor.Set_rudder, Motor.Rudder, Motor.Set_heading, Motor.Heading, Motor.Enabled, Motor.In_range,
					Motor.Compass_gain, Motor.Helm_gain)
				Monitor(rep, true, true)
			}
		case ".\n":
			{
				rep := fmt.Sprintf("Monitor; duty_power: %d, rudder: %.0f, heading: %.1f, compass_gain: %.1f, helm_gain: %.1f, compass_ki: %.1f, compass_kd: %.1f, helm_ki: %.1f, helm_kd: %.1f", 
				Motor.Duty_Power, Motor.Rudder, Motor.Heading, Motor.Compass_gain, Motor.Helm_gain, Motor.Compass_ki,
				Motor.Compass_kd, Motor.Helm_ki, Motor.Helm_kd)
				Monitor(rep, true, true)
			}
		}
	}
}

func Exit() {
	out, err := exec.Command("shutdown", "-h", "now").Output()

	// if there is an error with our execution
	// handle it here
	if err != nil {
		fmt.Printf("%s", err)
		Monitor(fmt.Sprintf("Error; shutown -h command has error: %s", err.Error()), true, true)
	}
	// as the out variable defined above is of type []byte we need to convert
	// this to a string or else we will see garbage printed out in our console
	// this is how we convert it to a string
	output := string(out[:])
	Monitor(output, true, true)
}
