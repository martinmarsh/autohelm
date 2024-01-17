/*
Copyright © 2022 Martin Marsh martin@marshtrio.com

*/

package io

import (
	"strconv"
	"time"
	"math"
	"sync"
	"fmt"
	"autohelm/pid"

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
    power uint32
	dutyPower uint32
	setRudder float64
	rudder float64
	setHeading float64
	heading float64
	enabled bool
	overRange bool
	compassGain float64
	helmGain float64
	compassKi float64
	compassKd float64
	helmKi float64
	helmKd float64
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
	c.dutyPower = 0
	c.setRudder = 0
	c.setHeading =0 
	c.rudder = 0
	c.heading = 0
	c.enabled = false
	c.overRange = false
	c.left_pin.Output()
	c.right_pin.Output()
	c.power_pin.Pwm()
	c.power_pin.DutyCycle(c.power, PWM_CYCLE_LEN)
	c.power_pin.Freq(PWM_FREQUENCY)
	c.left_pin.Low()
	c.right_pin.Low()
	rpio.StartPwm()   
}

func (c *HelmCtrl) Enable(set bool){
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = set
}

func (c *HelmCtrl) IsEnabled() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.enabled
}


func (c *HelmCtrl) SetActualRudder(rudder float64){
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rudder = rudder
}


func (c *HelmCtrl) RudderInRange(rawRudder float64, rawMax float64, rawMin float64) bool {
	ret := false
	c.mu.Lock()
	defer c.mu.Unlock()
	if rawRudder > rawMax {
		c.overRange = true
		c.offPreLocked()
	} else if rawRudder < rawMin {
		c.overRange = true
		c.offPreLocked()
	} else {
		c.overRange = false
		ret = true	
	}
	
	return ret
}
 
func (c *HelmCtrl) IncrDesiredHeading(inrc float64) float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.setHeading = compass_direction(c.setHeading + inrc)
	return c.setHeading
}

func (c *HelmCtrl) SetDesiredHeading(heading float64) float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.setHeading = compass_direction(heading)
	return c.setHeading
}

func (c *HelmCtrl) ProcessHeading(heading float64) (float64, bool){
	c.mu.Lock()
	defer c.mu.Unlock()
	var diff float64
	c.heading = heading
	if c.enabled {
		diff = c.setHeading - c.heading
		// Monitor(fmt.Sprintf("Course; helm: on, heading: %.1f, set-heading: %.1f, sp-pv: %.1f, Rudder required: %.0f\n",
		//	 Motor.Heading, Motor.Set_heading, sp_pv, Motor.Set_rudder), false, true)
	} else {
		c.setRudder = c.heading
		//Monitor(fmt.Sprintf("Course; helm: off, heading: %.1f, set-heading: %.1f, set-rudder: %.0f\n",
		//	 Motor.Heading, Motor.Set_heading, Motor.Set_rudder), false, true)
	}

	return diff, c.enabled
}

func (c *HelmCtrl) GetDesiredRudder() float64{
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.setRudder
}

func (c *HelmCtrl) SetDesiredRudder(rudder float64){
	c.mu.Lock()
	defer c.mu.Unlock()
	c.setRudder = rudder
}


func (c *HelmCtrl) SetPidByKeyCode(keyCode string, value float64) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	changed := ""
	switch keyCode{
		case "0/": {
			c.compassGain = value 
			changed = "compassGain"
		}
		case "0.":{
			c.compassKd = value
			changed = "compassKd"
		}
		case "0*":{
			c.compassKi = value
			changed = "compassKi"
		}
		case "1/":{
			c.helmGain = value
			changed = "helmGain"
		}
		case "1*":{
			c.helmKi = value
			changed = "helmKi"
		}
		case "1.":{
			c.helmKd = value
			changed = "helmKd"
		}
	}
	return changed
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
	c.offPreLocked()

}

func (c *HelmCtrl) offPreLocked(){
	if c.power != 0 {
		c.power = 0
		c.dutyPower = 0
		c.power_pin.DutyCycle(c.dutyPower, PWM_CYCLE_LEN) 
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

func (c *HelmCtrl) GetMonitorString(strType uint32) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var rep string
	if strType == 1 {
		rep = fmt.Sprintf("Monitor; power: %d, set_rudder: %.0f, rudder: %.0f, set_heading: %.1f, heading: %.1f, Enabled: %t, OverRange: %t, compass_gain: %.1f, helm_gain: %.1f", 
					c.power, c.setRudder, c.rudder, c.setHeading, c.heading, c.enabled, c.overRange,
					c.compassGain, c.helmGain)
	} else {
		rep = fmt.Sprintf("Monitor; duty_power: %d, rudder: %.0f, heading: %.1f, compass_gain: %.1f, helm_gain: %.1f, compass_ki: %.1f, compass_kd: %.1f, helm_ki: %.1f, helm_kd: %.1f", 
		c.dutyPower, c.rudder, c.heading, c.compassGain, c.helmGain, c.compassKi,
		c.compassKd, c.helmKi, c.helmKd)
	}
	return rep
}

func (c *HelmCtrl) onPreLocked(power uint32){
	var p uint32 = (power * PWM_CYCLE_LEN)/100
	c.power = power
	if p < PWM_MIN_DUTY {
		p = 0
	}
	if  p > PWM_MAX_DUTY {
		p = PWM_CYCLE_LEN
	}
	c.dutyPower = p
	c.power_pin.DutyCycle(c.dutyPower, PWM_CYCLE_LEN) 
} 

func (c *HelmCtrl) SetPidCompass(pid  *pid.Pid){
	c.mu.Lock()
	defer c.mu.Unlock()
	pid.Scale_gain = c.compassGain
	pid.Scale_kd = c.compassKd
	pid.Scale_ki = c.compassKi
}

func (c *HelmCtrl) SetCompassPid(pid  *pid.Pid){
	c.mu.Lock()
	defer c.mu.Unlock()
	c.compassGain = pid.Scale_gain
	c.compassKd = pid.Scale_kd
	c.compassKi = pid.Scale_ki
}

func (c *HelmCtrl) SetPidHelm(pid  *pid.Pid){
	c.mu.Lock()
	defer c.mu.Unlock()
	pid.Scale_gain = c.helmGain
	pid.Scale_kd = c.helmKd
	pid.Scale_ki = c.helmKi
}

func (c *HelmCtrl) SetHelmPid(pid  *pid.Pid){
	c.mu.Lock()
	defer c.mu.Unlock()
	c.helmGain = pid.Scale_gain
	c.helmKd = pid.Scale_kd
	c.helmKi = pid.Scale_ki
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
	
