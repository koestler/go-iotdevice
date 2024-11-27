# Generator Controller

## States

The generator controller is a Moore machine, a finite-state machine where the outputs only depend on the current state and the inputs are only used
to activate transitions.

There are the following states:

-2) *Failed*: A failure occured, engine shut of immediately. Re-start only after manual reset.
-1) *Reset*: Reset button pressed
 0) *Off*: Generator not running, sensors might be offline.
 1) *Ready*: Generator not running, all sensors are online.
 2) *Cranking*: Starter tryingn to start the engine
 3) *Warm-Up*: Engine running but not producing power. Waiting for temperature to rise.
 4) *Producing*: Engine producing power
 5) *Engine cool down*: Engine not producing power, idling to cool down
 6) *Enclosure cool down*: Engine off, but enclusre still ventilated

## Inputs
* *Master switch*: On/Off switch (X)
* *Reset switch*: Emegrgency off and failure reset switch (X)

* *Time in current state*: Timer reset by state changes.
* I/O controller:
  * *I/O availabe*: True, when all devices providing outputs or inputs relevant to the temperature and fire checks are available.
  * *Arm switch*: Local user enabled auto-start of generator. (X)
  * *Fire detected*: True, when fire or smoke is detected. (X)
  * *Engine block temperature*: Temperature messured on the aluminium block of the generate. (X)
  * *Air intake temperature*: Temperature measured of the outside air blown into the enclosure. (X)
  * *Air exhaust temperature*: Temeperatured mesaured of the air blown out of the enclosure. (X)
* Output measurement:
  * *Output measurement available*: True, when all devices providing measurements for the output checks are available.
  * *Phase Voltage U[1-3]*: Voltage output of the generator.
  * *Phase Power L[1-3]*: Power measurement of the generator.
  * *Frequency*: Frequency measurement of the generator.

## Outputs
 * *Ignition*: Turns on the ignition of the generator or the command input of the generators own engine management.
 * *Starter*: Turns on the starter motor.
 * *Fan*: Turns on the enclosure fan.
 * *Load*: Connects the load to the generator output.

 ## Configuration Variables
 * *Warm up*
   * *timeout*: after this time, the warm up is completed
   * *engine block temperature*: when the engine is warmer than this value, warm up is completed 
 * *Engine Cool down*
   * *timeout*: after this time, the engine cool down is completed
   * *engine block temperature*: when the engine is colder than this value, cool down is completed
 * *Enclosure cool down*
   * *timeout*: after this time, the enclosure cool down is completed
   * *temerature*: when the air exhaust temperature is colder than this value, cool down is completed.
 * *Output checks*:
   * UMin / UMax : tolerable voltage range
   * FMin / FMax : tolerable frequency range
   * PMax: tolerable power range
