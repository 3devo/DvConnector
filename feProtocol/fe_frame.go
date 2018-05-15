package feProtocol

import (
	"encoding/binary"
	"log"

	. "github.com/logrusorgru/aurora"
)

//Frame is the structure of a frame that's specified in the protocol specification.
type Frame struct {
	Sequence uint8
	Command  uint8
	Length   uint16
	Payload  []byte
}

//Print is a helper function to easily print a human readable frame
func (frame *Frame) Print() {
	log.Printf("\tSequence -> %v", frame.Sequence)
	log.Printf("\tCommand -> %v", frame.Command)
	log.Printf("\tLength -> %v", frame.Length)
	log.Printf("\tPayload -> %v -> %v", string(frame.Payload), frame.Payload)
}

//ToBytes turns the frame structure into an array of bytes
//so it can be easily through a communication channel
func (frame *Frame) ToBytes() []byte {
	buffer := []byte{}
	buffer = append(buffer, []byte{frame.Command}...)
	lengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lengthBytes, frame.Length)
	buffer = append(buffer, lengthBytes...)
	buffer = append(buffer, frame.Payload...)
	return buffer
}

//UnstuffPayload removes the escapebytes from the payload
func (frame *Frame) UnStuffPayload() {
	payload := frame.Payload
	escapeByte := uint8(0)
	for i := 0; i < len(payload)-1; i++ {
		switch payload[i] {
		case END_FRAME, START_FRAME, START_FRAME ^ SEQUENCE_MASK:
			{
				log.Print(Green(payload[i]))
				if escapeByte != 0 {
					if i+1 < len(payload)-1 {
						payload = append(payload[:i], payload[i+1:]...)
					} else {
						payload = payload[:len(payload)-1]
					}
				}
				escapeByte = payload[i]
			}
		}
	}
	frame.Payload = payload
}

//StuffPayload checks the given frame payload
//For the predefined protocol frames and if they are found
//They will be stuffed with a copy of the found byte
func (frame *Frame) StuffPayload() {
	payload := frame.Payload
	length := uint16(len(frame.Payload))
	for i := 0; i < len(frame.Payload); i++ {
		switch payload[i] {
		case END_FRAME, START_FRAME, START_FRAME ^ SEQUENCE_MASK:
			{
				if i < len(frame.Payload)-1 {
					payload = append(payload[:i], append([]byte{payload[i]}, payload[i:]...)...)
				} else {
					payload = append(payload, payload[i])
				}
				i++
				length++
			}
		}
	}

	frame.Length = length
	frame.Payload = payload
}
