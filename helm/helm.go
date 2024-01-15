package helm

import (
	"fmt"
	"autohelm/pid"
	"strconv"

)

type calib struct {
	max float64
	min float64
	centre float64
	scale float64
}


func helmProcess(name string, config map[string][]string, channels *map[string](chan string)) {
	// the helm position is normalised to +- 10000
	// an error offset 
	helm_calib := calib{
		max: 1000,
		min: -1000,
		centre: 0,
		scale: 0.1,
	}
	
	if p, e := strconv.ParseFloat(config["max_helm"][0], 32); e == nil {
		helm_calib.max = p
	}
	
	if p, e := strconv.ParseFloat(config["min_helm"][0], 32); e == nil {
		helm_calib.min = p
	}

	if p, e := strconv.ParseFloat(config["centre_helm"][0], 32); e == nil {
		helm_calib.centre = p
	}

	helm_calib.scale = (helm_calib.max - helm_calib.min)/20000
	
	pid := pid.MakePid(1, 0.01, 0.01, 0.01, 100)

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

	Motor.Helm_gain = pid.Scale_gain
	Motor.Helm_ki = pid.Scale_ki
	Motor.Helm_kd = pid.Scale_kd

	go helm_controller(name, input, channels, pid, &helm_calib)
	
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
func helm_controller(name string,  input string, channels *map[string](chan string), pid *pid.Pid, helm_calib *calib) {
	for {	
		str := <-(*channels)[input]
		Monitor(fmt.Sprintf("Received; helm command: %s", str), false, true)
		if str[0:1] == "%" {
			pid.Scale_gain = Motor.Helm_gain
			pid.Scale_ki = Motor.Helm_ki
			pid.Scale_kd = Motor.Helm_kd
			rudder, err := strconv.ParseFloat(str[1:], 64)
			if err == nil {
				if rudder > helm_calib.max {
					rudder = helm_calib.max
					Motor.In_range = false
				} else if rudder < helm_calib.min {
					rudder = helm_calib.min
					Motor.In_range = false
				} else {
					Motor.In_range = true
				}
				if Motor.In_range {
					Motor.Rudder = rudder / helm_calib.scale - helm_calib.centre
					if Motor.Enabled{
						sp_pv := Motor.Set_rudder - Motor.Rudder
						power := pid.Compute(sp_pv, -Motor.Rudder) 
						Motor.Helm(power)
					} else {
						Motor.Set_rudder = Motor.Rudder
						Motor.Off()
					}
				} else {
					Motor.Off()
				}
			} else {
				Monitor(fmt.Sprintf("Error; error in rudder parsing: %s \n", err), true,true)
			}
		}
		
	}
		
}
