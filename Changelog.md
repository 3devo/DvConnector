Version: Proof concept
=============

This is the first version of the FeConnector.

Keep in mind that this is a proof of concept and not everything will work as expected. The core of the application is tightly coupled and based on the [serial-port-json-server](https://github.com/chilipeppr/serial-port-json-server).

- Added a systray icon to give the user visibility that the application is running and able to close and open the webpage.
- Added Dtr configuration for the serial port you can now enable a connection without resetting it first
- Use a static file server to serve html instead of inline html
- A restful api to manage resources for the workspace cms
  - GET for Charts, Sheets, Workspaces, Logs
  - POST for Charts, Sheets, Workspaces
  - PUT for Charts, Sheets, Workspaces
  - DELETE for Charts, Sheets, Workspaces, Logs
- Custom 3Devo buffer algorithm te receive incoming serial data
- Removed all unneeded buffer algorithms
- Removed unneeded commands and configuration

Issues
----

- There is a sporadic change of serial corruption. The issue hasn't been pinpointed yet but might have to do something with the go-serial library that the project uses
- Exiting a serialport handler when closing can cause a lockup in the system.
- The restful api is extremely limited and connected to json files on the filesystem. This needs to be remade into using a database and per resource routing instead of a generic handler.