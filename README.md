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

Key pad control when running (where \n is return - note no extra charaters allowed for protection against accidental use) 

*999\n   - Shutdown RPI

*1\n     - Autohlem motor not running, set heading and set helm position updated with boat heading and helm position
*7\n     - Autohelm is active

+x\n  - adjust course by x degrees to starbox can be any valid decimal  350.0 +x = 10.2 if x = 20.2
-x\n  - adjust course by x degrees to port x can be any valid decimal  10.2 -x = 350.0 if x = -20.2


