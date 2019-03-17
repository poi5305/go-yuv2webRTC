package gomobilelib

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/poi5305/go-yuv2webRTC/webrtc"
)

var webRTC *webrtc.WebRTC
var width int
var height int

func init() {
	webRTC = webrtc.NewWebRTC()
}

// InitWebRTC expose to android
func InitWebRTC(w, h int) {
	width = w
	height = h

	router := mux.NewRouter()
	router.HandleFunc("/", getWeb).Methods("GET")
	router.HandleFunc("/session", postSession).Methods("POST")

	go http.ListenAndServe(":8000", router)
}

func nV21ToYuv420(nv21 []byte) []byte {
	yuv := make([]byte, len(nv21))
	framesize := width * height
	copy(yuv, nv21[0:framesize])
	u := 0
	v := 0
	for j := 0; j < framesize/2; j += 2 {
		yuv[framesize+u] = nv21[framesize+j+1]
		yuv[framesize*5/4+v] = nv21[framesize+j]
		u++
		v++
	}
	return yuv
}

// OnPreviewFrame expose to android for writting image. Convert nv21 to yuv.
func OnPreviewFrame(nv21 []byte) {
	if webRTC.IsConnected() {
		if len(webRTC.ImageChannel) < cap(webRTC.ImageChannel) {
			yuv := nV21ToYuv420(nv21)
			webRTC.ImageChannel <- yuv
		}
	}
}

func getWeb(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(html))
}

func postSession(w http.ResponseWriter, r *http.Request) {
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	r.Body.Close()

	localSession, err := webRTC.StartClient(string(bs), width, height)
	if err != nil {
		log.Fatalln(err)
	}

	w.Write([]byte(localSession))
}

var html = `
<html>

<style>
textarea {
  width: 60%;
  height: 50px;
}
</style>
<div>
  <a href="https://github.com/poi5305/go-yuv2webRTC">https://github.com/poi5305/go-yuv2webRTC</a>
  <br />
  <a href="https://github.com/pions/webrtc/tree/v1.2.0/examples/gstreamer-send/jsfiddle">https://github.com/pions/webrtc/tree/v1.2.0/examples/gstreamer-send/jsfiddle</a>
</div>

<div id="remoteVideos"></div> <br />
Browser base64 Session Description <br /><textarea id="localSessionDescription" readonly="true"></textarea> <br />

Golang base64 Session Description: <br /><textarea id="remoteSessionDescription"> </textarea> <br/>

<button onclick="window.startSession()"> Start Session </button>
<div id="div"></div>
  
<div>
  Refresh to retry
</div>
<script>
function postSession(session) {
  if (session == "") {
    return;
  }
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4 && this.status == 200) {
      document.getElementById('remoteSessionDescription').value = this.responseText;
    }
  };
  xhttp.open("POST", "/session", true);
  xhttp.setRequestHeader("Content-type", "text/plain");
  xhttp.send(session);
}

let pc = new RTCPeerConnection({
  iceServers: [
    {
      urls: 'stun:stun.l.google.com:19302'
    }
  ]
})
let log = msg => {
  document.getElementById('div').innerHTML += msg + '<br>'
}

pc.ontrack = function (event) {
  var el = document.createElement(event.track.kind)
  el.srcObject = event.streams[0]
  el.autoplay = true
  el.controls = true

  document.getElementById('remoteVideos').appendChild(el)
}

pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)
pc.onicecandidate = event => {
  if (event.candidate === null) {
    var session = btoa(JSON.stringify(pc.localDescription));
    document.getElementById('localSessionDescription').value = session;
    postSession(session)
  }
}

pc.createOffer({offerToReceiveVideo: true, offerToReceiveAudio: true}).then(d => pc.setLocalDescription(d)).catch(log)

window.startSession = () => {
  let sd = document.getElementById('remoteSessionDescription').value
  if (sd === '') {
    return alert('Session Description must not be empty')
  }

  try {
    pc.setRemoteDescription(new RTCSessionDescription(JSON.parse(atob(sd))))
  } catch (e) {
    alert(e)
  }
}

</script>
</html>
`
