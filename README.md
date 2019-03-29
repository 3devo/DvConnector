DvConnector
===========
DvConnector is part of the [DevoVision](https://3devo.com/devovision/)
software used to monitor the Filament Makers built by 3devo. DevoVision
consists of two parts:
 - DvConnector, which runs on a computer, can connect to the Filament
   Maker serial port and has a small webserver to server the DvFrontend
   code.
 - DvFrontend, which is a javascript-based in-browser application that
   can talk to DvConnector to get info from the machine and show that
   info graphically.

This repository contains the golang-based DvConnector code. The
DvFrontend files (or any other frontend to be served by DvConnector)
should be put into the `frontend` directory.

This application is written in golang and based on the
[serial-json-server](https://github.com/chilipeppr/serial-port-json-server)
implementation. Most of it has been stripped away but the core
connection mechanics remained mostly the same.

## Usage

Usually, you can just start dvconnector. It will run a webserver on port
8989 and open up browser window to show the frontend (passing a
token to authenticate the connection). Additionally, it creates a system
tray icon, which can be used to control the application and open
additional browser windows if needed.

dvconnector accepts various commandline options:

```
  -port string
        port to start webserver on. example 8800 to run on port 8800
        address (default "8989")

  -gc string
        Type of garbage collection. std = Normal garbage collection allowing system to
        decide (this has been known to cause a stop the world in the middle of a CNC
        job which can cause lost responses from the CNC controller and thus stalled
        jobs. use max instead to solve.), off = let memory grow unbounded (you have to
        send in the gc command manually to garbage collect or you will run out of RAM
        eventually), max = Force garbage collection on each recv or send on a serial
        port (this minimizes stop the world events and thus lost serial responses, but
        increases CPU usage) (default "std")

  -ls
        Launch self 5 seconds later. This flag is used when you ask for a restart from
        a websocket client.

  -regex string
        Regular expression to filter serial port list, i.e. -regex usb|acm
        (note that there is also hardcoded filtering on usb vidpid)

  -v    show debug logging

  -b    Do not open a browser at startup

  -browserport port
        When opening the browser, use this this part rather than the
        port that the webserver listens on. This can be useful during
        debugging, e.g. when running a yarn dev server on a different
        port (proxying to the webserver port for DvConnector API
        requests).
```

## Building

Requirements:
 - Go 1.11 for the `go.mod` file to be understood properly.
 - The `goversioninfo` go package, for Windows builds with version info and an icon.
 - libgtk-3 and libappindicator3, for Linux builds with systray support.

To build the golang application you can run `go build`.

To build without systray support, add `-tags cli`.

If you want to build and deploy you will need to use and install
[packr2](https://github.com/gobuffalo/packr/tree/master/v2) and build
with `packr2 build` this will package all the assets into the
executable.

## License
DvConnector is based on the [Chilipeppr serial-port-json-server][spjs] project,
which is licensed under the GPL license. Hence the connector binary as
well as the sources, are licensed under the GPL as well. Note that
the original project does not clearly specify the GPL versions that can
be used, but it includes version 2 of the GPL, so DvConnector is
licensed under GPL v2-only to be sure.

[spjs]: https://github.com/chilipeppr/serial-port-json-server

In particular, the following terms apply to the DvConnector sources:

> Copyright (C) John Lauer and various contributors
> Copyright (C) 2018-2019 3devo B.V. (https://www.3devo.com)
>
> This program is free software; you can redistribute it and/or modify
> it under the terms of the GNU General Public License as published by
> the Free Software Foundation; version 2
>
> This program is distributed in the hope that it will be useful,
> but WITHOUT ANY WARRANTY; without even the implied warranty of
> MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
> GNU General Public License for more details.
>
> You should have received a copy of the GNU General Public License along
> with this program; if not, write to the Free Software Foundation, Inc.,
> 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

The full license text can be found in the LICENSE.md file.

Note that the DvFrontend source code is not currently available under an open
license.
