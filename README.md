# go-iotdevice
[![Audit & Test](https://github.com/koestler/go-iotdevice/actions/workflows/audit.yml/badge.svg)](https://github.com/koestler/go-iotdevice/actions/workflows/audit.yml)
[![Docker Image CI](https://github.com/koestler/js-iotsensor/actions/workflows/docker-image.yml/badge.svg?branch=main)](https://github.com/koestler/js-iotsensor/actions/workflows/docker-image.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/koestler/go-iotdevice/v3.svg)](https://pkg.go.dev/github.com/koestler/go-iotdevice/v3)

This tool reads values from various IoT devices like solar charges directly connected by USB
or relay boards connected by ethernet and publishes those on [MQTT](http://mqtt.org/) servers.

Additionally, a REST- and websocket-API and a [web frontend](https://github.com/koestler/js-iotdevice)
is made available via an internal http server.

![Device overview](https://raw.githubusercontent.com/koestler/go-iotdevice-docs/main/external-overview.png)

The tool does not store any historical data. Instead,
[go-mqtt-to-influx](https://github.com/koestler/go-mqtt-to-influx) is can be used to write
to an [Influx Database](https://github.com/influxdata/influxdb). [Grafana](https://grafana.com/)
can be used to easily create custom dashboards showing the data.

The tool was written with the following two scenarios in mind:
* An off-grid holiday home installation running two batteries with
  Victron Energy [SmartSolar](https://www.victronenergy.com/solar-charge-controllers/bluesolar-mppt-150-35) / 
  [SmartShunt](https://www.victronenergy.com/battery-monitors/smart-battery-shunt) for solar and battery monitoring, 
 and a [Shelly 3EM](https://www.shelly.cloud/en-ch/products/product-overview/shelly-3-em) for generator power monitoring.
 The tool runs on a single [Raspberry Pi Zero 2 W](https://www.raspberrypi.com/products/raspberry-pi-zero-2-w/).
* Remote control of a generator set using a [Teracom TCW241](https://www.teracomsystems.com/ethernet/ethernet-io-module-tcw241/)
  for start/stop and temperature monitoring. Control is integrated into [Homea Assistant](https://www.home-assistant.io/)
  via MQTT.

## Supported protocols and devices

The tool currently implements the following devices, which are all used in an active project of mine.
However, it is made to be extended. Feel free to send pull requests or 
[create an issue](https://github.com/koestler/go-iotdevice/issues).

| Configuration section              | Kind=              | Name                                                                                                                                                                                                                                                | State                              | 
|------------------------------------|--------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------|
| [VictronDevcies](#Victron-devices) | Vedirect           | Victron Energy [BlueSolar MPPT](https://www.victronenergy.com/solar-charge-controllers/mppt7510) (different versions)                                                                                                                               | production ready                   |
| [VictronDevcies](#Victron-devices) | Vedirect           | Victron Energy [SmartSolar MPPT](https://www.victronenergy.com/solar-charge-controllers/smartsolar-150-35) (different versions)                                                                                                                     | production ready                   |
| [VictronDevcies](#Victron-devices) | Vedirect           | Victron Energy Battery Monitor [BMV 700](https://www.victronenergy.com/battery-monitors/bmv-700)                                                                                                                                                    | production ready                   |
| [VictronDevcies](#Victron-devices) | Vedirect           | Victron Energy Battery Monitor [BMV 702](https://www.victronenergy.com/battery-monitors/bmv-702)                                                                                                                                                    | production ready                   |
| [VictronDevcies](#Victron-devices) | Vedirect           | Victron Energy Battery Monitor [BMV-712 Smart](https://www.victronenergy.com/battery-monitors/bmv-712-smart)                                                                                                                                        | production ready                   |
| [VictronDevcies](#Victron-devices) | Vedirect           | Victron Energy [SmartShunt](https://www.victronenergy.com/battery-monitors/smart-battery-shunt)                                                                                                                                                     | production ready                   |
| [VictronDevcies](#Victron-devices) | Vedirect           | Victron Energy [Phoenix Inverter](https://www.victronenergy.com/inverters)                                                                                                                                                                          | production ready                   |
| [VictronDevcies](#Victron-devices) | Vebus              | Victron Energy [Multiplus](https://www.victronenergy.com/inverters-chargers/multiplus-12v-24v-48v-800va-3kva)                                                                                                                                       | in development, see v3vebus branch |
| [ModbusDevices](#Modbus-devices)   | WaveshareRtuRelay8 | [Waveshare Industrial Modbus RTU 8-ch Relay Module](https://www.waveshare.com/modbus-rtu-relay.htm)                                                                                                                                                 | production ready                   |
| [ModbusDevices](#Modbus-devices)   | Finder7M38         | [Finder TYPE 7M.38 - bi-directional multi-functional energy meters](https://www.findernet.com/en/uk/series/7m-series-smart-energy-meters/type/type-7m-38-three-phase-multi-function-bi-directional-energy-meters-with-backlit-matrix-lcd-display/)  | beta testing                       |
| [HttpDevcies](#http-devices)       | Teracom            | Teracom [TCW241](https://www.teracomsystems.com/ethernet/ethernet-io-module-tcw241/) industrial relay/sensor board                                                                                                                                  | production ready                   | 
| [HttpDevcies](#http-devices)       | ShellyEm3          | Shelly [3EM](https://www.shelly.cloud/en-ch/products/product-overview/shelly-3-em) 3-phase energy power monitor                                                                                                                                     | production ready                   |
| [MqttDevcies](#mqtt-devices)       | GoIotdeviceV3      | Another go-iotdevice instance connected to the same MQTT server                                                                                                                                                                                     | production ready                   |

See [Devices](#devices) section on how to configure each.

## Terminology

This project uses the following terminology:

* **Device** refers to a physical unit like a solar charger, a relay board
* A **Register** is a measurement or output that a device can have.
  E.g. "Battery Voltage" or "Relay 0". Each device can have multiple registers. For some devices the list
  of registers is simply defined by the type of device (e.g. Victron Energy MPPT 100 | 50) and for others it is configurable
  on the device (e.g. the TCW241).
  A Register has a technical name (alphanumeric, no spaces, used as the key is various JSON objects, e.g. "BatteryVoltage"),
  a description (shown in the frontend, e.g. "Battery Voltage"), a Unit (e.g. "mV"), and a type (string, float, enum).
* A **value** is a reference to a register plus a number/string/enum-index depending on the type of register.
  E.g.: A value can be 13.80 and reference to the register "BatteryVoltage".
* A **view** is used in the HTTP server and the front-end.
  The front-end shows different subsets of devices/registers on different routes. A view defines such a subset.

## Deployment

<details>
<summary>
Deployment without docker
</summary>
I use docker to deploy this tool.
Alternatively, you can use `go install` to build binary locally.
If you need the [front-end](https://github.com/koestler/js-iotdevice), you have to manually `npm run build` it and copy the files to the correct location.
Also, the swagger documentation will be missing. See [Local Development](#Local-development) for more details.

```bash
go install github.com/koestler/go-iotdevice/v3@latest
curl https://raw.githubusercontent.com/koestler/go-iotdevice/main/documentation/config.yaml -o config.yaml

# optional download the full and commented configuration file for reference 
curl https://raw.githubusercontent.com/koestler/go-iotdevice/main/documentation/full-config.yaml -o full-config.yaml
# adapt config.yaml and configure devices

# start the tool
go-iotdevice --config=config.yaml
```
</details>

### Docker

There are [GitHub actions](https://github.com/koestler/go-iotdevice/actions/workflows/docker-image.yml)
to automatically cross-compile amd64, arm64, and arm/v7
publicly available [docker images](https://github.com/koestler/go-iotdevice/pkgs/container/go-iotdevice).
The docker-container is built on top of Alpine, the binary is `/go-iotdevice` and the config is
expected to be at `/config.yaml`. The container runs as a non-root user `app`.

The GitHub tags use semantic versioning and whenever a tag like v2.3.4 is built, it is pushed to docker tags
v2, v2.3, and v2.3.4.

For auto-restart on system reboots, configuration, and networking I use `docker compose`. Here is an example file:
```yaml
# documentation/docker-compose.yml

services:
  go-iotdevice:
    restart: always
    image: ghcr.io/koestler/go-iotdevice:v3
    volumes:
      - ${PWD}/config.yaml:/config.yaml:ro
      #- ${PWD}/auth.passwd:/auth.passwd:ro
      - /dev:/dev
    privileged: true # used to access serial devices
    group_add: # add app user running to software to the dialout group
      - dialout

```

### Configuration
The configuration is stored in a single yaml file. By default, it is read from `./config.yaml`.
This can be changed using the `--config=another-config.yaml` command line option.

There are mandatory fields and there are optional fields which have reasonable default values.

See [Explained Full Configuration](#explained-full-configuration) for a complete list of all available configuration options.

### Quick setup
[Install Docker](https://docs.docker.com/engine/install/) first.

```bash
# create a directory for the docker-composer project and config file
mkdir -p /srv/dc/go-iotdevice # or wherever you want to put docker-compose files
cd /srv/dc/go-iotdevice
curl https://raw.githubusercontent.com/koestler/go-iotdevice/main/documentation/docker-compose.yml -o docker-compose.yml
curl https://raw.githubusercontent.com/koestler/go-iotdevice/main/documentation/config.yaml -o config.yaml

# optional download the full and commented configuration file for reference 
curl https://raw.githubusercontent.com/koestler/go-iotdevice/main/documentation/full-config.yaml -o full-config.yaml
# adapt config.yaml and configure devices

# start the container
docker compose up -d

# optional: check the log output to see how it's going
docker compose logs -f

# when config.yaml is changed, the container needs to be restarted
docker compose restart

# upgrade to the newest tag
docker compose pull
docker compose up -d
```

## Devices
All devices have in common that this software extracts a relatively static set of registers
(list of available measurements/outputs) and repetitively polls those registers and extracts current values (readings).

How the list of registers and the values are gathered depends on the type of device / connection.

### Victron devices
All Victron Energy solar chargers, some inverters and the BMV devices share the same VE.Direct protocol.
It is a binary protocol and requires the user to know the addresses of registers and how to decode enums.

The easiest way of connection is to use a [VE.Direct to USB interface](https://www.victronenergy.com/accessories/ve-direct-to-usb-interface).

This tool reads the deviceId, which is present in all devices, and then uses this id to determine if it is a
solar charger, an inverter or a battery monitor. A hardcoded list of known registers for this device is than used.

Configuration:

```yaml
VictronDevices:
  main-bmv: # used for reference in the view section
    Kind: Vedirect # tells the tool that we use the VE.Direct protocol
    Device: /dev/serial/by-id/usb-VictronEnergy_BV_VE_Direct_cable_VEXXXXX-if00-port0
    # Device: Could also be /dev/ttyUSB0, make sure the device is present / accessible
    # connect the interface and use ls -la /dev/serial/by-id/ to see what devices are available
    # Using directly a /dev/ttyUSB0 device can result in chaos after a reboot when multiple interfaces are connected 
    Filter:
      # The tool does not know if you have an auxiliary battery connected. You might want to skip some unused registers.
      SkipRegisters:
        - AuxVoltage
        - BatteryTemperature
        - MidPointVoltage
        - MidPointVoltageDeviation
        - AuxVoltageMinimum
        - AuxVoltageMaximum
```

### Modbus devices
[Modbus](https://en.wikipedia.org/wiki/Modbus) [RS485](https://en.wikipedia.org/wiki/RS-485) is an old industry bus
used in various devices like power meters. It has the advantage of connecting multiple devices via one serial device.
There are some relatively cheap relay boards
(e.g. [Waveshare Industrial Modbus RTU 8-ch Relay Module](https://www.waveshare.com/modbus-rtu-relay-b.htm))
available, which have a much lower power consumption when compared to ethernet-connected devices.

First, you need to configure the serial device connected to the bus.
Secondly, you need to configure each device on the bus individually.
Make sure that all devices on the bus have a unique address (use external tools like the Python scripts provided by some vendors).
Alternatively, you can add multiple Modbus serial services.

Configuration:

```yaml
Modbus:
  bus0:
    Device: /dev/ttyACM0 # the serial device
    BaudRate: 9600 # Choose a BaudRate supported by all connected devices, Often the BaudRate can be changed.

ModbusDevices:
  relay-board: # used for reference in the view section
    Bus: bus0 # the name of the bus as chosen above
    Kind: WaveshareRtuRelay8 # the type of board used
    Address: 0x01 # the address of the device on the Modbus
    Relays:
      # The tool does not know what you have connected to the relays. It simply gives them names like CH1, CH2, ...
      # use this section to add nice descriptions and labels for the open and closed state
      # This is shown in the HTTP frontend and also exposed via MQTT.
      CH1:
        Description: Main Inverter
        OpenLabel: Off
        ClosedLabel: On
      CH2:
        Description: Main To Aux Transfer
        OpenLabel: On
        ClosedLabel: Off
    Filter: # You can skip unused outputs
      SkipRegisters: [CH3, CH4, CH5, CH6, CH7, CH8]
```

### Http devices
HTTP devices do not have a direct serial connection to go-iotdevice.
Instead, they must be reachable via a network connection which makes them very versatile.

Configuration:

```yaml
HttpDevices:
  control0:
    Url: http://control0/ # howto reach the device; can also be http://192.168.0.42/
    Kind: Teracom
    Username: admin # optionally, if you have a login configured, add the cleartext user/password here
    Password: letMeIn
    Filter:
      # Some devices have quite an extensive list of registers
      SkipCategories:
        # use this list to skip certain unused sections (all digital inputs)
        - Analog Inputs
        - Virtual Inputs
        - Digital Inputs
        - Alarms
      # or skip certain registers
      SkipRegisters: [R2, R3, R4]
```

### MQTT devices
MQTT devices receive values from an MQTT broker. E.g. if you have multiple computers running go-iotdevice,
and you want to have all the devices in the same front-end.

For this to work, you need to define a MqttDevice and then configure a MqttClient to connect to a broker and send data
to this MQTT device. To allow for redundant setups, you can have multiple MqttClients sending data to the same
MqttDevice.

For the `GoIotdeviceV3` kind, you need to give it the Topic of the structure (see [Mqtt Interface Structure](#structure))
message. This structure is then used to generate the list of registers and subscribe to the
[Realtime](#realtime) and [Telemetry](#telemetry) messages. They work both at the same time.
Additionally, if the [Command](#command) topic is available, it is used to control outputs.

```yaml
MqttClients:
  local:
    Broker: tcp://127.0.0.1:1833
    User: user
    Password: 424242

    MqttDevices:
      main-bmv:
        MqttTopics:
          - go-iotdevice/struct/main-bmv

MqttDevices:
  main-bmv:
    Kind: GoIotdeviceV3
```

## Http Interface
There is a stable REST-API to fetch the views, devices, registers, and values.
Additionally, patch requests are implemented to set a controllable register (e.g. an output of a relay board).
This API is used by the build-in front-end and can also be used for custom integrations.
See /api/v2/docs and /api/v2/docs/swagger.json for built-in swagger documentation.

## Authentication
The tool can use [JWT](https://jwt.io/) to make certain views only available after a login. The user database
is stored in an Apache htaccess file which can be changed without restarting the server. 

### Configuration

There are two relevant sections in the configuration file:

1. Adding the `Authentication:` section enables the login/authentication mechanism.
The `JwtSecret` is used to sign the tokens. When left unconfigured, a new random secret is generated on each
startup of the backend. This results in all users being logged out after each restart of the server.
It's therefore best to hardcode a random secret.

2. Per `View`, you can define a list of `AllowedUsers`. When the list is present and has at least one entry, only
usernames on that list can access this view. If the list is empty, all users in the user database can access it. 

### User database
The only supported authentication backend at the moment is a simple Apache htaccess file. Set it up as follows:

```bash
htpasswd -c auth.passwd lorenz
# enter password twice
htpasswd auth.passwd daniela
# enter another password twice
```

### Nginx reverse proxy
I normally run this service behind a [Nginx](https://nginx.org/en/) server configured as a reverse proxy. It can take care of:
 * Serving multiple different applications on the same address using [SNI](https://en.wikipedia.org/wiki/Server_Name_Indication).
 * Caching on a fast cloud server in front of a device connected via a slow mobile connection.
 * https termination

#### Setup
It's assumed that you understand each of the following steps in detail. It's just to make the setup steps as quick as possible. 

```bash
# install nginx / curl
apt update && apt install -y nginx curl

# define on what URL the tool shall be reachable
SITE=foo.example.com
CONFFILE=$SITE".conf"

# create nginx configuration
cd /etc/nginx/sites-available/
curl https://raw.githubusercontent.com/koestler/go-iotdevice/main/documentation/nginx-example.conf -o $CONFFILE
sed -i "s/example.com/$SITE/g" $CONFFILE
cd ../sites-enabled
ln -s "../sites-available/"$CONFFILE

# edit the proxy_pass directive to the correct address for the service
emacs $CONFFILE 

# setup certbot. use something like this:
mkdir -p /srv/www-acme-challenge/$SITE
certbot certonly --authenticator webroot --webroot-path /srv/www-acme-challenge/$SITE -d $SITE

# reload the nginx config
service nginx reload 
```

## MQTT Interface
This tool can connect to one or multiple MQTT servers and provide publish messages regarding the availability
(whether a device is online), available registers, and current values.
There are 5 different kinds of outgoing messages
(**Availability**, **Structure**, **Telemetry**, **Realtime**, **HomeassistantDiscovery**)
and one kind of incoming message (**Command**).

For each kind of message, you can configure the MQTT retain flag (server stores messages) 
and what devices/registers shall be transmitted.

### Availability of go-iotdevice
go-iotdevice sends the message `online` whenever it connects to an MQTT broker and sends `offline` when it properly shuts down.
Additionally, the `offline` message is registered as the last will message and the broker will distribute it automatically
when the connection breaks.

Examples:
```
go-iotdevice/avail/go-iotdevice online
go-iotdevice/avail/go-iotdevice offline
```

### Availability of a device
Whenever a device becomes connected/disconnected (serial connection established/lost) an `online`/`offline`
message per device can be sent. However, MQTT only allows for one last will message. Therefore to reliably check if
a device is available, both, the availability message of the client and the device must be checked.

Examples:
```
go-iotdevice/avail/my-device online
go-iotdevice/avail/my-device offline
```

### Structure
For other instances of go-iotdevice or also for third-party software to know when devices are available and
what registers they have, a message containing the current structure of the device is published. It contains a list
of registers as well as the topics for the availability, telemetry, real-time, and command message such that the receiver
knows where to subscribe to get current values/control the outputs.

Structure messages are sent when the device comes online for the first time by default (Interval=0s) with the retain
flags set (the broker stores those messages for new clients).
Alternatively, it can also be sent repeatedly (Interval>0).

Example:
```
go-iotdevice/struct/my-device {
  "Avail":["go-iotdevice/avail/go-iotdevice","go-iotdevice/avail/my-device"],
  "Tele":"go-iotdevice/tele/my-device",
  "Real":"go-iotdevice/real/my-device/%RegisterName%",
  "Cmnd":"go-iotdevice/cmnd/my-device/%RegisterName%",
  "Regs":[
    {"Cat":"Analog Inputs","Name":"AI1","Desc":"inputA","Type":"number","Unit":"V","Sort":100,"Cmnd":false},
    {"Cat":"Digital Inputs","Name":"DI1","Desc":"Digital Input 1","Type":"enum","Enum":{"0":"OPEN","1":"CLOSED"},"Sort":300,"Cmnd":false},
    {"Cat":"Relays","Name":"R1","Desc":"Relay 1","Type":"enum","Enum":{"0":"OFF","1":"ON","2":"in pulse"},"Sort":400,"Cmnd":true},
    {"Cat":"Alarms","Name":"AI2Alarm","Desc":"Analog Input 2","Type":"enum","Enum":{"0":"OK","1":"ALARMED"},"Sort":501,"Cmnd":false},
    {"Cat":"General","Name":"Time","Desc":"Time","Type":"string","Sort":602,"Cmnd":false},
    {"Cat":"Device Info","Name":"DeviceName","Desc":"Device Name","Type":"string","Sort":700,"Cmnd":false},
  ]
}
```

Since MQTT payloads are sent uncompressed, size matters and fields are abbreviated:
Avail=AvailabilityTopics, Tele=TelemetryTopic, Real=RealtimeTopic, Cmnd=CommandTopic/Writable, Regs=Registers, Cat=Category, Desc=Description

### Telemetry
There are two ways to receive values. Telemetry messages are sent periodically (1s by default) per device and contain
all the current values.

Example:
```
dev0/tele/teracom {
  "Time":"2023-11-15T16:55:42+01:00",
  "NextTelemetry":"2023-11-15T16:55:52+01:00",
  "Model":"Teracom",
  "NumericValues":{
    "AI1":{"Cat":"Analog Inputs","Desc":"inputA","Val":0.02,"Unit":"V"},
  },
  "TextValues":{
    "DeviceName":{"Cat":"Device Info","Desc":"Device Name","Val":"TCW241"},
    "Time":{"Cat":"General","Desc":"Time","Val":"18:58:19"}
  },
  "EnumValues":{
    "AI2Alarm":{"Cat":"Alarms","Desc":"Analog Input 2","Idx":1,"Val":"ALARMED"},
    "DI1":{"Cat":"Digital Inputs","Desc":"Digital Input 1","Idx":0,"Val":"OPEN"},
    "R1":{"Cat":"Relays","Desc":"Relay 1","Idx":1,"Val":"ON"},
  }
}
```

For easy parsing, values are separated by type. To make the telemetry without the struct message, it also
includes Cat=Category, Desc=Description, and Unit fields.

### Realtime
Real-time messages are sent per device and register only when a value changes. They can either be sent
immediately (Interval=0) or debounced (Interval>0). This is useful for some devices that change some values very often.

Examples:
```
go-iotdevice/real/my-device/AI1 {"NumVal":0.02}
go-iotdevice/real/my-device/DI1 {"EnumIdx":0}
go-iotdevice/real/my-device/Time {"TextVal":"19:00:24"}
```

Real-time messages are small and only contain the value. The unit and nice names must be retrieved separately (e.g. via the structure messages).

### Command
This tool can subscribe to command topics to receive commands to set an output to a specific state (e.g. switch a relay).
The topic encodes the device and the register name of the output that shall be changed. The payload has the same format
as real-time messages and encodes the desired value.

Examples:
```
go-iotdevice/cmnd/my-device/R1 {"EnumIdx": 0}
go-iotdevice/cmnd/my-device/R1 {"EnumIdx": 1}
```

Example of howto switch Relay 1 of a device called dev0 into the on position via mosquitto:

```bash
mosquitto_pub -h 172.19.0.4 -t dev1/cmnd/dev0/R1 -m "{\"EnumIdx\": 1}"
```

### HomeassistantDiscovery
These messages are such that Homeassistant automatically shows read-only registers as sensors and writable registers
as switches. See [Home Assistant MQTT](https://www.home-assistant.io/integrations/mqtt/#mqtt-discovery).

Discovery messages for sensors are only sent for devices/registers for which real-time messages are active because
they are used to transmit the actual values. Switches are only advertised for registers for which the command topic is active. 
Use filters in the Realtime and Command configuration section to restrict what devices/registers are shown in Homeassistant.
Also, consider setting `Realtime->Interval=500ms`. Homeassistant can easily be overloaded by hundreds of registers.

Examples:
```
homeassistant/sensor/go-iotdevice/my-device-ai1/config {
  "uniq_id":"my-device-ai1",
  "name":"my-device inputA",
  "avty":[{"t":"go-iotdevice/avail/go-iotdevice"},{"t":"go-iotdevice/avail/my-device"}],
  "avty_mode":"all",
  "stat_t":"go-iotdevice/real/my-device/AI1",
  "val_tpl":"{{ value_json.NumVal }}",
  "unit_of_meas":"V"
}
homeassistant/switch/go-iotdevice/my-device-r1/config {
  "uniq_id":"my-device-r1",
  "name":"my-device Relay 1",
  "avty":[{"t":"go-iotdevice/avail/go-iotdevice"},{"t":"go-iotdevice/avail/my-device"}],
  "avty_mode":"all",
  "cmd_t":"go-iotdevice/cmnd/my-device/R1",
  "stat_t":"go-iotdevice/real/my-device/R1",
  "val_tpl":"{{ value_json.EnumIdx }}",
  "pl_off":"{\"EnumIdx\":0}",
  "pl_on":"{\"EnumIdx\":1}",
  "stat_off":"0",
  "stat_on":"1"
}
```

## Development

### Local development
For development, this backend can be compiled and run locally.
In addition, it's then best to also run the [front-end](https://github.com/koestler/js-iotdevice) locally. 

This tool can proxy requests to a local server serving the front-end. Use e.g.:

```yaml
HttpServer:                                                # optional, when missing: http server is not started
  Bind: "[::1]"                                            # mandatory, use [::1] (ipv6 loopback) to enable on both ipv4 and 6 and 0.0.0.0 to only enable ipv4
  Port: 8000                                               # optional, default 8000
  FrontendProxy: "http://127.0.0.1:3000/"
 ```  
 
Build and run: 
  
```bash
go build && ./go-iotdevice
```

### Compile and run inside docker
Alternatively, if you don't have Golang installed locally, you can compile and run 

```bash
docker build -f docker/Dockerfile -t go-iotdevice .
docker run --rm --name go-iotdevice -p 127.0.0.1:8000:8000 \
  -v "$(pwd)"/documentation/config.yaml:/config.yaml:ro \
  go-iotdevice
```

### Generate swagger documentation
```bash
go install github.com/swaggo/swag/cmd/swag@latest
go generate docs.go
```

### Run tests
[gomock](https://github.com/uber-go/mock) is used to generate stubs and mocks for the unit tests.

```bash
go install go.uber.org/mock/mockgen@latest
go generate ./...
go test ./...
```

### Update README.md
```bash
npx embedme README.md
```

## Explained Full Configuration
```yaml
# documentation/full-config.yaml

Version: 2                                                 # mandatory, configuration file format; must be set to 2 for >=v3 of this tool. Older formats are not supported anymore.
ProjectTitle: Configurable Title of Project                # optional, default go-iotdevice, title is shown in the http frontend
LogConfig: true                                            # optional, default true, outputs the full configuration structure including used defaults on startup
LogWorkerStart: true                                       # optional, default true, outputs what devices and mqtt clients are started
LogStateStorageDebug: false                                # optional, default false, outputs all write to the internal state value storage
LogCommandStorageDebug: false                              # optional, default false, outputs all write to the internal command value storage

HttpServer:                                                # optional, when missing: http server is not started
  Bind: "[::1]"                                            # mandatory, use [::1] (ipv6 loopback) to enable on both ipv4 and 6 and 0.0.0.0 to only enable ipv4
  Port: 8000                                               # optional, default 8000, what tcp port to listen on, low-ports like 80 only work when started as root
  LogRequests: true                                        # optional, default true, enable the http access log to stdout
  # configure FrontendProxy xor FrontendPath
  #FrontendProxy: "http://127.0.0.1:3000/"                 # optional, default deactivated; proxies the frontend to another server; useful for development
  FrontendPath: ./frontend-build/                          # optional, default "./frontend-build/": path to a static frontend build
  FrontendExpires: 5m                                      # optional, default 5min, what cache-control header to send for static frontend files
  ConfigExpires: 1m                                        # optional, default 1min, what cache-control header to send for configuration endpoints
  LogDebug: false                                          # optional, default false, output debug messages related to the http server

Authentication:                                            # optional, when missing: login is disabled
  #JwtSecret: 'insert a random string here and uncomment'  # optional, default new random string on startup, used to sign the JWT tokens
                                                           # use a fixed, secure, random value (e.g. `pwgen -s 64 1`) to allow users to stay logged in on restart
  JwtValidityPeriod: 1h                                    # optional, default 1h, users are logged out after this time
  HtaccessFile: ./auth.passwd                              # mandatory, where the file generated by htpasswd can be found

MqttClients:                                               # optional, when empty, no mqtt connection is made
  local:                                                   # mandatory, an arbitrary name used for logging and for referencing in other config sections
    Broker: tcp://mqtt.example.com:1883                    # mandatory, the URL to the server, use tcp:// or ssl://
    ProtocolVersion: 5                                     # optional, default 5, must be 5 always, only mqtt protocol version 5 is supported

    User: dev                                              # optional, default empty, the username used for authentication
    Password: zee4AhRi                                     # optional, default empty, the plain text password used for authentication
    #ClientId: go-iotdevice                                # optional, default go-iotdevice-UUID (-> random on each startup), mqtt client id, make sure it is unique per mqtt-server

    KeepAlive: 1m                                          # optional, default 60s, how often a ping is sent to keep the connection alive
    ConnectRetryDelay: 10s                                 # optional, default 10s, when disconnected: after what delay shall a connection attempt is made
    ConnectTimeout: 5s                                     # optional, default 5s, how long to wait for the SYN+ACK packet, increase on slow networks
    TopicPrefix: go-iotdevice/                             # optional, %Prefix% is replaced with this string
    ReadOnly: false                                        # optional, default false, when true, no messages are sent to the server (overriding MaxBacklogSize, AvailabilityEnable, StructureEnable, TelemetryEnable, RealtimeEnable)
    MaxBacklogSize: 256                                    # optional, default 256, max number of mqtt messages to store when tconnection is offline

    MqttDevices:                                           # optional, default empty, which mqtt devices shall receive messages from this client
      bmv1:                                                # mandatory, the identifier of the MqttDevice
        MqttTopics:                                        # mandatory, at least 1 topic must be defined
          - stat/go-iotdevice/bmv1/+                       # what topic to subscribe to; must match StructureTopic of the sending device

    AvailabilityClient:
      Enabled: true                                        # optional, default true, whether to send online messages and register an offline message as will
      TopicTemplate: '%Prefix%avail/%ClientId%'            # optional, what topic to use for online/offline messages of the go-iotdevice instance
      Retain: true                                         # optional, default true, the mqtt retain flag for availability messages
      Qos: 1                                               # optional, default 1, what quality-of-service level shall be used

    AvailabilityDevice:
      Enabled: true                                        # optional, default true, whether to send online messages and register an offline message as will
      TopicTemplate: '%Prefix%avail/%DeviceName%'          # optional, what topic to use for online/offline messages of a specific device
      Retain: true                                         # optional, default true, the mqtt retain flag for availability messages
      Qos: 1                                               # optional, default 1, what quality-of-service level shall be used
      Devices:                                             # optional, default all, a list of devices to match
        bmv0:                                              # use device identifiers of the VictronDevices, ModbusDevices etc. sections

    Structure:
      Enabled: true                                        # optional, default false, whether to send messages containing the list of registers/types
      TopicTemplate: '%Prefix%struct/%DeviceName%'         # optional, what topic to use for structure messages
      Interval: 0s                                         # optional, default 0, 0 means disabled only send initially, otherwise the structure is repeated after this interval (useful when retain is false)
      Retain: true                                         # optional, default true, the mqtt retain flag for structure messages
      Qos: 1                                               # optional, default 1, what quality-of-service level shall be used
      Devices:                                             # optional, default all, a list of devices to match
        bmv0:                                              # use device identifiers of the VictronDevices, ModbusDevices etc. sections
          Filter:                                          # optional, default include all, defines which registers are shown in the view,
                                                           # The rules are applied in order beginning with IncludeRegisters (highest priority) and ending with DefaultInclude (lowest priority).
            IncludeRegisters:                              # optional, default empty, if a register is on this list, it is returned
            SkipRegisters:                                 # optional, default empty, if a register is on this list, it is not returned
            IncludeCategories:                             # optional, default empty, all registers of the given category that are not explicitly skipped are returned
            SkipCategories:                                # optional, default empty, all registers of the given category that are not explicitly included are not returned
            DefaultInclude: True                           # optional, default true,  whether to return the registers that do not match any include/skip rule


    Telemetry:
      Enabled: true                                        # optional, default false, whether to send telemetry messages (one per device)
      TopicTemplate: '%Prefix%tele/%DeviceName%'           # optional, what topic to use for telemetry messages
      Interval: 1s                                         # optional, default 1s, how often to sent telemetry mqtt messages
      Retain: false                                        # optional, default false, the mqtt retain flag for telemetry messages
      Qos: 1                                               # optional, default 1, what quality-of-service level shall be used
      Devices:                                             # optional, default all, a list of devices to match
        bmv0:                                              # use device identifiers of the VictronDevices, ModbusDevices, etc. sections
          Filter:                                          # optional, default include all, defines which registers are show in the view,
            # The rules are applied in order beginning with IncludeRegisters (highest priority) and ending with DefaultInclude (lowest priority).
            IncludeRegisters:                              # optional, default empty, if a register is on this list, it is returned
            SkipRegisters:                                 # optional, default empty, if a register is on this list, it is not returned
            IncludeCategories:                             # optional, default empty, all registers of the given category that are not explicitly skipped are returned
            SkipCategories:                                # optional, default empty, all registers of the given category that are not explicitly included are not returned
            DefaultInclude: True                           # optional, default true,  whether to return the registers that do not match any include/skip rule

    Realtime:
      Enabled: true                                        # optional, default false, whether to enable sending real-time messages
      TopicTemplate: '%Prefix%real/%DeviceName%/%RegisterName%' # optional, what topic to use for real-time messages
      Interval: 0s                                         # optional, default 0; 0 means send immediately when a value changes, otherwise only changed values are sent once per interval
      Retain: false                                        # optional, default false, the mqtt retain flag for real-time messages
      Qos: 1                                               # optional, default 1, what quality-of-service level shall be used
      Devices:                                             # optional, default all, a list of devices to match
        bmv0:                                              # messages are only sent for this device
          Filter:                                          # optional, default include all, defines which registers are show in the view,
                                                           # The rules are applied in order beginning with IncludeRegisters (highest priority) and ending with DefaultInclude (lowest priority).
            IncludeRegisters:                              # optional, default empty, if a register is on this list, it is returned
              - BatteryVoltage                             # the BatteryVoltage register is sent no matter if it's category is listed unter categories
              - Power
            SkipRegisters:                                 # optional, default empty, if a register is on this list, it is not returned
            IncludeCategories:                             # optional, default empty, all registers of the given category that are not explicitly skipped are returned
              - Essential                                  # all registers of the category essential are sent; no matter if thy are listed in registers
            SkipCategories:                                # optional, default empty, all registers of the given category that are not explicitly included are not returned
            DefaultInclude: False                          # optional, default true,  whether to return the registers that do not match any include/skip rule

    HomeassistantDiscovery:
      Enabled: true                                        # optional, default false, whether to enable sending homeassistant auto-disovery messages
      TopicTemplate: 'homeassistant/%Component%/%NodeId%/%ObjectId%/config' # optional, topic to use for homeassistant discovery messages
      Interval: 0s                                         # optional, default 0, 0 means disabled only send initially, otherwise the disovery messages are repeated after this interval (useful when retain is false)
      Retain: false                                        # optional, default false, the mqtt retain flag for homeassistant disovery messages
      Devices:                                             # optional, default all, a list of devices to match
        bmv0:                                              # use device identifiers of the VictronDevices, ModbusDevices, etc. sections
          Filter:                                          # optional, default include all, defines which registers are shown in the view,
            # The rules are applied in order beginning with IncludeRegisters (highest priority) and ending with DefaultInclude (lowest priority).
            IncludeRegisters:                              # optional, default empty, if a register is on this list, it is returned
            SkipRegisters:                                 # optional, default empty, if a register is on this list, it is not returned
            IncludeCategories:                             # optional, default empty, all registers of the given category that are not explicitly skipped are returned
            SkipCategories:                                # optional, default empty, all registers of the given category that are not explicitly included are not returned
            DefaultInclude: True                           # optional, default true,  whether to return the registers that do not match any include/skip rule

    Command:
      Enabled: true                                        # optional, default false, whether to receive and execute command messages
      TopicTemplate: '%Prefix%cmnd/%DeviceName%/%RegisterName%' # optional, what topic to use for real-time messages
      Qos: 1                                               # optional, default 1, what quality-of-service level shall be used
      Devices:                                             # optional, default all, a list of devices to match
        bmv0:                                              # messages are only sent for this device
          Filter:                                          # optional, default include all, defines which registers are show in the view,
            # The rules are applied in order beginning with IncludeRegisters (highest priority) and ending with DefaultInclude (lowest priority).
            IncludeRegisters:                              # optional, default empty, if a register is on this list, it is returned
              - BatteryVoltage                             # the BatteryVoltage register is sent no matter if its category is listed under categories
              - Power
            SkipRegisters:                                 # optional, default empty, if a register is on this list, it is not returned
            IncludeCategories:                             # optional, default empty, all registers of the given category that are not explicitly skipped are returned
              - Essential                                  # all registers of the category essential are sent; no matter if they are listed in registers
            SkipCategories:                                # optional, default empty, all registers of the given category that are not explicitly included are not returned
            DefaultInclude: False                          # optional, default true, whether to return the registers that do not match any include/skip rule

    LogDebug: false                                        # optional, default false, very verbose debug log of the mqtt connection
    LogMessages: false                                     # optional, default false, log all incoming mqtt messages

Modbus:                                                    # optional, when empty, no modbus handler is started
  bus0:                                                    # mandatory, an arbitrary name used for logging and for referencing in other config sections
    Device: /dev/ttyACM0                                   # mandatory, the RS485 serial device
    BaudRate: 4800                                         # mandatory, eg. 9600
    ReadTimeout: 100ms                                     # optional, default 100ms, how long to wait for a response
    LogDebug: false                                        # optional, default false, verbose debug log

VictronDevices:                                            # optional, a list of Victron Energy devices to connect to
  bmv0:                                                    # mandatory, an arbitrary name used for logging and for referencing in other config sections
    Device: /dev/serial/by-id/usb-VictronEnergy_BV_VE_Direct_cable_VEHTVQT-if00-port0 # mandatory except if Kind: Random*, the path to the usb-to-serial converter
    Kind: Vedirect                                         # mandatory, possibilities: Vedirect, RandomBmv, RandomSolar, always set to Vedirect expect for development
    PollInterval: 500ms                                    # optional, default 0.5s, how often to fetch the registers
    IoLog:                                                 # optional, default empty, path to a file where the raw io is logged
    Filter:                                                # optional, default include all, defines which registers are show in the view,
                                                           # The rules are applied in order beginning with IncludeRegisters (highest priority) and ending with DefaultInclude (lowest priority).
      IncludeRegisters:                                    # optional, default empty, if a register is on this list, it is returned
      SkipRegisters:                                       # optional, default empty, if a register is on this list, it is not returned
        - Temperature                                      # for BMV devices without a temperature sensor connected
        - AuxVoltage                                       # for BMV devices without a mid- or starter-voltage reading
      IncludeCategories:                                   # optional, default empty, all registers of the given category that are not explicitly skipped are returned
      SkipCategories:                                      # optional, default empty, all registers of the given category that are not explicitly included are not returned
        - Settings                                         # for solar devices it might make sense to not fetch/output the settings
      DefaultInclude: True                                 # optional, default true,  whether to return the registers that do not match any include/skip rule
    RestartInterval: 200ms                                 # optional, default 200ms, how fast to restart the device if it fails / disconnects
    RestartIntervalMaxBackoff: 1m                          # optional, default 1m; when it fails, the restart interval is exponentially increased up to this maximum
    LogDebug: false                                        # optional, default false, enable debug log output
    LogComDebug: false                                     # optional, default false, enable a verbose log of the communication with the device

ModbusDevices:                                             # optional, a list of devices connected via ModBus
  modbus-rtu0:                                             # mandatory, an arbitrary name used for logging and for referencing in other config sections
    Bus: bus0                                              # mandatory, the identifier of the modbus to use
    Kind: WaveshareRtuRelay8                               # mandatory, type/model of the device; possibilities: WaveshareRtuRelay8, Finder7M38
    Address: 0x01                                          # mandatory, the modbus address of the device in hex as a string, e.g. 0x0A
    Relays:                                                # optional: a map of custom labels for the relays
      CH1:
        Description: Lamp                                  # optional: show the CH1 relay as "Lamp" in the frontend
        OpenLabel: Off                                     # optional, default "open", a label for the open state
        ClosedLabel: On                                    # optional, default "closed", a label for the closed state
    PollInterval: 1s                                       # optional, default 1s, how often to fetch the device status

    Filter:                                                # optional, default include all, defines which registers are show in the view,
      # The rules are applied in order beginning with IncludeRegisters (highest priority) and ending with DefaultInclude (lowest priority).
      IncludeRegisters:                                    # optional, default empty, if a register is on this list, it is returned
      SkipRegisters:                                       # optional, default empty, if a register is on this list, it is not returned
      IncludeCategories:                                   # optional, default empty, all registers of the given category that are not explicitly skipped are returned
      SkipCategories:                                      # optional, default empty, all registers of the given category that are not explicitly included are not returned
      DefaultInclude: True                                 # optional, default true,  whether to return the registers that do not match any include/skip rule
    RestartInterval: 200ms                                 # optional, default 200ms, how fast to restart the device if it fails / disconnects
    RestartIntervalMaxBackoff: 1m                          # optional, default 1m; when it fails, the restart interval is exponentially increased up to this maximum
    LogDebug: false                                        # optional, default false, enable debug log output
    LogComDebug: false                                     # optional, default false, enable a verbose log of the communication with the device

  modbus-finder:                                           # mandatory, an arbitrary name used for logging and for referencing in other config sections
    Bus: bus0                                              # mandatory, the identifier of the modbus to use
    Kind: Finder7M38                                       # mandatory, type/model of the device; possibilities: WaveshareRtuRelay8, Finder7M38
    Address: 33                                            # mandatory, the modbus address of the device in hex as a string, e.g. 0x0A

HttpDevices:                                               # optional, a list of devices controlled via http
  tcw241:                                                  # mandatory, an arbitrary name used for logging and for referencing in other config sections
    Url: http://control0/                                  # mandatory, URL to the device; supported protocol is http/https; e.g. http://device0.local/
    Kind: Teracom                                          # mandatory, type/model of the device; possibilities: Teracom, Shelly3m
    Username: admin                                        # optional, username used to log in
    Password: my-secret                                    # optional, password used to log in
    PollInterval: 1s                                       # optional, default 1s, how often to fetch the device status
    Filter:                                                # optional, default include all, defines which registers are show in the view,
      # The rules are applied in order beginning with IncludeRegisters (highest priority) and ending with DefaultInclude (lowest priority).
      IncludeRegisters:                                    # optional, default empty, if a register is on this list, it is returned
      SkipRegisters:                                       # optional, default empty, if a register is on this list, it is not returned
      IncludeCategories:                                   # optional, default empty, all registers of the given category that are not explicitly skipped are returned
      SkipCategories:                                      # optional, default empty, all registers of the given category that are not explicitly included are not returned
      DefaultInclude: True                                 # optional, default true,  whether to return the registers that do not match any include/skip rule
    RestartInterval: 200ms                                 # optional, default 200ms, how fast to restart the device if it fails / disconnects
    RestartIntervalMaxBackoff: 1m                          # optional, default 1m; when it fails, the restart interval is exponentially increased up to this maximum
    LogDebug: false                                        # optional, default false, enable debug log output
    LogComDebug: false                                     # optional, default false, enable a verbose log of the communication with the device

MqttDevices:                                               # optional, a list of devices receiving its values via a mqtt server from another instance
  bmv1:                                                    # mandatory, an arbitrary name used for logging and for referencing in other config sections
    Kind: GoIotdeviceV3                                    # mandatory, only GoIotdevice is supported at the moment
    Filter:                                                # optional, default include all, defines which registers are show in the view,
      # The rules are applied in order beginning with IncludeRegisters (highest priority) and ending with DefaultInclude (lowest priority).
      IncludeRegisters:                                    # optional, default empty, if a register is on this list, it is returned
      SkipRegisters:                                       # optional, default empty, if a register is on this list, it is not returned
      IncludeCategories:                                   # optional, default empty, all registers of the given category that are not explicitly skipped are returned
      SkipCategories:                                      # optional, default empty, all registers of the given category that are not explicitly included are not returned
      DefaultInclude: True                                 # optional, default true,  whether to return the registers that do not match any include/skip rule
    RestartInterval: 200ms                                 # optional, default 200ms, how fast to restart the device if it fails / disconnects
    RestartIntervalMaxBackoff: 1m                          # optional, default 1m; when it fails, the restart interval is exponentially increased up to this maximum
    LogDebug: false                                        # optional, default false, enable debug log output
    LogComDebug: false                                     # optional, default false, enable a verbose log of the communication with the device

GensetDevices:                                             # optional, a list generator set control devices
  genset0:                                                 # mandatory, an arbitrary name used for logging and for referencing in other config sections
    Filter:                                                # optional, default include all, defines which registers are show in the view,
      # The rules are applied in order beginning with IncludeRegisters (highest priority) and ending with DefaultInclude (lowest priority).
      IncludeRegisters:                                    # optional, default empty, if a register is on this list, it is returned
      SkipRegisters:                                       # optional, default empty, if a register is on this list, it is not returned
      IncludeCategories:                                   # optional, default empty, all registers of the given category that are not explicitly skipped are returned
      SkipCategories:                                      # optional, default empty, all registers of the given category that are not explicitly included are not returned
      DefaultInclude: True                                 # optional, default true,  whether to return the registers that do not match any include/skip rule
    RestartInterval: 200ms                                 # optional, default 200ms, how fast to restart the device if it fails / disconnects
    RestartIntervalMaxBackoff: 1m                          # optional, default 1m; when it fails, the restart interval is exponentially increased up to this maximum
    LogDebug: false                                        # optional, default false, enable debug log output
    LogComDebug: false                                     # optional, default false, enable a verbose log of the communication with the device

    InputBindings:                                         # mandatory, a list of input bindings
      tcw241:                                              # the device name of the input device
        Available: IOAvailable                             # key: register name of input device; value: the target value of the genset controller
        DI0: ArmSwitch
        DI1: ResetSwitch
        DI2: FireDetected

      modbus-finder:                                       # the device name of a second input device providing data
        Available: OutputAvailable
        U1: U1
        U2: U2
        U3: U3
        P1: P1
        P2: P2
        P3: P3
        F: F

    OutputBindings:                                        # mandatory, a list of output bindings
      modbus-rtu0:                                         # the device name of the output device
        CH0: Ignition                                      # key: register name of output device; value: the source value of the genset controller
        CH1: Starter
        CH2: Fan
        CH3: Pump
        CH4: Load

    PrimingTimeout: 10s                                    # optional, default 10s, time in priming (only fuel pump on) state
    CrankingTimeout: 10s                                   # optional, default 10s, maximum time in cranking state
    WarmUpTimeout: 10m                                     # optional, default 10m, maximum time in warm-up state
    WarmUpMinTime: 2m                                      # optional, default 2m, minimum time in warm-up state
    WarmUpTemp: 50                                         # optional, default 50, minimum temperature to transition from warm-up to producing state
    EngineCoolDownTimeout: 5m                              # optional, default 5m, maximum time in engine cool-down state
    EngineCoolDownMinTime: 2m                              # optional, default 2m, minimum time in engine cool-down state
    EngineCoolDownTemp: 70                                 # optional, default 70, maximum temperature to transition from engine cool-down to enclosure cool-down state
    EnclosureCoolDownTimeout: 10m                          # optional, default 10m, maximum time in enclosure cool-down state
    EnclosureCoolDownMinTime: 2m                           # optional, default 2m, minimum time in enclosure cool-down state
    EnclosureCoolDownTemp: 30                              # optional, default 30, maximum temperature to transition from enclosure cool-down to ready state

    EngineTempMin: -20                                     # optional, default -10, minimum temperature the engine must have to not trigger the error state
    EngineTempMax: 90                                      # optional, default 90, maximum temperature the engine must have to not trigger the error state
    AuxTemp0Min: -20                                       # optional, default -20, minimum temperature the aux temperature sensor 0 must have to not trigger the error state
    AuxTemp0Max: 120                                       # optional, default 120, maximum temperature the aux temperature sensor 0 must have to not trigger the error state
    AuxTemp1Min: -20                                       # optional, default -20, minimum temperature the aux temperature sensor 1 must have to not trigger the error state
    AuxTemp1Max: 120                                       # optional, default 120, maximum temperature the aux temperature sensor 1 must have to not trigger the error state

    SinglePhase: false                                     # optional, default false, whether the generator is single phase or a three-phase system
    UMin: 200                                              # optional, default 200, minimum voltage the generator must have to not trigger the error state
    UMax: 250                                              # optional, default 260, maximum voltage the generator must have to not trigger the error state
    FMin: 45                                               # optional, default 45, minimum frequency the generator must have to not trigger the error state
    FMax: 55                                               # optional, default 55, maximum frequency the generator must have to not trigger the error state
    PMax: 1E6                                              # optional, default 1E6, maximum power the generator must have to not trigger the error state
    PTotMax: 1E6                                           # optional, default 1E6, maximum total power the generator must have to not trigger the error state

Views:                                                     # optional, a list of views (=categories in the frontend / paths in the api URLs)
  - Name: victron                                          # mandatory, a technical name used in the URLs
    Title: Victron                                         # mandatory, a nice title displayed in the frontend
    Devices:                                               # mandatory, a list of devices using
      - Name: bmv0                                         # mandatory, the arbitrary names defined above
        Title: Battery Monitor                             # mandatory, a nice title displayed in the frontend
        Filter:                                            # optional, default include all, defines which registers are show in the view,
                                                           # The rules are applied in order beginning with IncludeRegisters (highest priority) and ending with DefaultInclude (lowest priority).
          IncludeRegisters:                                # optional, default empty, if a register is on this list, it is returned
          SkipRegisters:                                   # optional, default empty, if a register is on this list, it is not returned
          IncludeCategories:                               # optional, default empty, all registers of the given category that are not explicitly skipped are returned
          SkipCategories:                                  # optional, default empty, all registers of the given category that are not explicitly included are not returned
          DefaultInclude: True                             # optional, default true,  whether to return the registers that do not match any include/skip rule
      - Name: modbus-rtu0                                  # mandatory, the arbitrary names defined above
        Title: Relay Board                                 # mandatory, a nice title displayed in the frontend
    Autoplay: true                                         # optional, default true, when true, live updates are enabled automatically when the view is open in the frontend
    AllowedUsers:                                          # optional, if empty, all users of the HtaccessFile are considered valid, otherwise only those listed here
      - test0                                              # username which is allowed to access this view
    Hidden: false                                          # optional, default false, if true, this view is not shown in the menu unless the user is logged in
```
