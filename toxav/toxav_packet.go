package toxav

// implement Tox's custom RTP "Payloading"

import (
	"fmt"
	"encoding/binary"

)

type ToxAVPayloader struct{}


const (
	toxAVHeaderSize = 80
)

// Payload fragments a VP8 packet across one or more byte arrays
func (p *ToxAVPayloader) Payload(mtu int, payload []byte) [][]byte {
	maxFragmentSize := mtu - toxAVHeaderSize

	payloadData := payload
	payloadDataRemaining := len(payload)

	//var toxhd ToxAVPacket


	payloadDataIndex := 0
	var payloads [][]byte

	// Make sure the fragment/payload size is correct
	if min(maxFragmentSize, payloadDataRemaining) <= 0 {
		return payloads
	}
	for payloadDataRemaining > 0 {
		currentFragmentSize := min(maxFragmentSize, payloadDataRemaining)
		out := make([]byte, toxAVHeaderSize+currentFragmentSize)
		if payloadDataRemaining == len(payload) {
			out[0] = 0x10
		}

		copy(out[toxAVHeaderSize:], payloadData[payloadDataIndex:payloadDataIndex+currentFragmentSize])
		payloads = append(payloads, out)

		payloadDataRemaining -= currentFragmentSize
		payloadDataIndex += currentFragmentSize
	}

	return payloads
}

// ToxAVPacket represents the additional headers added before start of the payload by ToxAV
type ToxAVPacket struct {
	rtp_head1 byte // to lazy to write out all these bitfields right now
	rtp_head2 byte // they're not set correctly anyway

	seqnum uint16
	timestamp uint32
	ssrc uint32
	
	// Those proprietary tox fields use the fully extended CSRC headers:

    // Bit mask of \ref RTPFlags setting features of the current frame.
	flags uint64
	offset_full uint32
	data_length_full uint32
	received_length_full uint32

	// These are ToxBlinkenwall specific extensions
	frame_record_timestamp uint64
	fragment_num uint32
	real_frame_num uint32
	encoder_bit_rate_used uint32
	client_video_capture_delay_ms uint32
	// End of extensions

	// Some more "padding" fields left...
	padding0 uint32
	padding1 uint32
	padding2 uint32
	padding3 uint32
	padding4 uint32

	// These are from a former ToxAV version, that couldn't handle packets larger than 64k
	offset_lower uint16
	data_length_lower uint16

	Payload []byte
}

func (p ToxAVPacket) String() string {
	out := "RTP PACKET:\n"

	out += fmt.Sprintf("\tFlags: %d\n", p.flags)
	out += fmt.Sprintf("\tOffset: %d\n", p.offset_full)
	out += fmt.Sprintf("\tOffset (lower): %d\n", p.offset_lower)
	out += fmt.Sprintf("\tDataLength: %d\n", p.data_length_full)
	out += fmt.Sprintf("\tDataLength (lower): %d\n", p.offset_lower)
	out += fmt.Sprintf("\tReceivedLength: %d\n", p.received_length_full)

	
	out += fmt.Sprintf("\tFrameRecordTimestamp: %d\n", p.frame_record_timestamp)
	out += fmt.Sprintf("\tEncoderBitrateUser: %d\n", p.encoder_bit_rate_used)
	out += fmt.Sprintf("\tClientVideoCaptureDelayMs: %d\n", p.client_video_capture_delay_ms)


	out += fmt.Sprintf("\tPadding: %d %d %d %d %d\n", p.padding0, p.padding1, p.padding2, p.padding3, p.padding4)


	return out
}


// Unmarshal parses the passed byte slice and stores the result in the ToxAVPacket this method is called upon
func (p *ToxAVPacket) Unmarshal(payload []byte) ([]byte, error) {
	if payload == nil {
		return nil, fmt.Errorf("invalid nil packet")
	}

	payloadLen := len(payload)

	if payloadLen < 17 /*constant likely way off*/ {
		return nil, fmt.Errorf("Payload is not large enough to container header")
	}

	// TODO(strfry): read the other RTP fields
	p.rtp_head1 = payload[0]
	p.rtp_head2 = payload[1]
	
	p.seqnum = binary.BigEndian.Uint16(payload[2 : 4])
	p.timestamp = binary.BigEndian.Uint32(payload[4:8])
	p.timestamp = binary.BigEndian.Uint32(payload[8:12])

	p.flags = binary.BigEndian.Uint64(payload[12 : 20])
	p.offset_full = binary.BigEndian.Uint32(payload[20 : 24])
	p.data_length_full = binary.BigEndian.Uint32(payload[24 : 28])
	p.received_length_full = binary.BigEndian.Uint32(payload[28 : 32])

	p.frame_record_timestamp = binary.BigEndian.Uint64(payload[32 : 40])
	p.fragment_num = binary.BigEndian.Uint32(payload[40 : 44])
	p.real_frame_num = binary.BigEndian.Uint32(payload[44 : 48])
	p.encoder_bit_rate_used = binary.BigEndian.Uint32(payload[48 : 52])
	p.client_video_capture_delay_ms = binary.BigEndian.Uint32(payload[52 : 56])

	// payload[56:76] // 5xu32 padding fields

	p.offset_lower = binary.BigEndian.Uint16(payload[76 : 78])
	p.data_length_lower = binary.BigEndian.Uint16(payload[78 : 80])

	if toxAVHeaderSize >= payloadLen {
		return nil, fmt.Errorf("Payload is not large enough")
	}
	p.Payload = payload[toxAVHeaderSize:]
	return p.Payload, nil
}


// TODO(strfry): move to common
func min(a, b int) int {
	if a < b {
			return a
	}
	return b
}
