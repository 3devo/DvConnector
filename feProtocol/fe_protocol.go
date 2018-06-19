package feProtocol

import (
	"encoding/binary"
	"io"
	"log"
	"time"

	. "github.com/logrusorgru/aurora"
	"github.com/sigurn/crc8"
)

var crc8Table = crc8.MakeTable(crc8.Params{0x07, 0xff, false, false, 0x00, 0xF4, "CRC-8"})

type FeProtocolReturntypes bool

const (
	OK  FeProtocolReturntypes = false
	NOK FeProtocolReturntypes = true
)

type FeProtocolError int

//Enum for the different error types
const (
	NONE               FeProtocolError = iota
	CRC                FeProtocolError = iota
	PAYLOAD_CORRUPTION FeProtocolError = iota
	OTHER              FeProtocolError = iota
)

//Constants for defined protocol frames
const (
	START_FRAME   uint8 = uint8(0xaa)
	END_FRAME     uint8 = uint8(0x55)
	SEQUENCE_MASK uint8 = uint8(0x80)
)

type BufferedReader struct {
	buffer []byte
	PortIo io.ReadWriteCloser
}

//FeProtocolHandler is a communication hub
//It provides handlers for incoming frames and error_frames.
//It also manages incoming nacks and acks and send frames
//that are waiting for a response
type FeProtocolHandler struct {
	ReceivedFrames []Frame
	Transport      BufferedReader
}

func (bufferedreader *BufferedReader) ReadBytes(n int) ([]byte, bool, int) {
	bytes := make([]byte, n)
	read, err := bufferedreader.PortIo.Read(bytes)
	return bytes, (read < 0 || err != nil), read
}

//OnFrameReceived is called when a frame gets successfully received.
//From OnFrameReceived the frame can be used and propagated through the rest of the system.
func (fe *FeProtocolHandler) OnFrameReceived(frame *Frame) {
	fe.ReceivedFrames = append(fe.ReceivedFrames, *frame)
	frame.UnStuffPayload()
	log.Print(Green("START frame"))
	frame.Print()
	log.Print(Green("END frame\n\n"))

	switch frame.Command {
	case 2:
		{
			// log.Println(Magenta("Setup command received"))
			// log.Println(binary.LittleEndian.(frame.Payload))
			test := time.NewTimer(1 * time.Second)
			<-test.C
			heater := binary.LittleEndian.Uint16(frame.Payload)
			if heater <= 190 {
				heater = 190 + 10
			} else {
				heater = heater - 10
			}
			heaterBytes := make([]byte, 2)
			binary.LittleEndian.PutUint16(heaterBytes, heater)
			newFrame := Frame{
				Sequence: 1,
				Command:  2,
				Length:   uint16(3),
				Payload:  append([]byte{0}, heaterBytes...)}
			newFrame.StuffPayload()
			startFrame := uint8(START_FRAME)
			packet := []byte{startFrame}
			packet = append(packet, newFrame.ToBytes()...)
			packet = append(packet, []byte{crc8.Checksum(append(packet, []byte{END_FRAME}...), crc8Table)}...)
			packet = append(packet, []byte{END_FRAME}...)
			fe.Transport.PortIo.Write(packet)
			log.Println(packet)
			// // setup := []string{}
			// // for i := 0; i < len(frame.Payload); i++ {
			// // 	log.Println(frame.Payload[i])
			// // 	length := int(frame.Payload[i])
			// // 	setup = append([]string{string(frame.Payload[i:length])}, setup...)
			// // 	i = i + length
			// // }
		}
	case 4:
		// log.Println(Magenta("Setup command received"))
		log.Printf("Received mem -> %v", binary.LittleEndian.Uint16(frame.Payload))
	}
}

//OnFrameError handles failed incomming frames
//Based on the implementation OnFrameError could send a NACk for a failed frame.
//Or just discard the message
func (fe *FeProtocolHandler) OnFrameError(protocolError FeProtocolError) {
	errorMessage := ""
	switch protocolError {
	case CRC:
		errorMessage = "CRC error occurred"
		break
	case PAYLOAD_CORRUPTION:
		errorMessage = "Payload corruption error occurred"
		break
	case OTHER:
		errorMessage = "Other error occurred"
		break
	}
	newFrame := Frame{
		Sequence: 1,
		Command:  2,
		Length:   uint16(3),
		Payload:  []byte{0, 190, 0}}
	newFrame.StuffPayload()
	startFrame := uint8(START_FRAME)
	packet := []byte{startFrame}
	packet = append(packet, newFrame.ToBytes()...)
	packet = append(packet, []byte{crc8.Checksum(append(packet, []byte{END_FRAME}...), crc8Table)}...)
	packet = append(packet, []byte{END_FRAME}...)
	fe.Transport.PortIo.Write(packet)
	log.Println(packet)
	log.Print(Red("Next 1.0 -> " + errorMessage))
}

//ReadInPayload reads in the byte one by one through the transport layer.
//And checks if the payload is escaped correctly if not it will return an error else it will return the payload
func (handler *FeProtocolHandler) ReadInPayload(length uint16) ([]byte, FeProtocolReturntypes) {
	escapeByte := uint8(0)
	payload := []byte{}
	read := 0
	for read < int(length) {
		buffer, _, b := handler.Transport.ReadBytes(1)
		// log.Print(payload, length)
		if b > 0 {
			switch buffer[0] {
			case START_FRAME, START_FRAME ^ SEQUENCE_MASK, END_FRAME:
				{
					if escapeByte != 0 && escapeByte != buffer[0] {
						return nil, NOK
					} else if escapeByte == buffer[0] {
						escapeByte = 0
					} else {
						escapeByte = buffer[0]
					}
					break
				}
			default:
				{
					if escapeByte != 0 {
						return nil, NOK
					}
				}

			}
			read = read + b
			payload = append(payload, buffer[:b]...)
		}
	}
	return payload, OK
}
