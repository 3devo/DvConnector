package feProtocol

import (
	"encoding/binary"
	"log"
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
	buffer := []byte{}
	length := int(frame.Length)
	for i := 0; i < length; i++ {
		switch frame.Payload[i] {
		case START_FRAME:
		case END_FRAME:
			{
				buffer = append(buffer, frame.Payload[i])
				i++
				break
			}
		default:
			buffer = append(buffer, frame.Payload[i])
		}
	}
	frame.Length = uint16(len(buffer))
	frame.Payload = buffer
}

//StuffPayload checks the given frame payload
//For the predefined protocol frames and if they are found
//They will be stuffed with a copy of the found byte
func (frame *Frame) StuffPayload() {
	buffer := []byte{}
	length := uint16(len(frame.Payload))
	counter := 0
	for i := 0; i < int(length); i++ {
		index := uint8(frame.Payload[counter])
		counter++
		switch index {
		case START_FRAME, END_FRAME:
			{
				buffer = append(buffer, index)
				i++
				buffer = append(buffer, index)
				length++
				break
			}
		default:
			buffer = append(buffer, index)
		}
	}
	frame.Length = length
	frame.Payload = buffer
	// log.Fatal(length, "  ", string(buffer))
}
