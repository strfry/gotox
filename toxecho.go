package main

import (
	"fmt"
	"io/ioutil"
	"log"
	// "os"
	"math/rand"
	"time"

	"github.com/strfry/go-toxcore-c"
	"github.com/notedit/gstreamer-go"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

var server = []interface{}{
	"205.185.116.116", uint16(33445), "A179B09749AC826FF01F37A9613F6B57118AE014D4196A0E1105A98F93A54702",
}
var fname = "./toxecho.data"
var debug = true
var nickPrefix = "GoToX."
var statusText = "Send me audio, video."

func start_gstreamer_pipeline(msi *tox.MSICall) {
	pipeline, err := gstreamer.New("videotestsrc  ! capsfilter name=filter ! vp8enc ! rtpvp8pay ! appsink name=sink")
	if err != nil {
		log.Println("pipeline create error", err)
		//t.FailNow()
	}

	filter := pipeline.FindElement("filter")

	if filter == nil {
		//t.Error("pipeline find element error ")
		log.Println("pipeline find element error", filter)
	}

	filter.SetCap("video/x-raw,width=1280,height=720")

	appsink := pipeline.FindElement("sink")

	pipeline.Start()

	out := appsink.Poll()

	for {
		buffer := <-out
		fmt.Println("push ", len(buffer))

		t := msi.Tox()
		friendNumber := msi.FriendNumber()
		t.FriendSendLossyPacket(friendNumber, string(buffer))
	}
}

func main() {
	log.Println("!!! main")

	opt := tox.NewToxOptions()
	opt.Local_discovery_enabled = true

	if tox.FileExist(fname) {
		data, err := ioutil.ReadFile(fname)
		if err != nil {
			log.Println(err)
		} else {
			opt.Savedata_data = data
			opt.Savedata_type = tox.SAVEDATA_TYPE_TOX_SAVE
		}
	}
	opt.Tcp_port = 33445

	var t *tox.Tox
	for i := 0; i < 5; i++ {
		t = tox.NewTox(opt)
		if t == nil {
			opt.Tcp_port += 1
		} else {
			break
		}
	}

	r, err := t.Bootstrap(server[0].(string), server[1].(uint16), server[2].(string))
	r2, err := t.AddTcpRelay(server[0].(string), server[1].(uint16), server[2].(string))
	if debug {
		log.Println("bootstrap:", r, err, r2)
		log.Println("bootstrap:", r, err, r2)
	}

	pubkey := t.SelfGetPublicKey()
	seckey := t.SelfGetSecretKey()
	toxid := t.SelfGetAddress()
	if debug {
		log.Println("keys:", pubkey, seckey, len(pubkey), len(seckey))
	}
	log.Println("toxid:", toxid)

	defaultName := t.SelfGetName()
	humanName := nickPrefix + toxid[0:5]
	if humanName != defaultName {
		t.SelfSetName(humanName)
	}
	humanName = t.SelfGetName()
	if debug {
		log.Println(humanName, defaultName, err)
	}

	defaultStatusText, err := t.SelfGetStatusMessage()
	if defaultStatusText != statusText {
		t.SelfSetStatusMessage(statusText)
	}
	if debug {
		log.Println(statusText, defaultStatusText, err)
	}

	sz := t.GetSavedataSize()
	sd := t.GetSavedata()
	if debug {
		log.Println("savedata:", sz, t)
		log.Println("savedata", len(sd), t)
	}
	err = t.WriteSavedata(fname)
	if debug {
		log.Println("savedata write:", err)
	}

	// add friend norequest
	fv := t.SelfGetFriendList()
	for _, fno := range fv {
		fid, err := t.FriendGetPublicKey(fno)
		if err != nil {
			log.Println(err)
		} else {
			t.FriendAddNorequest(fid)
		}
	}
	if debug {
		log.Println("add friends:", len(fv))
	}

	// callbacks
	t.CallbackSelfConnectionStatus(func(t *tox.Tox, status int, userData interface{}) {
		if debug {
			log.Println("on self conn status:", status, userData)
		}
	}, nil)
	t.CallbackFriendRequest(func(t *tox.Tox, friendId string, message string, userData interface{}) {
		log.Println(friendId, message)
		num, err := t.FriendAddNorequest(friendId)
		if debug {
			log.Println("on friend request:", num, err)
		}
		if num < 100000 {
			t.WriteSavedata(fname)
		}
	}, nil)
	t.CallbackFriendMessage(func(t *tox.Tox, friendNumber uint32, message string, userData interface{}) {
		if debug {
			log.Println("on friend message:", friendNumber, message)
		}
		n, err := t.FriendSendMessage(friendNumber, "Re: "+message)
		if err != nil {
			log.Println(n, err)
		}
	}, nil)
	t.CallbackFriendConnectionStatus(func(t *tox.Tox, friendNumber uint32, status int, userData interface{}) {
		if debug {
			friendId, err := t.FriendGetPublicKey(friendNumber)
			log.Println("on friend connection status:", friendNumber, status, friendId, err)
		}
	}, nil)

	t.CallbackFriendLossyPacketAdd(func(t *tox.Tox, friendNumber uint32, data string, userData interface{}) {
		if debug {
			//var pkgid = data[0]
                        if rand.Int()%23 == 3 {
			//	log.Println("got lossy data from, pkgid, data :", friendNumber, pkgid, data)
			}

			//err := t.FriendSendLossyPacket(friendNumber, data)
			if err != nil {
				log.Println("FriendSendLossyPacket error :", err)
			}
		}
	}, nil)

	t.CallbackFriendLosslessPacketAdd(func(t *tox.Tox, friendNumber uint32, data string, userData interface{}) {
		if debug {
			var pkgid = data[0]
                        if rand.Int()%23 == 3 {
				log.Println("got lossless data from, pkgid, data :", friendNumber, pkgid, data)
			}

			//err := t.FriendSendLosslessPacket(friendNumber, data)
			if err != nil {
				log.Println("FriendSendLosslessPacket error :", err)
			}
		}
	}, nil)

	// audio/video
	msi, err := tox.NewMSISession(t)
	if err != nil {
		log.Println(err, msi)
	}

	// MSI_ON_INVITE
	msi.RegisterCallback(tox.MSI_ON_INVITE, func(x interface{}, call *tox.MSICall) {
		log.Println("!!! MSI ACTION CALLBACK", call)
		var err = call.Answer(255)
		log.Println("call.Answer: ", err)

		go start_gstreamer_pipeline(call)
	})

/*
	msi.RegisterCallback(tox.MSI_ON_START, func(x interface{}, call *tox.MSICall) {
		log.Println("MSI_ON_START")
	})
	msi.RegisterCallback(tox.MSI_ON_END, func(x interface{}, call *tox.MSICall) {
		log.Println("MSI_ON_STOP")
	})
	msi.RegisterCallback(tox.MSI_ON_ERROR, func(x interface{}, call *tox.MSICall) {
		log.Println("MSI_ON_ERROR")
	})
	msi.RegisterCallback(tox.MSI_ON_PEERTIMEOUT, func(x interface{}, call *tox.MSICall) {
		log.Println("MSI_ON_PEERTIMEOUT")
	})
	msi.RegisterCallback(tox.MSI_ON_CAPABILITIES, func(x interface{}, call *tox.MSICall) {
		log.Println("MSI_ON_CAPABILITIES")
	})
*/
	// toxcore loops
	shutdown := false
	loopc := 0
	itval := 0
	for !shutdown {
		iv := t.IterationInterval()
		if iv != itval {
			if debug {
				if itval-iv > 20 || iv-itval > 20 {
					log.Println("tox itval changed:", itval, iv)
				}
			}
			itval = iv
		}

		t.Iterate()
		status := t.SelfGetConnectionStatus()
		if loopc%5500 == 0 {
			if status == 0 {
				if debug {
					fmt.Print(".")
				}
			} else {
				if debug {
					fmt.Print(status, ",")
				}
			}
		}
		loopc += 1
		time.Sleep(1000 * 50 * time.Microsecond)
	}

	t.Kill()
}

func makekey(no uint32, a0 interface{}, a1 interface{}) string {
	return fmt.Sprintf("%d_%v_%v", no, a0, a1)
}

func _dirty_init() {
	log.Println("ddddddddd")
	tox.KeepPkg()
}
