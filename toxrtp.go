package main

// try to emulate the Tox "RTP" protocol from RTP, and vice-versa

/*

#include <tox/tox.h>

struct RTPHeader {
    // Standard RTP header 
    unsigned ve: 2; // Version has only 2 bits! 
    unsigned pe: 1; // Padding 
    unsigned xe: 1; // Extra header 
    unsigned cc: 4; // Contributing sources count 

    unsigned ma: 1; // Marker 
    unsigned pt: 7; // Payload type 

    uint16_t sequnum;
    uint32_t timestamp;
    uint32_t ssrc;

    // Non-standard Tox-specific fields

    
    // Bit mask of \ref RTPFlags setting features of the current frame.
	uint64_t flags;
	
    // The full 32 bit data offset of the current data chunk. The \ref
    // offset_lower data member contains the lower 16 bits of this value. For
    // frames smaller than 64KiB, \ref offset_full and \ref offset_lower are
    // equal.
    uint32_t offset_full;
    // The full 32 bit payload length without header and packet id.
    uint32_t data_length_full;
    // Only the receiver uses this field (why do we have this?).
    uint32_t received_length_full;
    // Data offset of the current part (lower bits).
    uint16_t offset_lower;
    // Total message length (lower bits).
    uint16_t data_length_lower;
};


struct RTPMessage {
     // This is used in the old code that doesn't deal with large frames, i.e.
     // the audio code or receiving code for old 16 bit messages. We use it to
     // record the number of bytes received so far in a multi-part message. The
     // multi-part message in the old code is stored in \ref RTPSession::mp.
    uint16_t len;

    struct RTPHeader header;
    uint8_t data[];
};

enum { TOX_MTU = 1200 };

*/
import "C"
import (
	"log"
	"github.com/pion/rtp"
)

const (
	TOXRTP_TYPE_VIDEO = 193
)

func RTPToTox(data []byte) ([]byte) {
	packet := new(rtp.Packet)
	packet.Unmarshal(data)

	log.Println("RTPToTox: found header: ",packet.String())

	if packet.Header.Extension {
		panic("RTP packets with extension header not supported!")
	}

	if len(packet.Header.CSRC) > 0 {
		panic("RTP packets with contributing sources header not supported!")
	}

	//payload := packet.Payload

	buffer := new([C.TOX_MTU]byte)

	buffer[0] = TOXRTP_TYPE_VIDEO
	//copy(buffer[1:], packet.Header)

	header_offset := 1 + 32

	copy(buffer[header_offset:], data[packet.Header.PayloadOffset:])

	//serialized, err := packet.Marshal()
	return buffer[:]
}

func ToxToRTP(data []byte) []byte {
	packet := new(rtp.Packet)
	err := packet.Unmarshal(data)

	if err != nil {
		panic(err)
	}

	log.Println("ToxToRTP: decoded header: ", packet.String())

	return data
}