Autohelm

version 0  untested and in development

Boat Helm driver controller

Boat Helm driver controller runs under TMUX so that as a background task it can
process inputs from a keypad to set desired course, control parameters etc.  

Listening on UDP ports the controller gets streams of data defining the compass heading and helm
position.

The controller writes to RPI hardware ports to control the power and direction of the
helm motor using a motor contoller board and also to a buzzer via a buffer chip.

must run under sudo or else crashes due to permission errors when activating PWM hardware

Cammands:

./autohelm version

sudo ./autohelm run

Key pad control when running (where \n is return - note: no extra charaters allowed for protection against accidental use.  On a standard bluetooth keypad Num Lock in effect acts as on off as at least one number must be in any valid command) 

*999\n   - Shutdown RPI

*1\n     - Autohlem motor not running, set heading and set helm position updated with boat heading and helm position
*7\n     - Autohelm is now active using current heading and rudder position
            *1\n pause 1s then *7\n to reset
*7x\n    - Autohelm is now active using x value 0 to 359.9 as desired compass heading
*0\n     - Print out Motor data        

+x\n  - adjust course by x degrees to starboard can be any valid decimal  350.0 +x = 10.2 if x = 20.2
-x\n  - adjust course by x degrees to port x can be any valid decimal  10.2 -x = 350.0 if x = -20.2

PID dynamic ks and overall gain are hard coded parameters so that for each parameter a setting of 100 should be a good starting point.  These scaling factors are adjustable via the config and keyboard ie 50 reduces the paramter by 50% 1000 makes it 10 times more.  Kp is not adjustable via the key pad since the overall gain can be controlled as well as kd and ki. Kp can be alterred in config so if the hard coded setting are 10x too low or you would like finer control without using decimals you could base everything on 1000 instead of 100; simply define 1000 for kp, ki, kd in config and set overall gain as required.
 
Setting via keypad: (\0 = compass, \1 = helm then \ = gain, . = kd, * =ki) i.e.
\0\x\n - set compass gain - normalised to 100 eg \0\100 <return>
\0*x\n - set compass differential kd - normalised to 100 eg \0*100 <return>
\0.x\n - set compass integral ki - normalised to 100 eg \0.100 <return>

\1\x\n - set helm gain - normalised to 100 eg \0\100 <return>
\1*x\n - set helm differential kd - normalised to 100 eg \0*100 <return>
\1.x\n - set helm integral ki - normalised to 100 eg \0.100 <return>

Note: to set new course use *1\n manually ensure boat is stable on new course then use *7\n to engage autohelm

Config file defines UDP listening Ports for 10hz NMEA compass bearing and for helm position sent as %v\n ie %-1000 to %1000
