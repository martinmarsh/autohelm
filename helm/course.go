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
	
	pid := pid.MakePid(1, 0.1, 0.3, 0.00001, 0.95)

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

	
	go course(name, input, channels, pid)
	
}

// Helm takes collects instructions and a compass bearing at 10hz from the input channel
// A PID is used to calculate the actuating signal which is sent to the motor controller.
// The helm motor runs with a speed defined by actuating signal (AS) value either left or right
// rotation; turning the wheel continuously.  The rudder position cannot be sensed nor is it 
// driven to a position as determined by the AS value. This would be required for the error signal
// to be based on the course error.  Then at zero error the rudder will be near centre but in our
// system it would be just be comming to a halt at the maximum deflection.  In short there is an 
// integration effect  making steering unstable at any proportional gain setting.

// To overcome this problem the effective rudder position can be sensed by the rate of turn of the boat.
// Then Zero rate of turn is straight ahead. Hence the rate of course change is used as the feedback signal
// to calculate the error for the PID.  The set point can then be defined as the desired rate of turn to
// reach the desired course.  This has 2 benfits: 
// 1) The turn rate is controlled and is not too excessive for large corrections or tacking.
// 2) The rudder effectiveness which varies greatly with boat speed is automatically compensated.
// 
// The PID calculates the AS signal for every Compass input at a constant 10Hz.  The motor calls
// back through a channel when it is ready to receive the next AS instruction.
//
func course(name string,  input string, channels *map[string](chan string), pid *pid.Pid) {

	var heading float64 = 0.0

	for {
		str := <-(*channels)[input]
		var err error
		fmt.Printf("Received course command %s\n", str)
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
				heading, _ = strconv.ParseFloat(parts[1], 64)
				
				// the turn rate is averaged across 3s, 1.5s and 0.5s periods
				fmt.Printf("heading= %f\n", heading)
			}
		}
				
	}
}
