# go-iotdevice
[![Docker Image CI](https://github.com/koestler/js-iotsensor/actions/workflows/docker-image.yml/badge.svg?branch=main)](https://github.com/koestler/js-iotsensor/actions/workflows/docker-image.yml)

This tool reads values from various IoT devices like solar charges directly connected by USB
or relay boards connected by ethernet and publishes those on [MQTT](http://mqtt.org/) servers.

Additionally, a REST- and websocket-API and a [web frontend](https://github.com/koestler/js-iotdevice)
is made available via an internal http server.

![example frontend view](documentation/frontend.png)

The tool does not store any historical data. Instead,
[go-mqtt-to-influx](https://github.com/koestler/go-mqtt-to-influx) is used to write this data
to an [Influx Database](https://github.com/influxdata/influxdb) from where it is displayed
on a [Grafana](https://grafana.com/) dashboard.

The tool was written with the following two scenarios in mind:
* An off-grid holiday home installation running two batteries with
  Victron Energy [SmartSolar](https://www.victronenergy.com/solar-charge-controllers/bluesolar-mppt-150-35) / 
  [SmartShunt](https://www.victronenergy.com/battery-monitors/smart-battery-shunt) for solar and battery monitoring, 
 a [Shelly 3EM](https://www.shelly.cloud/en-ch/products/product-overview/shelly-3-em) for generator power monitoring.
 The tool runs on a single [Raspberry Pi Zero 2 W](https://www.raspberrypi.com/products/raspberry-pi-zero-2-w/).
* Remote control of a generator set using a [Teracom TCW241](https://www.teracomsystems.com/ethernet/ethernet-io-module-tcw241/)
  for start / stop and temperature monitoring. Control is integrated into [Homea Assistant](https://www.home-assistant.io/)
  via MQTT.

## Supported protocols and devices

The tool currently implements the following devices which are all used in an active project of mine.
However, it is made to be extended. Feel free to send pull requests or 
[create an issue](https://github.com/koestler/go-iotdevice/issues).

The following protocols are supported:
* [Victron Energy](https://www.victronenergy.com/) [VE.Direct](https://www.victronenergy.com/live/vedirect_protocol:faq)
* HTTP GET json / xml files
* MQTT

The following devices are supported:
* via a VE.Direct:
  * Victron Energy [BlueSolar MPPT](https://www.victronenergy.com/solar-charge-controllers/mppt7510)
  * Victron Energy [SmartSolar MPPT](https://www.victronenergy.com/solar-charge-controllers/smartsolar-150-35)
  * Victron Energy Battery Monitor [BMV 700](https://www.victronenergy.com/battery-monitors/bmv-700),
      [BMV 702](https://www.victronenergy.com/battery-monitors/bmv-702),
      [BMV-712 Smart](https://www.victronenergy.com/battery-monitors/bmv-712-smart)
  * Victron Energy [SmartShunt](https://www.victronenergy.com/battery-monitors/smart-battery-shunt)
* via HTTP:
  * [Shelly 3EM](https://www.shelly.cloud/en-ch/products/product-overview/shelly-3-em) 3 phase energy power monitor
  * [Teracom TCW241](https://www.teracomsystems.com/ethernet/ethernet-io-module-tcw241/) industrial relay / sensor board
* via MQTT:
  * Another go-iotdevice instance connected to the same MQTT broker. This allows to connect devices
    to different linux machines at different location but still having one single frontend showing all devices.

## Configuration
The tool supports command line options to for profiling. See:
```bash
./go-iotdevice --help
Usage:
  go-iotdevice [-c <path to yaml config file>]

Application Options:
      --version     Print the build version and timestamp
  -c, --config=     Config File in yaml format (default: ./config.yaml)
      --cpuprofile= write cpu profile to <file>
      --memprofile= write memory profile to <file>

Help Options:
  -h, --help        Show this help message
```

All other configuration defined in a single yaml file which is read from `./config.yaml` by default.

Here is a fully documented configuration file including all options:
```yaml
# documentation/full-config.yaml

Version: 1                                                 # configuration file format; must be set to 1 for >v2 of this tool.
ProjectTitle: Configurable Title of Project                # optional, default go-iotdevice: is shown in the http frontend
LogConfig: True                                            # optional, default False, outputs the used configuration including defaults on startup
LogWorkerStart: True                                       # optional, default False, outputs what devices and mqtt clients are started
LogDebug: False                                            # optional, default False, outputs various debug information

Authentication:                                            # optional, when missing: login is disabled
  #JwtSecret: "insert a random string here and uncomment"  # optional, default random, used to sign the JWT tokens
  JwtValidityPeriod: 1h                                    # optional, default 1h, users are logged out after this time
  HtaccessFile: ./auth.passwd                              # mandatory, where the file generated by htpasswd can be found

HttpServer:                                                # optional, when missing: http server is not started
  Bind: "::1"                                              # optional, default ::1 (ipv6 loopback), what address to bind to, use "0:0:0:0" when started within docker
  Port: 8000                                               # optional, default 8000
  LogRequests: True                                        # optional, default true, enables the http access log to stdout
  # configure FrontendProxy xor FrontendPath
  #FrontendProxy: "http://127.0.0.1:3000/"                 # optional: default deactivated; proxies the frontend to another server; useful for development
  FrontendPath: "./frontend-build/"                        # optional: default "frontend-build": path to a static frontend build
  FrontendExpires: "5min"                                  # optional: default 5min, what cache-control header to send for static frontend files
  ConfigExpires: ""

Devices:
  #0-solar:
  #  Kind: RandomSolar
  1-bmv:
    Kind: RandomBmv
  2-bmv:
    Kind: Vedirect
    Device: /dev/ttyUSB1
    LogDebug: True

Views:
  - Name: public
    Title: Simple
    Devices:
      #- Name: 0-solar
      #  Title: Solar Charger
      - Name: 1-bmv
        Title: Battery One
      - Name: 2-bmv
        Title: Battery Two

  - Name: private
    Title: Details
    Devices:
      - Name: 2-bmv
        Title: Battery Two
    AllowedUsers:
      - tester

```

## Authentication
The tool can use [JWT](https://jwt.io/) to make certain views only available after a login. The user database
is stored in an apache htaccess file which can be changed without restarting the server. 

### Configuration

There are two relevant section in the configuration file:

1. Adding the `Authentication:` section enables the login / authentication mechanism.
The `JwtSecret` is used to sign the tokens. When left unconfigured, a new random secret is generated on each
startup of the backend. This results in all users being logged out after each restart of the server.
It's therefore best to hardcode a random secret.

2. Per `View` you can define a list of `AllowedUsers`. When the list is present and has at lest one entry, only
usernames on that list can access this view. If the list is empty, all users in the user database can access. 


### User database
The only supported authentication backend at the moment is a simple apache htaccess file. Set it up as follows:

```bash
htpasswd -c auth.passwd lorenz
# enter password twice
htpasswd auth.passwd daniela
# enter another password twice
```

## Development

### Run locally
For development this backend can be compiled and run locally.
In addition, it's then best to also und run the [frontend](https://github.com/koestler/js-iotdevice) locally. 
There is a good starting point for a development configuration.

```bash
cp documentation/dev-config.yml config.yml
go build && ./go-iotdevice
```


### Compile and run inside docker
Alternatively, if you don't have golang installed locally, you can compile and run 

```bash
docker build -f docker/Dockerfile -t go-iotdevice .
docker run --rm --name go-iotdevice -p 127.0.0.1:8000:8000 \
  -v "$(pwd)"/documentation/dev-config.yaml:/config.yaml:ro \
  go-iotdevice
```

### run tests
```bash
go install github.com/golang/mock/mockgen@v1.6.0
go genreate ./...
go test ./...
```

### Update README.md
```bash
npx embedme README.md
```
