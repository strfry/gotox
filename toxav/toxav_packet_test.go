package toxav

import (
	"testing"
	"fmt"

	"net"
	"time"

	pionrtp "github.com/pion/rtp"
	pioncodecs "github.com/pion/rtp/codecs"
)


func TestToxAVPacketUnmarshal(t *testing.T) {
	var reader FrameReader
	var packet ToxAVPacket

	counter := 0

	frame := reader.NextFrame()
	for ; frame != nil; frame = reader.NextFrame() {
		counter++
		packet.Unmarshal(frame[1:]) // drop the channel/pt byte

		if uint16(packet.data_length_lower) != uint16(packet.data_length_full) {
			t.Fatal("Difference between data_length fields in ", packet.String())
		}
	}
}


// Verify RTP transmission with this command:
// gst-launch-1.0 -vvv udpsrc address=:: port=1337 caps="application/x-rtp" ! rtpvp8depay ! vp8dec ! autovideosink
func TestToxAV_RXQueue(t *testing.T) {
	var reader FrameReader
	queue := NewRXQueue()

	packetizer := pionrtp.NewPacketizer(1200, 96, 0x13371137, &pioncodecs.VP8Payloader{}, pionrtp.NewRandomSequencer(), 90000)

	socket, err := net.Dial("udp", "[::1]:1337")

	if err != nil {
		panic(err)
	}

	for frame := reader.NextFrame(); frame != nil; frame = reader.NextFrame() {
		var packet ToxAVPacket
		packet.Unmarshal(frame[1:]) // drop the channel/pt byte

		nalu := queue.Receive(&packet)

		if nalu != nil {
			if ! (&pioncodecs.VP8PartitionHeadChecker{}).IsPartitionHead(nalu) {
				t.Fatal("VP8 Packet is not a partition header")
			}

			packets:= packetizer.Packetize(nalu, 2000 /* samples, arbitrary ??? */)

			for _, packet := range packets {
				buf, err := packet.Marshal()

				if err != nil {
					fmt.Println(buf)
					panic(err)
				}

				socket.Write(buf)
				time.Sleep(1 * time.Millisecond)
			}
		}
	}


}
