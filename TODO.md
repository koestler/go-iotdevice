# Todos
- Device: Remove c.shutdown ? use context instead?
- Device: implement supervisor which restarts devices on crash etc.; make it usable for modbus as well 
- handle case of startup when usb-serial is present but no Victron device is connected
- redo / check homeassistant integration using mqtt realtime messages
- victron: fetch VoltageCompensation depending on firmware version
- handle case when http device becomes unreachable (null all values, do not send telemetry)
- shelly http: create available / unavailable state and remove values when unavailable
- handle serial device reconnect

May  2 14:24:44 srv2 kernel: [594489.793425] usb 1-1.2-port2: disabled by hub (EMI?), re-enabling...
May  2 14:24:44 srv2 kernel: [594489.793955] usb 1-1.2.2: USB disconnect, device number 6
May  2 14:24:44 srv2 kernel: [594489.795267] ftdi_sio ttyUSB1: FTDI USB Serial Device converter now disconnected from ttyUSB1
May  2 14:24:44 srv2 kernel: [594489.795378] ftdi_sio 1-1.2.2:1.0: device disconnected
May  2 14:24:44 srv2 kernel: [594490.022584] usb 1-1.2.2: new full-speed USB device number 9 using xhci-hcd
May  2 14:24:44 srv2 kernel: [594490.131626] usb 1-1.2.2: New USB device found, idVendor=0403, idProduct=6015, bcdDevice=10.00
May  2 14:24:44 srv2 kernel: [594490.131662] usb 1-1.2.2: New USB device strings: Mfr=1, Product=2, SerialNumber=3
May  2 14:24:44 srv2 kernel: [594490.131678] usb 1-1.2.2: Product: VE Direct cable
May  2 14:24:44 srv2 kernel: [594490.131690] usb 1-1.2.2: Manufacturer: VictronEnergy BV
May  2 14:24:44 srv2 kernel: [594490.131703] usb 1-1.2.2: SerialNumber: VE6GBH44
May  2 14:24:44 srv2 kernel: [594490.135852] ftdi_sio 1-1.2.2:1.0: FTDI USB Serial Device converter detected
May  2 14:24:44 srv2 kernel: [594490.136064] usb 1-1.2.2: Detected FT-X
May  2 14:24:44 srv2 kernel: [594490.137538] usb 1-1.2.2: FTDI USB Serial Device converter now attached to ttyUSB4