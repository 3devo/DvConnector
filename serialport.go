package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	//"github.com/johnlauer/goserial"
	"sync"

	//"github.com/facchinm/go-serial"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/3devo/feconnector/feProtocol"
	serial "github.com/bob-thomas/go-serial"
	. "github.com/logrusorgru/aurora"
	"github.com/sigurn/crc8"
)

type SerialConfig struct {
	Name string
	Baud int

	// Size     int // 0 get translated to 8
	// Parity   SomeNewTypeToGetCorrectDefaultOf_None
	// StopBits SomeNewTypeToGetCorrectDefaultOf_1

	// RTSFlowControl bool
	// DTRFlowControl bool
	// XONFlowControl bool

	// CRLFTranslate bool
	// TimeoutStuff int
	RtsOn bool
	DtrOn bool
}

// The serial port connection.
type serport struct {

	// Needed for original serial library
	// portConf *serial.Config
	// Needed for Arduino serial library
	portConf *SerialConfig

	portIo io.ReadWriteCloser

	serialPort serial.Port

	done chan bool // signals the end of this request

	// Keep track of whether we're being actively closed
	// just so we don't show scary error messages
	isClosing bool

	// counter incremented on queue, decremented on write
	itemsInBuffer int

	// buffered channel containing up to 25600 outbound messages.
	sendBuffered chan Cmd

	// unbuffered channel of outbound messages that bypass internal serial port buffer
	sendNoBuf chan Cmd

	// Do we have an extra channel/thread to watch our buffer?
	BufferType string
	//bufferwatcher *BufferflowDummypause
	bufferwatcher Bufferflow

	// Keep track of whether this is the primary serial port, i.e. cnc controller
	// or if its secondary, i.e. a backup port or arduino or something tertiary
	IsPrimary   bool
	IsSecondary bool

	// Feedrate override value
	feedRateOverride     float32
	isFeedRateOverrideOn bool
}

type Cmd struct {
	data                       string
	id                         string
	skippedBuffer              bool
	willHandleCompleteResponse bool
	pause                      int
}

type CmdComplete struct {
	Cmd     string
	Id      string
	P       string
	BufSize int    `json:"-"`
	D       string `json:"-"`
}

type qwReport struct {
	Cmd  string
	QCnt int
	Id   string
	D    string `json:"-"`
	Buf  string `json:"-"`
	P    string
}

type qwReportWithData struct {
	Cmd  string
	QCnt int
	Id   string
	D    string //`json:"-"`
	Buf  string `json:"-"`
	P    string
}

type SpPortMessage struct {
	P string // the port, i.e. com22
	D string // the data, i.e. G0 X0 Y0
}

func (p *serport) reader() {
	testReader := feProtocol.BufferedReader{PortIo: p.portIo}
	handler := feProtocol.FeProtocolHandler{
		Transport: testReader}

	for {
		if p.isClosing {
			strmsg := "Shutting down reader on " + p.portConf.Name
			log.Println(strmsg)
			h.broadcastSys <- []byte(strmsg)
			break
		}
		buffer, err, read := testReader.ReadBytes(1)
		frame := buffer
		if !err && read > 0 {
			switch frame[0] {
			case feProtocol.START_FRAME, feProtocol.START_FRAME ^ feProtocol.SEQUENCE_MASK:
				sequence := 0
				if frame[0] != 0xaa {
					sequence = 1
				}
				command, _, _ := testReader.ReadBytes(1)
				payloadLength, _, _ := testReader.ReadBytes(2)
				payload, err := handler.ReadInPayload(binary.LittleEndian.Uint16(payloadLength))
				if err {
					handler.OnFrameError(feProtocol.PAYLOAD_CORRUPTION)
					break
				}
				crc, _, _ := testReader.ReadBytes(1)
				end, _, _ := testReader.ReadBytes(1)
				var crc_sum []byte
				crc_sum = append(crc_sum, frame...)
				crc_sum = append(crc_sum, command...)
				crc_sum = append(crc_sum, payloadLength...)
				crc_sum = append(crc_sum, payload...)
				crc_sum = append(crc_sum, end...)
				calc_crc := crc8.Checksum(crc_sum, crc8Table)

				crc_check := crc[0] == calc_crc
				if end[0] != feProtocol.END_FRAME {
					handler.OnFrameError(feProtocol.PAYLOAD_CORRUPTION)
					break
				}
				if !crc_check {
					handler.OnFrameError(feProtocol.CRC)
					break
				} else {
					frame := feProtocol.Frame{
						Sequence: uint8(sequence),
						Command:  uint8(command[0]),
						Length:   binary.LittleEndian.Uint16(payloadLength),
						Payload:  payload}
					handler.OnFrameReceived(&frame)
				}
			case feProtocol.END_FRAME:
				log.Print(Red("Failed frame - no start frame"))
			}

		}
	}
	p.portIo.Close()
}

// this method runs as its own thread because it's instantiated
// as a "go" method. so if it blocks inside, it is ok
func (p *serport) writerBuffered() {

	// this method can panic if user closes serial port and something is
	// in BlockUntilReady() and then a send occurs on p.sendNoBuf

	defer func() {
		if e := recover(); e != nil {
			// e is the interface{} typed-value we passed to panic()
			log.Println("Got panic: ", e) // Prints "Whoops: boom!"
		}
	}()

	// this for loop blocks on p.sendBuffered until that channel
	// sees something come in
	for data := range p.sendBuffered {

		log.Printf("Got p.sendBuffered. data:%v, id:%v, pause:%v\n", strings.Replace(string(data.data), "\n", "\\n", -1), string(data.id), data.pause)

		// we want to block here if we are being asked
		// to pause.
		goodToGo, willHandleCompleteResponse, newGcode := p.bufferwatcher.BlockUntilReady(string(data.data), data.id)

		// BlockUntilReady can modify our Gcode now so it can possibly add tracking data
		// so if we got newGcode then we must swap it for our original gcode
		if len(newGcode) > 0 {
			data.data = newGcode
		}

		if goodToGo == false {
			log.Println("We got back from BlockUntilReady() but apparently we must cancel this cmd")
			// since we won't get a buffer decrement in p.sendNoBuf, we must do it here
			p.itemsInBuffer--
		} else {
			// send to the non-buffered serial port writer
			//log.Printf("About to send to p.sendNoBuf channel. cmd:%v", data)
			data.willHandleCompleteResponse = willHandleCompleteResponse
			p.sendNoBuf <- data
		}
	}
	msgstr := "writerBuffered just got closed. make sure you make a new one. port:" + p.portConf.Name
	log.Println(msgstr)
	h.broadcastSys <- []byte(msgstr)
}

// this method runs as its own thread because it's instantiated
// as a "go" method. so if it blocks inside, it is ok
func (p *serport) writerNoBuf() {
	// this for loop blocks on p.send until that channel
	// sees something come in
	for data := range p.sendNoBuf {

		log.Printf("Got p.sendNoBuf. id:%v, pause:%v, data:%v\n", string(data.id), data.pause, strings.Replace(string(data.data), "\n", "\\n", -1))

		// if we get here, we were able to write successfully
		// to the serial port because it blocks until it can write

		// decrement counter
		p.itemsInBuffer--
		log.Printf("Items In SPJS Queue List:%v\n", p.itemsInBuffer)
		//h.broadcastSys <- []byte("{\"Cmd\":\"Write\",\"QCnt\":" + strconv.Itoa(p.itemsInBuffer) + ",\"Byte\":" + strconv.Itoa(n2) + ",\"Port\":\"" + p.portConf.Name + "\"}")

		// Figure out buffered or not buffered
		buf := "Buf"
		if data.skippedBuffer {
			buf = "NoBuf"
		}

		// WARNING - Feedrate Override doesn't really belong in here because this is supposed
		// to be a generic implementation of sending/receiving to serial ports
		// However, there's not really a better place to put this because you need to know
		// last minute what the feedrate override is and let the user adjust it at any time
		// If you want a generic serial port implementation, remove this last minute call from this code

		didWeOverride := false
		newData := ""
		if p.isFeedRateOverrideOn {
			didWeOverride, newData = doFeedRateOverride(data.data, p.feedRateOverride)
		}

		if didWeOverride {
			// We need to reset the gcode and make the qwReport be what we want
			// Since we changed the gcode, we need to report it back to the user
			// For reducing load on websocket, stop transmitting write data
			data.data = newData
			qwr := qwReportWithData{
				Cmd:  "Write",
				QCnt: p.itemsInBuffer,
				Id:   string(data.id),
				D:    string(data.data),
				Buf:  buf,
				P:    p.portConf.Name,
			}
			qwrJson, _ := json.Marshal(qwr)
			h.broadcastSys <- qwrJson
		} else {
			// For reducing load on websocket, stop transmitting write data
			qwr := qwReport{
				Cmd:  "Write",
				QCnt: p.itemsInBuffer,
				Id:   string(data.id),
				D:    string(data.data),
				Buf:  buf,
				P:    p.portConf.Name,
			}
			qwrJson, _ := json.Marshal(qwr)
			h.broadcastSys <- qwrJson
		}

		// FINALLY, OF ALL THE CODE IN THIS PROJECT
		// WE TRULY/FINALLY GET TO WRITE TO THE SERIAL PORT!
		_, err := p.portIo.Write([]byte(data.data)) // n2, err :=

		// New Pause capability after we write. Added 9/23/15
		// This was needed because many Atmel microcontrollers just plain drop serial data
		// if it's being sent over while an EEPROM write is in play, so SPJS now
		// let's the user specify a pause after a serial command to solve for this error
		if data.pause > 0 {
			log.Printf("We are sleeping after the port write for milliseconds:%v\n", data.pause)
			time.Sleep(time.Duration(data.pause) * time.Millisecond)
		}

		// see if we need to send back the completeResponse
		if data.willHandleCompleteResponse == false {
			// we need to send back complete response
			// Send fake cmd:"Complete" back
			//strCmd := data.data
			m := CmdComplete{"CompleteFake", data.id, p.portConf.Name, -1, data.data}
			msgJson, err := json.Marshal(m)
			if err == nil {
				h.broadcastSys <- msgJson
			}

		}

		//log.Print("Just wrote ", n2, " bytes to serial: ", string(data.data))
		//log.Print(n2)
		//log.Print(" bytes to serial: ")
		//log.Print(data)
		if err != nil {
			errstr := "Error writing to " + p.portConf.Name + " " + err.Error() + " Closing port."
			log.Print(errstr)
			h.broadcastSys <- []byte(errstr)
			break
		}
	}
	msgstr := "Shutting down writer on " + p.portConf.Name
	log.Println(msgstr)
	h.broadcastSys <- []byte(msgstr)
	p.portIo.Close()
}

var spmutex = &sync.Mutex{}
var spIsOpening = false

func spHandlerOpen(portname string, baud int, isSecondary bool, dtrOn bool) {

	log.Print("Inside spHandler")

	if spIsOpening {
		log.Println("We are currently in the middle of opening a port. Returning...")
		return
	}
	spIsOpening = true
	spmutex.Lock()

	var out bytes.Buffer

	out.WriteString("Opening serial port ")
	out.WriteString(portname)
	out.WriteString(" at ")
	out.WriteString(strconv.Itoa(baud))
	out.WriteString(" baud")
	log.Print(out.String())

	isPrimary := true
	if isSecondary {
		isPrimary = false
	}

	conf := &SerialConfig{Name: portname, Baud: baud, RtsOn: true}
	conf.DtrOn = dtrOn

	// Needed for Arduino serial library
	mode := &serial.Mode{}
	mode.BaudRate = baud
	mode.DataBits = 8
	mode.Parity = 0
	mode.StopBits = 1
	mode.DTROn = dtrOn

	// Needed for original serial library
	// sp, err := serial.OpenPort(conf)
	// Needed for Arduino serial library
	sp, err := serial.Open(portname, mode)
	log.Print("Just tried to open port")
	if err != nil {
		//log.Fatal(err)
		log.Print("Error opening port " + err.Error())
		//h.broadcastSys <- []byte("Error opening port. " + err.Error())
		h.broadcastSys <- []byte("{\"Cmd\":\"OpenFail\",\"Desc\":\"Error opening port. " + err.Error() + "\",\"Port\":\"" + conf.Name + "\",\"Baud\":" + strconv.Itoa(conf.Baud) + "}")
		return
	}
	log.Print("Opened port successfully")
	sp.ResetInputBuffer()
	sp.ResetOutputBuffer()
	if dtrOn {
		sp.SetDTR(false)
	}
	//p := &serport{send: make(chan []byte, 256), portConf: conf, portIo: sp}
	// we can go up to 500,000 lines of gcode in the buffer
	p := &serport{sendBuffered: make(chan Cmd, 500000), sendNoBuf: make(chan Cmd), portConf: conf, portIo: sp, serialPort: sp, BufferType: "3Devo", IsPrimary: isPrimary, IsSecondary: isSecondary, isFeedRateOverrideOn: false}
	// if user asked for a buffer watcher, i.e. tinyg/grbl then attach here

	// nodemcu buffer only sends data back per line (which might be a bad call)
	// and it only sends 1 line at a time to the device and releases the next line
	// when it sees a > come back
	bw := &Bufferflow3Devo{Name: "3devo", Port: portname}
	bw.Init()
	p.bufferwatcher = bw

	sh.register <- p
	defer func() { sh.unregister <- p }()
	// this is internally buffered thread to not send to serial port if blocked
	go p.writerBuffered()
	// this is thread to send to serial port regardless of block
	go p.writerNoBuf()
	//v1.89 moved unlock here
	spIsOpening = false
	spmutex.Unlock()
	p.reader()
	//	go p.reader()
	//p.done = make(chan bool)
	//<-p.done

	// prior to 1.89 i had lock here.
	//	spIsOpening = false
	//	spmutex.Unlock()
}

func spHandlerCloseExperimental(p *serport) {
	h.broadcastSys <- []byte("Pre-closing serial port " + p.portConf.Name)
	p.isClosing = true
	//close the port

	p.bufferwatcher.Close()
	p.portIo.Close()
	h.broadcastSys <- []byte("Bufferwatcher closed")
	p.portIo.Close()
	//elicit response from hardware to close out p.reader()
	//_, _ = p.portIo.Write([]byte("?"))
	//p.portIo.Read(nil)

	//close(p.portIo)
	h.broadcastSys <- []byte("portIo closed")
	close(p.sendBuffered)
	h.broadcastSys <- []byte("p.sendBuffered closed")
	close(p.sendNoBuf)
	h.broadcastSys <- []byte("p.sendNoBuf closed")

	//p.done <- true

	// unregister myself
	// we already have a deferred unregister in place from when
	// we opened. the only thing holding up that thread is the p.reader()
	// so if we close the reader we should get an exit
	h.broadcastSys <- []byte("Closing serial port " + p.portConf.Name)
}

func spHandlerClose(p *serport) {
	p.isClosing = true
	//close the port
	//elicit response from hardware to close out p.reader()
	_, _ = p.portIo.Write([]byte("?"))
	p.bufferwatcher.Close()
	p.portIo.Close()
	// unregister myself
	// we already have a deferred unregister in place from when
	// we opened. the only thing holding up that thread is the p.reader()
	// so if we close the reader we should get an exit
	h.broadcastSys <- []byte("Closing serial port " + p.portConf.Name)
}
