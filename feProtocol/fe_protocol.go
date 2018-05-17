package feProtocol

import (
	"io"
	"log"

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
	log.Print(Red(errorMessage))
}

//ReadInPayload reads in the byte one by one through the transport layer.
//And checks if the payload is escaped correctly if not it will return an error else it will return the payload
func (handler *FeProtocolHandler) ReadInPayload(length uint16) ([]byte, FeProtocolReturntypes) {
	escapeByte := uint8(0)
	payload := []byte{}
	read := 0

	for read < int(length) {
		buffer, _, b := handler.Transport.ReadBytes(1)
		if b > 0 {
			switch buffer[0] {
			case START_FRAME, START_FRAME ^ SEQUENCE_MASK, END_FRAME:
				{
					if escapeByte != 0 && escapeByte != buffer[0] || (escapeByte && read >= length-1) {
						return nil, NOK
					}
					escapeByte = buffer[0]
					break
				}
			default:
				{
					escapeByte = 0
				}

			}
			read = read + b
			payload = append(payload, buffer[:b]...)
		}
	}
	return payload, OK
}
