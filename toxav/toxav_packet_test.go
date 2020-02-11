package toxrtp

import (
	"testing"
	"io/ioutil"
	"fmt"
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
		packet.Unmarshal(frame[12 + 1:]) // drop the channel byte, and the 

		if packet.data_length_lower != uint16(packet.data_length_full) {
			t.Logf("Data Length field differs from legacy field: %d != %d", packet.data_length_lower, packet.data_length_full)
		}

		fmt.Println(packet.String())


	}
}
