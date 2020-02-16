package toxav

import (
	"fmt"
	"io/ioutil"
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
		this.seqnum = 0
		return nil
	}

	this.seqnum += 1
	return file
}
