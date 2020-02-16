package main

import (
	"fmt"
	"math/rand"

	"github.com/pion/webrtc/v2"

	"github.com/strfry/gotox/toxav"

	pionrtp "github.com/pion/rtp"
	pioncodecs "github.com/pion/rtp/codecs"

	"time"
)


func feed_track(t *webrtc.Track) {
	fmt.Println("Feedibg packets")
	var reader toxav.FrameReader
	packetizer := pionrtp.NewPacketizer(1200, t.PayloadType(), t.SSRC(), &pioncodecs.VP8Payloader{}, pionrtp.NewRandomSequencer(), 90000)
	//packetizer := t.Packetizer

	queue := toxav.NewRXQueue()

	for frame := reader.NextFrame(); frame != nil; frame = reader.NextFrame() {
//	for {
		var packet toxav.ToxAVPacket
		packet.Unmarshal(frame[1:]) // drop the channel/pt byte

		nalu := queue.Receive(&packet)

		if nalu != nil {
			packets:= packetizer.Packetize(nalu, 2000 /* samples, arbitrary ??? */)

			for _, packet := range packets {
	fmt.Println("RTP")
				t.WriteRTP(packet)
				//socket.Write(buf)
				time.Sleep(1 * time.Millisecond)
			}
		}
	}

	// TODO: close the connection
}

func main() {
	// Everything below is the Pion WebRTC API! Thanks for using it ❤️.

	// Wait for the offer to be pasted
	offer := webrtc.SessionDescription{}
	Decode(MustReadStdin(), &offer)

	// We make our own mediaEngine so we can place the sender's codecs in it. Since we are echoing their RTP packet
	// back to them we are actually codec agnostic - we can accept all their codecs. This also ensures that we use the
	// dynamic media type from the sender in our answer.
	mediaEngine := webrtc.MediaEngine{}

	// Setup the codecs you want to use.
    // Only support VP8, this makes our proxying code simpler
    mediaEngine.RegisterCodec(webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000))


	videoCodecs := mediaEngine.GetCodecsByKind(webrtc.RTPCodecTypeVideo)
	if len(videoCodecs) == 0 {
		panic("Offer contained no video codecs")
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))

	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	// Create a new RTCPeerConnection
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}
	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Create Track that we send video back to browser on
	outputTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeVP8, rand.Uint32(), "video", "pion")
	if err != nil {
		panic(err)
	}

	// Add this newly created track to the PeerConnection
	if _, err = peerConnection.AddTrack(outputTrack); err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())

		if "connected" == connectionState.String() {
			go feed_track(outputTrack)
		}


	})



	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	// Output the answer in base64 so we can paste it in browser
	fmt.Println(Encode(answer))

	// Block forever
	select {}
}
