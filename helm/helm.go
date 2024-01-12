package helm

import (
	"fmt"
	"autohelm/pid"
	"strconv"

)


func helmProcess(name string, config map[string][]string, channels *map[string](chan string)) {
	// the helm position is normalised to +- 1000
	// an error offset 
	
	pid := pid.MakePid(1, 0.02, 0.01, 0.001, 95)

	pid.Scale_gain = 100
	pid.Scale_kd = 100
	pid.Scale_ki = 100
	pid.Scale_kp = 100

	input := config["input"][0]
	if p, e := strconv.ParseFloat(config["p_factor"][0], 32); e == nil {
		pid.Scale_kp = p
	}
    if i_f, e := strconv.ParseFloat(config["i_factor"][0], 32); e == nil {
		pid.Scale_ki = i_f
	}
    if d_f, e := strconv.ParseFloat(config["d_factor"][0], 32); e == nil {
		pid.Scale_kd = d_f
	}
    if gain_factor, e := strconv.ParseFloat(config["gain_factor"][0], 32); e == nil{
		pid.Scale_gain = gain_factor
	}

	
	go helm_controller(name, input, channels, pid)
	
}

// Helm receives the helm position from the ESP32 remote rudder angle sensor via  a UDP API and powers
// the Motor to move the rudder to the desrired rudder angle.
// A PID is used to calculate the desired power level.
// When under motor the paddle wheel effect means that some rudder is required to go ahead straight
// and when under sail to windard a much larger rudder input is also required as the boat heals and
// tries to round up into the wind.
// When rudder is applied there is a continuous force on the rudder which the helm motor
// must apply even when the rudder angle is constant.
// The PID p factor increases the power in proportion to the error in the rudder position which means
// rudder angle will tend to be greater than desired angle in order to counter act the continuous rudder
// force.
// The PID integration i factor adjusts the power the longer the rudder angle is off and brings the
// helm back to the desired position.
// The PID delta d factor tends to dampen changes in power to improve dynamic smothness and overshoot
// 
//
func helm_controller(name string,  input string, channels *map[string](chan string), pid *pid.Pid) {
	
	for {	
		str := <-(*channels)[input]
		fmt.Printf("Received helm command %s\n", str)
		if str[0:1] == "%" {
			rudder, err := strconv.ParseFloat(str[1:], 64)
			if err == nil {
				Motor.Rudder = rudder
				error := Motor.Set_rudder - Motor.Rudder
				power := pid.Compute(error, Motor.Rudder)
				fmt.Printf("Motor power: %f \n", power)
				Motor.Helm(power)
			} else {
				fmt.Printf("error in rudder parsing: %s \n", err)
			}
		}
		
	}
		
}
