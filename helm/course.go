package helm

import (
	"fmt"
	"autohelm/pid"
	"strconv"
	"strings"

)

func relative_direction(diff float64) float64 {
    if diff < -180.0 {
        diff += 360.0
	} else if diff > 180.0 {
        diff -= 360.0
	}
    return diff
}

func checksum(s string) string {
	check_sum := 0
	nmea_data := []byte(s)
	for i := 1; i < len(s); i++ {
		check_sum ^= (int)(nmea_data[i])
	}
	return fmt.Sprintf("%02X", check_sum)
}

func courseProcess(name string, config map[string][]string, channels *map[string](chan string)) {
	
	pid := pid.MakePid(1, 0.1, 0.3, .01, 10000)

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

	Motor.Compass_gain = pid.Scale_gain
	Motor.Compass_ki = pid.Scale_ki
	Motor.Compass_kd = pid.Scale_kd
	go course(name, input, channels, pid)
	
}

// Helm gets compass bearing at 10hz from the input channel
// A PID is used to calculate the actulating signal AS which in this case is simply a value passed to
// the helm control to move the rudder to the desired position.
// 
// The PID calculates the AS signal for every Compass input at a constant 10Hz.  It is left to the helm
// control to move the rudder hardware and detect the rudder poistion in real time and compensate for rudder
// forces. Near to the rudder set point the helm reduces power to the motor limiting overhoot.  So this task
// does not need to consider these factors and simply calculates a correcting rudder position proportional
// to the course compass error.
// Since a rudder angle is related to the error the course can be off by a factor relating to
// the gain.  The PID integration factor can progressively reduce this error but if set too large will cause
// helm instability.
//
// A differential error based on sp-pv is used to dampen occilations
//
// The set helm is assumed to be +/- 10,000 centred on zero.
// Calibrated Offsets and scale are corrected in helm functions
//
func course(name string,  input string, channels *map[string](chan string), pid *pid.Pid) {

	for {
		str := <-(*channels)[input]
		var err error
		err = nil
		Monitor(fmt.Sprintf("Received; course command: %s", str), false, true)
		if len(str)> 9 && str[0:6] == "$HCHDM"{
			end_byte := len(str)
			if str[end_byte-3] == '*' {
				check_code := checksum(str[:end_byte-3])
				end_byte -= 2
				if check_code != str[end_byte:] {
					err_mess := fmt.Sprintf("error: %s != %s", check_code, str[end_byte:])
					err = fmt.Errorf("check sum error: %s", err_mess)
				}
				end_byte--
			}
		
			if err == nil{
				parts := strings.Split(str[1:end_byte], ",")
				heading, _ := strconv.ParseFloat(parts[1], 64)
				pid.Scale_gain = Motor.Compass_gain
				pid.Scale_kd = Motor.Compass_kd
				pid.Scale_ki = Motor.Compass_ki
				Motor.Heading = heading
				if Motor.Enabled {
					sp_pv := relative_direction(Motor.Set_heading - Motor.Heading)
					Motor.Set_rudder = pid.Compute(sp_pv, sp_pv)
					Monitor(fmt.Sprintf("Course; helm: on, heading: %.1f, set-heading: %.1f, sp-pv: %.1f, Rudder required: %.0f\n",
						 Motor.Heading, Motor.Set_heading, sp_pv, Motor.Set_rudder), false, true)
				} else {
					Motor.Set_heading = Motor.Heading
					Monitor(fmt.Sprintf("Course; helm: off, heading: %.1f, set-heading: %.1f, set-rudder: %.0f\n",
						 Motor.Heading, Motor.Set_heading, Motor.Set_rudder), false, true)
				}
			} else {
				Monitor(fmt.Sprintf("Error; Compass NMEA 0183 error: %s", err), true, true)
			}
		}
				
	}
}
