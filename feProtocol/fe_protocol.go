package feProtocol

import (
	"log"

	. "github.com/logrusorgru/aurora"
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

//FeProtocolHandler is a communication hub
//It provides handlers for incoming frames and error_frames.
//It also manages incoming nacks and acks and send frames
//that are waiting for a response
type FeProtocolHandler struct {
	ReceivedFrames []Frame
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
	switch protocolError {
	case CRC:
		log.Print(Red("CRC error occurred"))
		break
	case PAYLOAD_CORRUPTION:
		log.Print(Red("Payload corruption error occurred"))
		break
	case OTHER:
		log.Print(Red("Other error occurred"))
		break
	}
}
