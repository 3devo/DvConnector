package main

import (
	"fmt"
	"log"

	"github.com/kardianos/osext"
	//"os"
	"os/exec"
	//"path"
	//"path/filepath"
	//"runtime"
	//"debug"
	"encoding/json"
	"runtime"
	"runtime/debug"
	"strings"
)

type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Inbound messages from the system
	broadcastSys chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

var h = hub{
	// buffered. go with 1000 cuz should never surpass that
	broadcast:    make(chan []byte, 1000),
	broadcastSys: make(chan []byte, 1000),
	// non-buffered
	//broadcast:    make(chan []byte),
	//broadcastSys: make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
			// send supported commands
			c.send <- []byte("{\"Version\" : \"" + version + "\"} ")
		case c := <-h.unregister:
			delete(h.connections, c)
			// put close in func cuz it was creating panics and want
			// to isolate
			func() {
				// this method can panic if websocket gets disconnected
				// from users browser and we see we need to unregister a couple
				// of times, i.e. perhaps from incoming data from serial triggering
				// an unregister. (NOT 100% sure why seeing c.send be closed twice here)
				defer func() {
					if e := recover(); e != nil {
						log.Println("Got panic: ", e)
					}
				}()
				close(c.send)
			}()
		case m := <-h.broadcast:
			//log.Print("Got a broadcast")
			//log.Print(m)
			//log.Print(len(m))
			if len(m) > 0 {
				//log.Print(string(m))
				//log.Print(h.broadcast)
				checkCmd(m)
				//log.Print("-----")

				for c := range h.connections {
					select {
					case c.send <- m:
						//log.Print("did broadcast to ")
						//log.Print(c.ws.RemoteAddr())
						//c.send <- []byte("hello world")
					default:
						delete(h.connections, c)
						close(c.send)
						go c.ws.Close()
					}
				}
			}
		case m := <-h.broadcastSys:
			//log.Printf("Got a system broadcast: %v\n", string(m))
			//log.Print(string(m))
			//log.Print("-----")

			for c := range h.connections {
				select {
				case c.send <- m:
					//log.Print("did broadcast to ")
					//log.Print(c.ws.RemoteAddr())
					//c.send <- []byte("hello world")
				default:
					delete(h.connections, c)
					close(c.send)
					go c.ws.Close()
				}
			}
		}
	}
}

func checkCmd(m []byte) {
	//log.Print("Inside checkCmd")
	s := string(m[:])
	log.Print(s)

	sl := strings.ToLower(s)

	if strings.HasPrefix(sl, "open") {
		// remove newline
		args := strings.Split(strings.TrimSpace(s), " ")
		if len(args) < 2 {
			go spErr("You did not specify a port in your open cmd")
			return
		}
		if len(args[1]) < 1 {
			go spErr("You did not specify a serial port")
			return
		}

		dtrOn := false
		if len(args) > 2 {
			dtrOn = true
		}

		go spHandlerOpen(args[1], 115200, false, dtrOn)

	} else if strings.HasPrefix(sl, "close") {

		log.Printf("About to split close commands. cmd:\"%v\"", s)
		// remove newline
		args := strings.Split(strings.TrimSpace(s), " ")
		//args := strings.Split(s, " ")
		log.Printf("The split args for close:%v", args)
		if len(args) > 1 {
			go spClose(args[1])
		} else {
			go spErr("You did not specify a port to close")
		}

	} else if strings.HasPrefix(sl, "sendjson") {
		// will catch sendjson

		go spWriteJson(s)

	} else if strings.HasPrefix(sl, "send") {
		// will catch send and sendnobuf

		//args := strings.Split(s, "send ")
		go spWrite(s)

	} else if strings.HasPrefix(sl, "list") {
		go spList()
		//go getListViaWmiPnpEntity()
	} else if strings.HasPrefix(sl, "restart") {
		restart()
	} else if strings.HasPrefix(sl, "exit") {
		exit()
	} else if strings.HasPrefix(sl, "memstats") {
		memoryStats()
	} else if strings.HasPrefix(sl, "gc") {
		garbageCollection()
	} else if strings.HasPrefix(sl, "version") {
		getVersion()
	} else {
		go spErr("Could not understand command.")
	}

	//log.Print("Done with checkCmd")
}

func memoryStats() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	json, _ := json.Marshal(memStats)
	log.Printf("memStats:%v\n", string(json))
	h.broadcastSys <- json
}

func getVersion() {
	h.broadcastSys <- []byte("{\"Version\" : \"" + version + "\"}")
}

func garbageCollection() {
	log.Printf("Starting garbageCollection()\n")
	h.broadcastSys <- []byte("{\"gc\":\"starting\"}")
	memoryStats()
	debug.SetGCPercent(100)
	debug.FreeOSMemory()
	debug.SetGCPercent(-1)
	log.Printf("Done with garbageCollection()\n")
	h.broadcastSys <- []byte("{\"gc\":\"done\"}")
	memoryStats()
}

func exit() {
	log.Println("Starting new spjs process")
	h.broadcastSys <- []byte("{\"Exiting\" : true}")
	log.Fatal("Exited current spjs cuz asked to")

}

func restart() {
	// relaunch ourself and exit
	// the relaunch works because we pass a cmdline in
	// that has serial-port-json-server only initialize 5 seconds later
	// which gives us time to exit and unbind from serial ports and TCP/IP
	// sockets like :8989
	log.Println("Starting new spjs process")
	h.broadcastSys <- []byte("{\"Restarting\" : true}")

	// figure out current path of executable so we know how to restart
	// this process
	/*
		dir, err2 := filepath.Abs(filepath.Dir(os.Args[0]))
		if err2 != nil {
			//log.Fatal(err2)
			fmt.Printf("Error getting executable file path. err: %v\n", err2)
		}
		fmt.Printf("The path to this exe is: %v\n", dir)

		// alternate approach
		_, filename, _, _ := runtime.Caller(1)
		f, _ := os.Open(path.Join(path.Dir(filename), "serial-port-json-server"))
		fmt.Println(f)
	*/

	// using osext
	exePath, err3 := osext.Executable()
	if err3 != nil {
		fmt.Printf("Error getting exe path using osext lib. err: %v\n", err3)
	}
	fmt.Printf("exePath using osext: %v\n", exePath)

	// figure out garbageCollection flag
	//isGcFlag := "false"

	var cmd *exec.Cmd
	/*if *isGC {
		//isGcFlag = "true"
		cmd = exec.Command(exePath, "-ls", "-addr", *addr, "-regex", *regExpFilter, "-gc")
	} else {
		cmd = exec.Command(exePath, "-ls", "-addr", *addr, "-regex", *regExpFilter)

	}*/
	cmd = exec.Command(exePath, "-ls", "-port", *port, "-regex", *regExpFilter, "-gc", *gcType)

	//cmd := exec.Command("./serial-port-json-server", "ls")
	err := cmd.Start()
	if err != nil {
		log.Printf("Got err restarting spjs: %v\n", err)
		h.broadcastSys <- []byte("{\"Error\" : \"" + fmt.Sprintf("%v", err) + "\"}")
	} else {
		h.broadcastSys <- []byte("{\"Restarted\" : true}")
	}
	log.Fatal("Exited current spjs for restart")
	//log.Printf("Waiting for command to finish...")
	//err = cmd.Wait()
	//log.Printf("Command finished with error: %v", err)
}

type CmdBroadcast struct {
	Cmd string
	Msg string
}

func (h *hub) sendErr(msg string) {
	msgMap := map[string]string{"error": msg}
	log.Println("Error: " + msg)
	bytes, err := json.Marshal(msgMap)
	if err != nil {
		log.Println("Failed to marshal data!")
		return
	}
	h.broadcastSys <- bytes
}

func (h *hub) sendMsg(name string, msg interface{}) {
	msgMap := make(map[string]interface{})
	msgMap[name] = msg
	msgMap["Type"] = name
	//log.Println("Sent: " + name)
	bytes, err := json.Marshal(msgMap)
	if err != nil {
		log.Println("Failed to marshal data!")
		return
	}
	h.broadcastSys <- bytes
}
