/*
Copyright © 2022 Martin Marsh martin@marshtrio.com

*/

package io

import (
	"strconv"
	"time"
	"math"
	"sync"
	"sync/atomic"

	"github.com/stianeikeland/go-rpio/v4"
	
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
	mu       sync.Mutex
	left_pin rpio.Pin
	right_pin rpio.Pin
	power_pin rpio.Pin
    Power uint32
	Duty_Power uint32
	Set_rudder float64
	Rudder float64
	Set_heading float64
	Heading float64
	Enabled atomic.Bool
	OverRange atomic.Bool
	Compass_gain float64
	Helm_gain float64
	Compass_ki float64
	Compass_kd float64
	Helm_ki float64
	Helm_kd float64
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
		Power: 0,
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
	c.Power = 0
	c.Duty_Power = 0
	c.Set_rudder = 0
	c.Set_heading =0 
	c.Rudder = 0
	c.Heading = 0
	c.Enabled.Store(false)
	c.OverRange.Store(false)
	c.left_pin.Output()
	c.right_pin.Output()
	c.power_pin.Pwm()
	c.power_pin.DutyCycle(c.Power, PWM_CYCLE_LEN)
	c.power_pin.Freq(PWM_FREQUENCY)
	c.left_pin.Low()
	c.right_pin.Low()
	rpio.StartPwm()   
}

func (c *HelmCtrl) Helm(power float64){
	if power < 0 {
		c.Port(uint32(math.Abs(power)))
	} else {
		c.Starboard(uint32(power))
	}
}


func (c *HelmCtrl) On(power uint32){
	c.mu.Lock()
    defer c.mu.Unlock()
	c.onPreLocked(power)
} 


func (c *HelmCtrl) Off(){
	c.mu.Lock()
    defer c.mu.Unlock()
	if c.Power != 0 {
		c.Power = 0
		c.Duty_Power = 0
		c.power_pin.DutyCycle(c.Duty_Power, PWM_CYCLE_LEN) 
	}
}

func (c *HelmCtrl) Port(power uint32){
	c.mu.Lock()
	defer c.mu.Unlock()
	c.right_pin.Low()
	c.left_pin.High() 
	c.onPreLocked(power)  
}

func (c *HelmCtrl) Starboard(power uint32){
	c.mu.Lock()
	defer c.mu.Unlock()
	c.left_pin.Low()
	c.right_pin.High()
	c.onPreLocked(power)  
}

func (c *HelmCtrl) onPreLocked(power uint32){
	var p uint32 = (power * PWM_CYCLE_LEN)/100
	c.Power = power
	if p < PWM_MIN_DUTY {
		p = 0
	}
	if  p > PWM_MAX_DUTY {
		p = PWM_CYCLE_LEN
	}
	c.Duty_Power = p
	c.power_pin.DutyCycle(c.Duty_Power, PWM_CYCLE_LEN) 
} 

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

				beep_pin.High() 
				<-t.C
				
				beep_pin.Low() 
				<-t.C
			}
		}
	}

}
