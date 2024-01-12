package helm

import (
	"fmt"
	"autohelm/pid"
	"strconv"
	"time"

)


func helmProcess(name string, config map[string][]string, channels *map[string](chan string)) {
	
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

	
	go helm_controller(name, input, channels, pid)
	
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
func helm_controller(name string,  input string, channels *map[string](chan string), pid *pid.Pid) {
	const max_power = 3000    // 3ms cycle time  300us min
	const max_loops = 85	 // 85 x 3  255 ms read channel period
	const max_power_slow = 20000    // 20ms cycle time  3ms min
	const max_loops_slow = 14	 // 14 x 20  280 ms read channel period
	//t1 := time.Duration(0)
	//t2 := time.Duration(max_power_slow) * time.Microsecond
	//mp := max_power_slow
	//ml := max_loops_slow
	
	
	Motor.Port(8)
	time.Sleep(time.Second)
	Motor.Port(4)
	time.Sleep(time.Second)
	Motor.Port(0)
	time.Sleep(time.Second)
	Motor.Starboard(1)
	time.Sleep(time.Second)
	Motor.Starboard(4)
	time.Sleep(4*time.Second)
	Motor.Port(1)

	

	for {	
		//ask helm to compute another value
		// (*channels)["to_helm"] <-"compute"
		str := <-(*channels)[input]
		fmt.Printf("Received helm command %s\n", str)
		

		/*select {
		case motor := <- Motor_channel:
			fmt.Println("motor control")
			switch motor.control {
			
			p1 := 0
			mp = max_power_slow
			ml = max_loops_slow

			if motor.power > 0.2 && motor.power < 0.8 {	
				mp = max_power
				ml = max_loops
			}
		    if motor.power > 0.95 {
				p1 = mp
			} else if motor.power < 0.02{
				p1 = 0
			} else {
				p1 = int(float64(mp) * motor.power)
			}
			t1 = time.Duration(p1) * time.Microsecond
			t2 = time.Duration(mp - p1) * time.Microsecond
			//fmt.Printf("%d %d\n", t1, t2)	
		default:
			// continue
		}

		if t1 == 0 {
			power_pin.Low()
			time.Sleep(250 * time.Millisecond)
		} else if t2 == 0 {
			power_pin.High()
			time.Sleep(250 * time.Millisecond)
		} else {
			for i := 0; i < ml; i++ {
				if t1 > 0 {
					power_pin.High()
					time.Sleep(t1)
				}
				if t2 > 0 {
					power_pin.Low()
					time.Sleep(t2)
				}
			}
		}
		*/

	}
		
}
