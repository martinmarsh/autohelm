/*
Copyright © 2022 Martin Marsh martin@marshtrio.com

*/

package io

import (
	// "math"
	"strconv"
	"time"
	"math"

	"github.com/stianeikeland/go-rpio/v4"
	//"periph.io/x/periph/conn/gpio"
	//"periph.io/x/periph/host"
	//"periph.io/x/periph/host/rpi"
	//"periph.io/x/conn/v3/physic"
)

// Note the PWM for chip is 20khz which implies 50% pulse of 25 miro secs
// the PWM_MIN_CYCLE prevents higher rates
// power is given as 0 to 100 

const (
	BEEP_OUT_PIN = 25 // Pin connected to Beeper output
	RIGHT_MOTOR_PIN = 23 // Pin to motor controller right turn direction (green led)
	LEFT_MOTOR_PIN = 24 // Pin to motor controller left turn direction (red led - port)
	PWM_MOTOR_PIN = 18 	// Pin to motor controller enable / PWM control
	PWM_CYCLE_LEN = 15   // Number of steps
	PWM_FREQUENCY = 152000 // at 50% duty = 10.1 kHz - min pulse = 6.579 micro secs
	PWM_MIN_DUTY = 3  //  means shortest on/off time is 3*6.58 = 20 miro secs
	PWM_MAX_DUTY = PWM_CYCLE_LEN - PWM_MIN_DUTY // ie duty can be 3 to 12  ie 20 to 80%
)

var Beep_channel chan string
var beep_pin = rpio.Pin(BEEP_OUT_PIN)

type HelmCtrl struct {
	left_pin rpio.Pin
	right_pin rpio.Pin
	power_pin rpio.Pin
    power uint32
	Set_rudder float64
	Rudder float64
	Set_heading uint16
	Heading uint16
}

func Beep(style string){
	Beep_channel <- style
}

func Init() *HelmCtrl{

	Beep_channel = make(chan string, 4)

	helm_io := HelmCtrl{
		left_pin: rpio.Pin(LEFT_MOTOR_PIN),
		right_pin: rpio.Pin(RIGHT_MOTOR_PIN),
		power_pin: rpio.Pin(PWM_MOTOR_PIN),
		power: 0,
	}
	
	helm_io.init()

	beep_pin.Output()
	go beeperTask()
	
	return &helm_io
}

func (c *HelmCtrl) init(){
	c.left_pin = rpio.Pin(24)
	c.right_pin = rpio.Pin(23)
	c.power_pin = rpio.Pin(18)
	c.power = 0
	c.Set_rudder = 0
	c.Set_heading =0 
	c.Rudder = 0
	c.Heading = 0
	c.left_pin.Output()
	c.right_pin.Output()
	c.power_pin.Pwm()
	c.power_pin.DutyCycle(c.power, PWM_CYCLE_LEN)
	c.power_pin.Freq(PWM_FREQUENCY)
	c.left_pin.Low()
	c.right_pin.Low()
	rpio.StartPwm()   
}

func (c *HelmCtrl) Port(power uint32){
	c.right_pin.Low()
	c.left_pin.High() 
	c.On(power)  
}

func (c *HelmCtrl) Starboard(power uint32){
	c.left_pin.Low()
	c.right_pin.High()
	c.On(power)  
}

func (c *HelmCtrl) Helm(power float64){
	if power < 0 {
		c.Port(uint32(math.Abs(power)))
	} else {
		c.Starboard(uint32(power))
	}
}


func (c *HelmCtrl) On(power uint32){
	var p uint32 = (power * PWM_CYCLE_LEN)/100
	if p < PWM_MIN_DUTY {
		p = 0
	}
	if  p > PWM_MAX_DUTY {
		p = PWM_CYCLE_LEN
	}
	c.power = p
	c.power_pin.DutyCycle(c.power, PWM_CYCLE_LEN) 
} 

func (c *HelmCtrl) Off(){
	c.power = 0
	c.power_pin.DutyCycle(c.power, PWM_CYCLE_LEN) 
}



/*
func Helm(control byte, power float64){
	// control = 1 if power is a signed value
	// control = 0 to turn off motor
	// control = 'L' or 'R' if power is unsigned
	if control == 1 {
		control = 'R'
		if power < 0 {
			control = 'L'
			power = math.Abs(power)
		}

	}
	message :=  helm_control{
		control: control,
		power: power,   	// <=1 1 = 100%
	}
	fmt.Println(message)
	Motor_channel <- message
}

func helmTask(channels *map[string](chan string)){
	const max_power = 3000    // 3ms cycle time  300us min
	const max_loops = 85	 // 85 x 3  255 ms read channel period
	const max_power_slow = 20000    // 20ms cycle time  3ms min
	const max_loops_slow = 14	 // 14 x 20  280 ms read channel period
	t1 := time.Duration(0)
	t2 := time.Duration(max_power_slow) * time.Microsecond
	mp := max_power_slow
	ml := max_loops_slow
		
	for {	
		//ask helm to compute another value
		// (*channels)["to_helm"] <-"compute"

		select {
		case motor := <- Motor_channel:
			fmt.Println("motor control")
			switch motor.control {
			case 'L':
				right_pin.Low()
				left_pin.High()
				//rpi.P1_16.Out(gpio.Low)
				//rpi.P1_18.Out(gpio.High)
			case 'R':
				left_pin.Low()
				right_pin.High()
				//rpi.P1_18.Out(gpio.Low)
				//rpi.P1_16.Out(gpio.High)
			case 1:
			default:
				left_pin.Low()
				right_pin.Low()
				//rpi.P1_18.Out(gpio.Low)
				//rpi.P1_16.Out(gpio.Low)
			}
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

	}
}
*/

func beeperTask(){
	for{
		beep := <- Beep_channel
		if len(beep) == 2 {
			count, _ := strconv.Atoi(string(beep[0]))
			
			length := 400
			if beep[1] == 's'{
				length = 100
			} else if beep[1] == 'l'{
				length = 800
			}
			for l := 0; l < count; l++  {
				t := time.NewTicker(time.Duration(length) * time.Millisecond)
				//rpi.P1_22.Out(gpio.High)

				beep_pin.High() 
				<-t.C
				//rpi.P1_22.Out(gpio.Low)
				beep_pin.Low() 
				<-t.C
			}
		}
	}

}
