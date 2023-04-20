# Todos
- RunDevice: merge deviceConfig / modbusConfig into single interface
- RunDevice: stateStorage should be interface not pointer; use same pattern as with *ModbusStruct & modbus.Modbus
- Device: Remove c.shutdown ? use context instead?
- Device: implement supervisor which restarts devices on crash etc.; make it usable for modbus as well 
- extend documentation
  - how to configure the tool
  - how to set up Victron usb-to-serial converters
- handle case of startup when usb-serial is present but no Victron device is connected
- redo / check homeassistant integration using mqtt realtime messages
- device: use context instead of shutdown channel?
- victron: fetch VoltageCompensation depending on firmware version
- handle case when http device becomes unreachable (null all values, do not send telemetry)
- shelly http: create available / unavailable state and remove values when unavailable
- handle serial device reconnect

Apr 19 10:06:19 srv1 kernel: [10075864.906581] usb 1-1.2.2: Detected FT-X
Apr 19 10:06:19 srv1 kernel: [10075864.909191] usb 1-1.2.2: FTDI USB Serial Device converter now attached to ttyUSB0
Apr 19 10:06:19 srv1 mtp-probe: checking bus 1, device 41: "/sys/devices/platform/soc/3f980000.usb/usb1/1-1/1-1.2/1-1.2.2"
Apr 19 10:06:19 srv1 mtp-probe: bus: 1, device: 41 was not an MTP device
Apr 19 10:06:19 srv1 mtp-probe: checking bus 1, device 41: "/sys/devices/platform/soc/3f980000.usb/usb1/1-1/1-1.2/1-1.2.2"
Apr 19 10:06:19 srv1 mtp-probe: bus: 1, device: 41 was not an MTP device
Apr 19 10:16:08 srv1 rngd[529]: stats: bits received from HRNG source: 84060064
Apr 19 10:16:08 srv1 rngd[529]: stats: bits sent to kernel pool: 83929728
Apr 19 10:16:08 srv1 rngd[529]: stats: entropy added to kernel pool: 83929728
Apr 19 10:16:08 srv1 rngd[529]: stats: FIPS 140-2 successes: 4199
Apr 19 10:16:08 srv1 rngd[529]: stats: FIPS 140-2 failures: 4
Apr 19 10:16:08 srv1 rngd[529]: stats: FIPS 140-2(2001-10-10) Monobit: 0
Apr 19 10:16:08 srv1 rngd[529]: stats: FIPS 140-2(2001-10-10) Poker: 1
Apr 19 10:16:08 srv1 rngd[529]: stats: FIPS 140-2(2001-10-10) Runs: 1
Apr 19 10:16:08 srv1 rngd[529]: stats: FIPS 140-2(2001-10-10) Long run: 2
Apr 19 10:16:08 srv1 rngd[529]: stats: FIPS 140-2(2001-10-10) Continuous run: 0
Apr 19 10:16:08 srv1 rngd[529]: stats: HRNG source speed: (min=331.797; avg=848.653; max=882.210)Kibits/s
Apr 19 10:16:08 srv1 rngd[529]: stats: FIPS tests speed: (min=4.252; avg=32.