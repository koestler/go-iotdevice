# Changelog

## 3.9.1
* httpServer: refactor middlewares
* httpServer: fix gzip middleware: do not gzip when in frontend proxy mode
* httpServer: print size in log in human-readable format

## 3.9.0
* httpServer:
  * remove dependency on gin and use http.ServeMux instead.
    This reduced the binary size from 34M to 14M. 
  * add extensive tests

## 3.8.0
* genset: add stabilizing state between cranking and warm up.
  This allows the frequency / voltage to go out of the defined tolerance for a short time after cranking.
* Genset: add ErrTrigger state variable

## 3.7.2
* bump go-victron to v0.2.1:
  veregister: filter unavailable ProductRevision and Description registers for BMV700,BMV702,BMV700H

## 3.7.1
* bugfix: finder7M38: do not publish register list when device is unavailable

## 3.7.0
* refactor http server: do not depend on devicePool directly, use RegisterDbOfDeviceFunc instead
* genset:
  * respect output bindings
  * switch the configuration structure of bindings
* upgrade from golang-jwt/jwt to golang-jwt/jwt/v5
* refactor websocket server
  * remove unused "op" in output Message
  * simplify send timer and send messages (if available) in fixed intervals for 250ms
  * add a send timeout of 5s for websocket messages
  * really make sure that the ws sender never stalls the state storage
* tests: remove dependency on google/go-cmp

## 3.6.1
* genset
  * Change register names L[0-2] -> P[1-3] (to align with Finder7M38)

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
