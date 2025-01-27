# Changelog

## 3.6.0
* go-victron: add solar load registers for 10/15/20A solar chargers.
* Bump dependencies

## 3.5.1
* Fix gpioDevice: handle case when no inputs or no outputs are configured
* Bump dependencies

## 3.5.0
* Add gpio device

## v3.4.0
* Add prototype of new generator set device.
* Fix a bug with VE.Direct solar chargers with Panel current 10A/15A/20A. Now they are working with default configuration.
  Before, they run into an error because the Panel current register does not exist.
* Bump various dependencies.