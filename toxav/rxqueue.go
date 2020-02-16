package toxrtp

import(
	"fmt"
)

// Reassembles the packets encoded in "Tox proprietary RTP Framing Extension"
type RXQueue struct {
	slots map[uint16][]byte
	last_sent_timestamp uint32
}

// this probably isnt idiomatic...
func NewQueue() RXQueue {
	var queue RXQueue
	queue.slots = make(map[uint16][]byte)
	return queue
}

func (this *RXQueue) Receive(p *ToxAVPacket) []byte {
	// seqnum stays constant between packets of the same NALU, and will serve as our key
	seqnum := p.seqnum

	buffer := this.slots[seqnum]

	// search for slot in buffer
	if buffer == nil {
		buffer = make([]byte, p.data_length_full)
		this.slots[seqnum] = buffer
	} else {
		if len(buffer) != int(p.data_length_full){
			fmt.Println("Warning: received inconsistent packets")
			return nil
		}
	}

	copy(buffer[p.offset_full:], p.Payload)

	// now the hacky part:
	if int(p.offset_full) + len(p.Payload) == len(buffer) {
		delete(this.slots, seqnum)
		return buffer
	} else {
		//fmt.Println("not full frame", p.offset_full + uint32(len(p.Payload)), p.data_length_full)
	}

	return nil

	// TODO(strfry): clean up old slots
}