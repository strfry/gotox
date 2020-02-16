package toxrtp

import (
	"testing"
	"io/ioutil"
	"fmt"

	"github.com/notedit/gstreamer-go"
	"time"

	pionrtp "github.com/pion/rtp"
	pioncodecs "github.com/pion/rtp/codecs"
)

type FrameReader struct
{
	seqnum int
}

func (this *FrameReader) NextFrame() []byte {
	filename := fmt.Sprintf("toxav_dumps/frame_%05d.bin", this.seqnum)

	file, err := ioutil.ReadFile(filename)

	if err != nil {
		//log.Println("ReadFile failed at", filename)
		return nil
	}

	this.seqnum += 1
	return file
}

func TestToxAVPacket_FirstTest(t *testing.T) {
	t.Log("first test")
}


func TestToxAVPacketUnmarshal(t *testing.T) {
	fmt.Println("first test")
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



func TestToxAV_RXQueue(t *testing.T) {
	var reader FrameReader
	queue := NewQueue()


	//pipeline, err := gstreamer.New("appsrc name=mysource ! rtpvp8depay ! vp8dec ! autovideosink")
	pipeline, err := gstreamer.New("appsrc name=mysource ! udpsink host=::1 port=1337")
	
	if err != nil {
		t.Error("pipeline create error", err)
		t.FailNow()
	}
	appsrc := pipeline.FindElement("mysource")
	
	//appsrc.SetCap("application/x-rtp")
	appsrc.SetCap("application/x-rtp, media=(string)video, clock-rate=(int)90000, encoding-name=(string)VP8")
	
	pipeline.Start()	


	packetizer := pionrtp.NewPacketizer(1200, 96, 0x13371137, &pioncodecs.VP8Payloader{}, pionrtp.NewRandomSequencer(), 90000)
	
	for frame := reader.NextFrame(); frame != nil; frame = reader.NextFrame() {
		var packet ToxAVPacket
		packet.Unmarshal(frame[1:]) // drop the channel/pt byte

		nalu := queue.Receive(&packet)

		if nalu != nil {
			packets:= packetizer.Packetize(nalu, 2000 /* samples, arbitrary */)

			fmt.Println("NALU split into ", len(packets), " packets")

			for _, packet := range packets {
				buf, err := packet.Marshal()

				if err != nil {
					panic(err)
				}

			//fmt.Println("Writing packet ", buf)
				appsrc.Push(buf)
				time.Sleep(1 * time.Millisecond)
			}
		}
	}


}