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

HttpServer:                                                # optional, when missing: http server is not started
  Bind: "::1"                                              # optional, default ::1 (ipv6 loopback), what address to bind to, use "0:0:0:0" when started within docker
  Port: 8000                                               # optional, default 8000
  LogRequests: True                                        # optional, default true, enables the http access log to stdout
  # configure FrontendProxy xor FrontendPath
  #FrontendProxy: "http://127.0.0.1:3000/"                 # optional: default deactivated; proxies the frontend to another server; useful for development
  FrontendPath: ./frontend-build/                          # optional: default "frontend-build": path to a static frontend build
  FrontendExpires: 5m                                      # optional: default 5min, what cache-control header to send for static frontend files
  ConfigExpires: 1m                                        # optional: default 1min, what cache-control header to send for configuration endpoints
  LogDebug: False                                          # optional: default false, output debug messages related to the http server

Authentication:                                            # optional, when missing: login is disabled
  #JwtSecret: "insert a random string here and uncomment"  # optional, default random, used to sign the JWT tokens
  JwtValidityPeriod: 1h                                    # optional, default 1h, users are logged out after this time
  HtaccessFile: ./auth.passwd                              # mandatory, where the file generated by htpasswd can be found

MqttClients:                                               # optional, when empty, no mqtt connection is made
  local:                                                   # mandatory: an arbitrary name used for logging and for referencing in other config sections
    Broker: tcp://mqtt.example.com:1883                    # mandatory: the URL to the server, use tcp:// or ssl://
    ProtocolVersion: 5                                     # optional, default 5, must be 5 always, only mqtt protocol version 5 is supported
    User: dev                                              # optional, default empty, the user used for authentication
    Password: zee4AhRi                                     # optional, default empty, the password used for authentication
    #ClientId: go-iotdevice                                # optional, default go-iotdevice-UUID, mqtt client id, make sure it is unique per mqtt-server
    Qos: 1                                                 # optional, default 1, what quality-of-service level shall be used for published messages and subscriptions
    KeepAlive: 60s                                         # optional, default 60s, how often a ping is sent to keep the connection alive
    ConnectionRetryDelay: 10s                              # optional, default 10s, when disconnected: after what delay shall a connection attempt is made
    ConnectTimeout: 5s                                     # optional, default 5s, how long to wait for the SYN+ACK packet, increase on slow networks
    AvailabilityTopic: %Prefix%tele/%ClientId%/status      # optional, what topic to use for online/offline messages
    TelemetryInterval: 10s                                 # optional, default 10s, how often to sent telemetry mqtt messages, 0s disables tlemetry messages
    TelemetryTopic: %Prefix%tele/go-iotdevice/%DeviceName%/state # optional, what topic to use for telemetry messages
    TelemetryRetain: false                                 # optional, default false, the mqtt retain flag for telemetry messages
    RealtimeEnable: false                                  # optional, default false, whether to enable sending realtime messages
    RealtimeTopic: %Prefix%stat/go-iotdevice/%DeviceName%/%ValueName% # optional, what topic to use for realtime messages
    RealtimeRetain: true                                   # optional, default true, the mqtt retain flag for realtime messages
    TopicPrefix:                                           # optional, default empty, %Prefix% is replaced with this string
    LogMessages: False                                     # optional, default false, log all incoming mqtt messages
    LogDebug: False                                        # optional, default false, very verbose debug log of the mqtt connection

VictronDevices:                                            # optional, a list of Victron Energy devices to connect to
  solar0:                                                  # mandatory: an arbitrary name used for logging and for referencing in other config sections
    Kind: Vedirect                                         # mandatory: possibilities: Vedirect, RandomBmv, RandomSolar, always set to Vedirect expect for development
    General:                                               # optional, this section is exactly the same for all devices
      SkipFields:                                          # optional, default empty, a list of field names that shall be ignored for this device
        - Temperature                                      # for BMV devices without a temperature sensor connect
        - AuxVoltage                                       # for BMV devices without a mid- or starter-voltage reading
      SkipCategories:                                      # optional, default empty, a list of category names that shall be ignored for this device
        - Settings                                         # for solar devices it might make sense to not fetch / output the settings
      TelemetryViaMqttClients:                             # optional, default all clients, to what mqtt servers shall telemetry messages be sent to
        - local                                            # state the arbitrary name of the mqtt client as defined in the MqttClients section of this file
      RealtimeViaMqttClients:                              # optional, default all clients, to what mqtt servers shall realtime messages be sent to
        - local
      LogDebug: False                                      # optional, default false, enable debug log output
      LogComDebug: False                                   # optional, default false, enable a verbose log of the communication with the device
    Device: /dev/serial/by-id/usb-VictronEnergy_BV_VE_Direct_cable_VEHTVQT-if00-port0 # mandatory, the path to the usb-to-serial converter

MqttDevices:
  dev-v1-main-bmv:
    MqttTopics:
      - piegn/stat/go-iotdevice/v1-main-bmv/+
  dev-v1-main-solar:
    MqttTopics:
      - piegn/stat/go-iotdevice/v1-main-solar/+
  dev-v1-aux-bmv:
    MqttTopics:
      - piegn/stat/go-iotdevice/v1-aux-bmv/+
  dev-v1-aux-solar:
    MqttTopics:
      - piegn/stat/go-iotdevice/v1-aux-solar/+



  #dev-random-solar:
  #  Kind: RandomSolar
  #  General:
  #    LogDebug: False
  #    LogComDebug: False
  #    SkipFields:
  #      - BatteryTemperature

HttpDevices:
  dev-tcw241:
    Url: http://control0/
    #Url: http://192.168.8.108/
    Kind: Teracom
    Username: admin
    Password: zooBohz7
    PollInterval: 500ms
    General:
      LogDebug: True

  dev-shelly-em3:
    Url: http://shelly-em3-0/
    Kind: ShellyEm3
    Username: admin
    Password: jair6aiK
    PollInterval: 500ms
    General:
      LogDebug: True

Views:
  - Name: victron
    Title: Victron
    Autoplay: true
    Devices:
      - Name: dev-v1-main-bmv
        Title: v1 main bmv
      - Name: dev-v1-main-solar
        Title: v1 main solar
      - Name: dev-v1-aux-bmv
        Title: v1 aux bmv
      - Name: dev-v1-aux-solar
        Title: v1 aux solar

  - Name: local
    Title: Local Dev
    Autoplay: true
    AllowedUsers:
      - lk
    Devices:
      - Name: dev-tcw241
        Title: TCW 241
      - Name: dev-shelly-em3
        Title: Shelly EM3
        #- Name: dev-solar
        #  Title: local solar
        #- Name: dev-random-solar
        #  Title: random solar


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

## Nginx reverse proxy
I normally run this service behind one or even two nginx server configured as a reverse proxy. It can take care of:
 * Serving multiple different applications on the same address using [SNI](https://en.wikipedia.org/wiki/Server_Name_Indication).
 * Caching on a fast cloud server in front of a device connected via a slow mobile connection.
 * https termination

### Setup
It's assumed that you understand each of the following steps in details. It's just to make the setup steps as quick as possible. 

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
