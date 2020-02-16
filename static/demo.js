/* eslint-env browser */

let pc = new RTCPeerConnection({
})
let log = msg => {
  document.getElementById('logs').innerHTML += msg + '<br>'
}

pc.ontrack = function (event) {
  var el = document.createElement(event.track.kind)
  el.srcObject = event.streams[0]
  el.autoplay = true
  el.controls = true

  document.getElementById('remoteVideos').appendChild(el)
}

pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)
pc.onicecandidate = async event => {
  if (event.candidate === null) {
    let sdp = btoa(JSON.stringify(pc.localDescription))
    document.getElementById('localSessionDescription').value = sdp

    const response = await fetch("/post/sdp", {
        method: "POST", 
        body: sdp
      })
    
    const body = await response.text()
    const answer = JSON.parse(atob(body))
    console.log("Request complete! response:", answer);  
    try {
      pc.setRemoteDescription(new RTCSessionDescription(answer))
    } catch (e) {
      alert(e)
    }
  }
}


// Offer to receive 1 audio, and 2 video tracks
pc.addTransceiver('audio', {'direction': 'recvonly'})
pc.addTransceiver('video', {'direction': 'recvonly'})
pc.addTransceiver('video', {'direction': 'recvonly'})
pc.createOffer().then(d => pc.setLocalDescription(d)).catch(log)
