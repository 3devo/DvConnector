FeConnector - Backend for the Next 1.0 logging application
======================
 > Version: Proof of concept

## Introduction

This application is written in golang and based on the [serial-json-server](https://github.com/chilipeppr/serial-port-json-server) implementation. Most of it has been stripped away but the core connection mechanics remained mostly the same.

This application when started starts a webserver that's listening on a random open port unless specified differently. It tries to serve the html files from the public folder unless specified differently.
The main purpose of the application is connecting to the connected Next 1.0 device and reading the incoming serial data from it parsing that and sending it to the connected websocket connections.

To use it as a fully fledged application you need to download a release from the [Fefrontend repository](https://github.com/3devo/fefrontend). And place the extracted content into the public folder.

## Usage

To use the feconnector executable there are is the standard way of just executing it but if you want more fine grain control of the port or folder it selects you can use flags to define those variables.

```
  -port string
        http service address. example 8800 to run on port 8800
        address (default ":8989")

  -allowexec
        Allow terminal commands to be executed (default false)

  -bufflowdebug string
        off = (default) We do not send back any debug JSON, on = We will send back a
        JSON response with debug info based on the configuration of the buffer flow
        that the user picked (default "off")

  -createstartupscript
        Create an /etc/init.d/serial-port-json-server startup script. Available only
        on Linux.

  -d string
        the directory of static file to host (default "./public")

  -disablecayenn
        Disable loading of Cayenn TCP/UDP server on port 8988

  -gc string
        Type of garbage collection. std = Normal garbage collection allowing system to
        decide (this has been known to cause a stop the world in the middle of a CNC
        job which can cause lost responses from the CNC controller and thus stalled
        jobs. use max instead to solve.), off = let memory grow unbounded (you have to
        send in the gc command manually to garbage collect or you will run out of RAM
        eventually), max = Force garbage collection on each recv or send on a serial
        port (this minimizes stop the world events and thus lost serial responses, but
        increases CPU usage) (default "std")

  -hibernate
        start hibernated

  -hostname string
        Override the hostname we get from the OS (default "unknown-hostname")

  -ls
        Launch self 5 seconds later. This flag is used when you ask for a restart from
        a websocket client.

  -regex string
        Regular expression to filter serial port list, i.e. -regex usb|acm

  -saddr string
        https service address. example :8801 to run https on port 8801 (default ":8990")

  -scert string
        https certificate file (default "cert.pem")

  -skey string
        https key file (default "key.pem")

  -v    show debug logging

```

## Requirements

* GO v10.0
* go dep
* Go path configured

## Build Setup

External dependencies

* github.com/3devo/feconnector/icon
* github.com/bob-thomas/go-serial
* github.com/facchinm/go-serial
* github.com/getlantern/systray
* github.com/go-ole/go-ole
* github.com/go-ole/go-ole/oleutil
* github.com/gorilla/websocket
* github.com/julienschmidt/httprouter
* github.com/kardianos/osext
* github.com/rs/cors
* github.com/skratchdot/open-golang/open
* github.com/tidwall/gjson
* github.com/gobuffalo/packr2/v2
* github.com/bob-thomas/configdir

To build the golang application you can run `go build` or use the build.sh utility.

If you want to build and deploy you will need to use and install [packr2](https://github.com/gobuffalo/packr/tree/master/v2) and build with `packr2 build` this will package the assets into executable.

The build utility can build the application and also generate + upload releases

```
    usage: ./build.sh [-c || -f || -r 0.1 \"cool release\" ||  -h
      -c             | --connector           : Build connector
      -r tag message | --release tag message : Build and upload release with tag and message
      -h             | --help                : This help message
```
