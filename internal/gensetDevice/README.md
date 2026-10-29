# Generator Controller

## States

The generator controller is a Moore machine, a finite-state machine where the control outputs (Ignition, Starter, Fan, Pump, Load)
only depend on the current state and the inputs are only used to activate transitions.
The diagnostic outputs (*Time in state*, *I/O check*, *Output check*) are computed from the inputs and the time as well.

There are the following states:

 0) *Error*: An error occurred (e.g. I/O unavailable, fire or temperature alarm, cranking or output check failed), engine shut off immediately. Restart only after manual reset.
 1) *Reset*: Reset button pressed
 2) *Off*: Generator not running, sensors might be offline.
 3) *Ready*: Generator not running, all sensors are online.
 4) *Priming*: Fuel pump running, engine not yet cranking.
 5) *Cranking*: Starter trying to start the engine
 6) *Stabilizing*: Engine started, waiting for the output to stabilize.
 7) *Warm-Up*: Engine running but not producing power. Waiting for temperature to rise.
 8) *Producing*: Engine producing power
 9) *Engine cool down*: Engine not producing power, idling to cool down
 10) *Enclosure cool down*: Engine off, but enclosure still ventilated

## Inputs
* *Command switch*: On/Off switch (X). Together with the *Arm switch*, it forms the master switch: the generator runs only while both are on.
* *Reset switch*: Emergency off and failure reset switch (X)

* *Time in current state*: Timer reset by state changes.
* I/O controller:
  * *I/O available*: True, when all devices providing outputs or inputs relevant to the temperature and fire checks are available.
  * *Arm switch*: Local user enabled auto-start of generator. (X)
  * *Fire detected*: True, when fire or smoke is detected. (X)
  * *Engine block temperature*: Temperature measured on the aluminium block of the generator. (X)
  * *Aux temperature 0* (`AuxTemp0`): Generic auxiliary temperature, e.g. the outside air blown into the enclosure. (X)
  * *Aux temperature 1* (`AuxTemp1`): Generic auxiliary temperature, e.g. the air blown out of the enclosure. (X)
* Output measurement:
  * *Output measurement available*: True, when all devices providing measurements for the output checks are available.
  * *Phase Voltage U[1-3]*: Voltage output of the generator.
  * *Phase Power P[1-3]*: Power measurement of the generator.
  * *Frequency*: Frequency measurement of the generator.

## Outputs
 * *Ignition*: Turns on the ignition of the generator or the command input of the generator's own engine management.
 * *Starter*: Turns on the starter motor.
 * *Fan*: Turns on the enclosure fan.
 * *Pump*: Turns on the fuel pump.
 * *Load*: Connects the load to the generator output.
 * *Time in state*, *I/O check*, *Output check*: Diagnostic values, not control outputs.

 ## Configuration Variables
 * *Priming*
   * *timeout*: after this time, cranking starts
 * *Cranking*
   * *timeout*: if the output check is not positive after this time, the controller goes to the error state
 * *Stabilizing*
   * *timeout*: after this time, the output is considered stable and warm up starts
 * *Warm up*
   * *timeout*: after this time, the warm up is completed
   * *min time*: warm up is never completed before this time
   * *engine block temperature*: when the engine is warmer than this value, warm up is completed
 * *Engine Cool down*
   * *timeout*: after this time, the engine cool down is completed
   * *min time*: cool down is never completed before this time
   * *engine block temperature*: when the engine is colder than this value, cool down is completed
 * *Enclosure cool down*
   * *timeout*: after this time, the enclosure cool down is completed
   * *min time*: cool down is never completed before this time
   * *engine block temperature*: when the engine block temperature is colder than this value, cool down is completed.
 * *I/O checks*: min / max range of the engine, aux 0 and aux 1 temperatures
 * *Output checks*:
   * UMin / UMax : tolerable voltage range
   * FMin / FMax : tolerable frequency range
   * PMax: tolerable power per phase
   * PTotMax: tolerable total power (the single phase P1 for single phase generators, otherwise P1 + P2 + P3)
