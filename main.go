// Version 1.95 - serial port json server
// Version POC - FeConnector
// Supports Windows, Linux, Mac, and Raspberry Pi, Beagle Bone Black

////go:generate swagger generate spec -m -o swagger.json
//go:generate goversioninfo

// Package classification FeConnector API.
//
// the purpose of this application is to provide an application
// that is using plain go code to define an API
//
// This should demonstrate all the possible comment annotations
// that are available to turn go code into a fully compliant swagger 2.0 spec
//
// Terms Of Service:
//
// there are no TOS at this moment, use at your own risk we take no responsibility
//
//     Schemes: http, https
//     Host: localhost:8989
//     BasePath: /api/v0.1.0
//     Version: 0.1.0
//     Contact: 3Devo<bobthomas@devo.com> http://3devo.com
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Security:
//     - Bearer:
//
//     SecurityDefinitions:
//     Bearer:
//          type: apiKey
//          name: Authorization
//          in: header
// swagger:meta
package main

import (
	"flag"
	"go/build"
	"log"
	"net/http"
	"path/filepath"

	"github.com/bob-thomas/configdir"
	packr "github.com/gobuffalo/packr/v2"

	//"path/filepath"
	"errors"
	"fmt"
	"net"
	"os"

	//"net/http/pprof"
	//"runtime"
	"io"
	"runtime/debug"
	"time"

	"github.com/3devo/feconnector/middleware"
	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/routing"
	"github.com/3devo/feconnector/utils"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/julienschmidt/httprouter"
	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/rs/cors"
	"github.com/skratchdot/open-golang/open"
	"github.com/urfave/negroni"
	validator "gopkg.in/go-playground/validator.v9"
)

var (
	version      = "0.1.0"
	versionFloat = float32(0.1)
	port         = flag.String("port", "8989", "Listening port.")
	browserport  = flag.String("browserport", "", "Port to open in the browser. Defaults to the listening port, only needs to be changed during development.")
	//	addr  = flag.String("addr", ":8980", "http service address. example :8800 to run on port 8800, example 10.0.0.2:9000 to run on specific IP address and port, example 10.0.0.2 to run on specific IP address")
	saddr     = flag.String("saddr", ":8990", "https service address. example :8801 to run https on port 8801")
	scert     = flag.String("scert", "cert.pem", "https certificate file")
	skey      = flag.String("skey", "key.pem", "https key file")
	hibernate = flag.Bool("hibernate", false, "start hibernated")
	noBrowser = flag.Bool("b", false, "Don't open the webpage")
	//assets       = flag.String("assets", defaultAssetPath(), "path to assets")
	//	verbose = flag.Bool("v", true, "show debug logging")
	verbose = flag.Bool("v", false, "show debug logging")
	//homeTempl *template.Template
	isLaunchSelf = flag.Bool("ls", false, "Launch self 5 seconds later. This flag is used when you ask for a restart from a websocket client.")

	// regular expression to sort the serial port list
	// typically this wouldn't be provided, but if the user wants to clean
	// up their list with a regexp so it's cleaner inside their end-user interface
	// such as ChiliPeppr, this can make the massive list that Linux gives back
	// to you be a bit more manageable
	regExpFilter = flag.String("regex", "", "Regular expression to filter serial port list, i.e. -regex usb|acm")

	// allow garbageCollection()
	//isGC = flag.Bool("gc", false, "Is garbage collection on? Off by default.")
	//isGC = flag.Bool("gc", true, "Is garbage collection on? Off by default.")
	gcType = flag.String("gc", "std", "Type of garbage collection. std = Normal garbage collection allowing system to decide (this has been known to cause a stop the world in the middle of a CNC job which can cause lost responses from the CNC controller and thus stalled jobs. use max instead to solve.), off = let memory grow unbounded (you have to send in the gc command manually to garbage collect or you will run out of RAM eventually), max = Force garbage collection on each recv or send on a serial port (this minimizes stop the world events and thus lost serial responses, but increases CPU usage)")

	// hostname. allow user to override, otherwise we look it up
	hostname = flag.String("hostname", "unknown-hostname", "Override the hostname we get from the OS")

	db        *storm.DB
	validate  = validator.New()
	env       *utils.Env
	ip        string
	dataDir   = configdir.DataDir("3devo", "FM-Monitor")     // Directory for user files like logs and notes
	configDir = configdir.SettingsDir("3devo", "FM-Monitor") // Directory for user configuration files
)

type NullWriter int

func (NullWriter) Write([]byte) (int, error) { return 0, nil }

func defaultAssetPath() string {
	//p, err := build.Default.Import("gary.burd.info/go-websocket-chat", "", build.FindOnly)
	p, err := build.Default.Import("github.com/johnlauer/serial-port-json-server", "", build.FindOnly)
	if err != nil {
		return "."
	}
	return p.Dir
}

func launchSelfLater() {
	log.Println("Going to launch myself 5 seconds later.")
	time.Sleep(2 * 1000 * time.Millisecond)
	log.Println("Done waiting 5 secs. Now launching...")
}

func launchBrowserWithToken() {
	port_to_use := *port
	if *browserport != "" {
		port_to_use = *browserport
	}
	token, _ := utils.GenerateJWTToken("browser", time.Now().Add(time.Minute*time.Duration(utils.StandardTokenExpiration)).Unix())
	open.Run("http://" + string(ip) + ":" + port_to_use + "/#/?token=" + token)
}

func main() {
	os.MkdirAll(dataDir, os.ModePerm)
	os.MkdirAll(configDir, os.ModePerm)
	setupSysTray(onInit)
}

func onInit() {
	fillSysTray()
	newDatabase := false
	dbPath := filepath.Join(dataDir, "database", "feconnector.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		newDatabase = true
	}
	webBox := packr.New("Frontend", "./frontend")
	os.MkdirAll(filepath.Join(dataDir, "logs"), os.ModePerm)
	os.MkdirAll(filepath.Join(dataDir, "notes"), os.ModePerm)
	os.MkdirAll(filepath.Join(dataDir, "database"), os.ModePerm)

	var err error
	db, err = storm.Open(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database %s: %s", dbPath, err)
	}
	db.Init(&models.User{})
	db.Init(&models.BlackListedToken{})
	db.Init(&models.Workspace{})
	db.Init(&models.Sheet{})
	db.Init(&models.Chart{})
	db.Init(&models.LogFile{})
	db.Init(&models.Config{})
	if newDatabase {
		log.Println("filling database with default values")
		FillDatabase(db)
	}
	var users []models.User
	var config models.Config
	db.One("ID", 1, &config)

	db.All(&users)
	if len(users) > 0 {
		config.UserCreated = true
	} else {

		config.OpenNetwork = false
		config.UserCreated = false
	}
	db.Save(&config)
	// Delete expired tokens
	db.Select(q.Lt("Expiration", time.Now().Unix())).Delete(new(models.BlackListedToken))

	defer db.Close()
	env = &utils.Env{Db: db, Validator: validate, DataDir: dataDir, ConfigDir: configDir}
	/** Custom validators **/
	validate.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		return utils.IsValidUUID(fl.Field().String())
	})

	validate.RegisterValidation("chart-exists", func(fl validator.FieldLevel) bool {
		var chart models.Chart
		err := db.One("UUID", fl.Field().String(), &chart)
		if err != nil {
			return false
		}
		return true
	})

	validate.RegisterValidation("sheet-exists", func(fl validator.FieldLevel) bool {
		var sheet models.Sheet
		err := db.One("UUID", fl.Field().String(), &sheet)
		if err != nil {
			return false
		}
		return true
	})

	// Test USB list
	//	GetUsbList()

	// parse all passed in command line arguments
	flag.Parse()

	// setup logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	if config.OpenNetwork {
		ip, _ = externalIP()
	} else {
		ip = "localhost"
	}

	// see if we are supposed to wait 5 seconds
	if *isLaunchSelf {
		launchSelfLater()
	}
	if !*noBrowser {
		launchBrowserWithToken()
	}

	//getList()
	log.Println("Version:" + version)

	// hostname
	hn, _ := os.Hostname()
	if *hostname == "unknown-hostname" {
		*hostname = hn
	}
	log.Println("Hostname:", *hostname)

	// turn off garbage collection
	// this is dangerous, as u could overflow memory
	//if *isGC {
	if *gcType == "std" {
		log.Println("Garbage collection is on using Standard mode, meaning we just let Golang determine when to garbage collect.")
	} else if *gcType == "max" {
		log.Println("Garbage collection is on for MAXIMUM real-time collecting on each send/recv from serial port. Higher CPU, but less stopping of the world to garbage collect since it is being done on a constant basis.")
	} else {
		log.Println("Garbage collection is off. Memory use will grow unbounded. You WILL RUN OUT OF RAM unless you send in the gc command to manually force garbage collection. Lower CPU, but progressive memory footprint.")
		debug.SetGCPercent(-1)
	}

	//homeTempl = template.Must(template.ParseFiles(filepath.Join(*assets, "home.html")))

	// see if they provided a regex filter
	if len(*regExpFilter) > 0 {
		log.Printf("You specified a serial port regular expression filter: %v\n", *regExpFilter)
	}

	//GetDarwinMeta()

	if !*verbose {
		log.Println("You can enter verbose mode to see all logging by starting with the -v command line switch.")
		//		log.SetOutput(new(NullWriter)) //route all logging to nullwriter
	}

	// list serial ports
	portList, _ := GetList()
	metaports, _ := GetMetaList()

	/*if errSys != nil {
		log.Printf("Got system error trying to retrieve serial port list. Err:%v\n", errSys)
		log.Fatal("Exiting")
	}*/

	// serial port list thread
	go func() {
		time.Sleep(1300 * time.Millisecond)
		log.SetOutput(io.Writer(os.Stdout))
		log.Println("Your serial ports:")
		if len(portList) == 0 {
			log.Println("\tThere are no serial ports to list.")
		}
		for _, element := range portList {
			// if we have meta data for this port, use it
			setMetaDataForOsSerialPort(&element, metaports)
			log.Printf("\t%v\n", element)

		}
		if !*verbose {
			//log.Println("You can enter verbose mode to see all logging by starting with the -v command line switch.")
			log.SetOutput(new(NullWriter)) //route all logging to nullwriter
		}
	}()

	// launch the hub routine which is the singleton for the websocket server
	go h.run()
	// launch our serial port routine
	go sh.run()
	// launch our dummy data routine
	//go d.run()

	router := httprouter.New()
	restURL := fmt.Sprintf("/api/v%v/", string(version))
	router.GET("/ws", wsHandle(env))

	/**	LOG FILE ROUTING */
	router.GET(restURL+"logFiles", middleware.AuthRequired(routing.GetAllLogFiles(env), env))
	router.GET(restURL+"logFiles/:uuid", middleware.AuthRequired(routing.GetLogFile(env), env))
	router.POST(restURL+"logFiles", middleware.AuthRequired(routing.CreateLogFile(env), env))
	router.DELETE(restURL+"logFiles/:uuid", middleware.AuthRequired(routing.DeleteLogFile(env), env))
	router.PUT(restURL+"logFiles/:uuid", middleware.AuthRequired(routing.UpdateLogFile(env), env))

	/**	CHART ROUTING */
	router.GET(restURL+"charts", middleware.AuthRequired(routing.GetAllCharts(env), env))
	router.GET(restURL+"charts/:uuid", middleware.AuthRequired(routing.GetChart(env), env))
	router.POST(restURL+"charts", middleware.AuthRequired(routing.CreateChart(env), env))
	router.DELETE(restURL+"charts/:uuid", middleware.AuthRequired(routing.DeleteChart(env), env))
	router.PUT(restURL+"charts/:uuid", middleware.AuthRequired(routing.UpdateChart(env), env))

	/**	SHEET ROUTING */
	router.GET(restURL+"sheets", middleware.AuthRequired(routing.GetAllSheets(env), env))
	router.GET(restURL+"sheets/:uuid", middleware.AuthRequired(routing.GetSheet(env), env))
	router.POST(restURL+"sheets", middleware.AuthRequired(routing.CreateSheet(env), env))
	router.DELETE(restURL+"sheets/:uuid", middleware.AuthRequired(routing.DeleteSheet(env), env))
	router.PUT(restURL+"sheets/:uuid", middleware.AuthRequired(routing.UpdateSheet(env), env))

	/**	WORKSPACE ROUTING */
	router.GET(restURL+"workspaces", middleware.AuthRequired(routing.GetAllWorkspaces(env), env))
	router.GET(restURL+"workspaces/:uuid", middleware.AuthRequired(routing.GetWorkspace(env), env))
	router.POST(restURL+"workspaces", middleware.AuthRequired(routing.CreateWorkspace(env), env))
	router.DELETE(restURL+"workspaces/:uuid", middleware.AuthRequired(routing.DeleteWorkspace(env), env))
	router.PUT(restURL+"workspaces/:uuid", middleware.AuthRequired(routing.UpdateWorkspace(env), env))

	/**	USER ROUTING */
	router.POST(restURL+"users", routing.CreateUser(env))
	router.DELETE(restURL+"users/:uuid", middleware.AuthRequired(routing.DeleteUser(env), env))
	router.PUT(restURL+"users/:uuid", middleware.AuthRequired(routing.UpdateUser(env), env))

	/**	CONFIG ROUTING */
	router.GET(restURL+"config", routing.GetConfig(env))
	router.PUT(restURL+"config", middleware.AuthRequired(routing.UpdateConfig(env), env))

	/**	AUTH ROUTING */
	router.POST(restURL+"refreshToken", middleware.AuthRequired(routing.RefreshToken(env), env))
	router.POST(restURL+"login", routing.Login(env))
	router.POST(restURL+"logout", middleware.AuthRequired(routing.Logout(env), env))

	router.NotFound = http.FileServer(webBox)
	f := flag.Lookup("port")
	log.Println("Starting http server and websocket on " + ip + ":" + f.Value.String())

	/** Hook in middlewares */
	negroniMiddleware := negroni.New()
	negroniMiddleware.Use(negronilogrus.NewMiddleware())
	negroniMiddleware.Use(cors.AllowAll())
	negroniMiddleware.UseHandler(router)

	go startHttp(ip, config, negroniMiddleware)
	ch := make(chan bool)
	<-ch
}

func startHttp(ip string, config models.Config, h http.Handler) {
	log.Println("Starting http server and websocket on " + ip + ":" + *port)
	serveInterfaces := ":" + *port
	if !config.OpenNetwork {
		serveInterfaces = ip + ":" + *port
	}
	if err := http.ListenAndServe(serveInterfaces, h); err != nil {
		fmt.Printf("Error trying to bind to http port: %v, so exiting...\n", err)
		log.Fatal("Error ListenAndServe:", err)
	}
}

func startHttps(ip string) {
	// generate self-signed cert for testing or local trusted networks
	// openssl req -x509 -nodes -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365

	f := flag.Lookup("saddr")
	cert, certErr := os.Open(*scert)
	key, keyErr := os.Open(*skey)

	cert.Close()
	key.Close()

	if certErr != nil || keyErr != nil {
		log.Println("Missing tls cert and/or key. Will not start HTTPS server.")
		//fmt.Println("Missing tls cert and/or key. Will not start HTTPS server.")
		return
	}

	log.Println("Starting https server and websocket on " + ip + "" + f.Value.String())
	if err := http.ListenAndServeTLS(*saddr, *scert, *skey, nil); err != nil {
		fmt.Printf("Error trying to bind to https port: %v, so exiting...\n", err)
		log.Fatal("Error ListenAndServeTLS:", err)
	}
}

func externalIP() (string, error) {
	//log.Println("Getting external IP")
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println("Got err getting external IP addr")
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			//log.Println("Iface down")
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			//log.Println("Loopback")
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			log.Println("Got err on iface.Addrs()")
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				//log.Println("Ip was nil or loopback")
				continue
			}
			ip = ip.To4()
			if ip == nil {
				//log.Println("Was not ipv4 addr")
				continue // not an ipv4 address
			}
			//log.Println("IP is ", ip.String())
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
