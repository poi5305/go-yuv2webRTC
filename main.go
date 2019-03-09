package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/poi5305/go-yuv2webRTC/screenshot"
	"github.com/poi5305/go-yuv2webRTC/webrtc"
)

var screenWidth int
var screenHeight int
var resizeWidth int
var resizeHeight int
var webRTC *webrtc.WebRTC

func init() {
	screenWidth, screenHeight = screenshot.GetScreenSize()
	resizeWidth, resizeHeight = screenWidth/2, screenHeight/2
	webRTC = webrtc.NewWebRTC()
	// start screenshot loop, wait for connection
	go screenshotLoop()
}

func main() {
	fmt.Println("http://localhost:8000")

	router := mux.NewRouter()
	router.HandleFunc("/", getWeb).Methods("GET")
	router.HandleFunc("/session", postSession).Methods("POST")

	http.ListenAndServe(":8000", router)
}

func getWeb(w http.ResponseWriter, r *http.Request) {
	bs, err := ioutil.ReadFile("./index.html")
	if err != nil {
		log.Fatal(err)
	}
	w.Write(bs)
}

func postSession(w http.ResponseWriter, r *http.Request) {
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	r.Body.Close()

	localSession, err := webRTC.StartClient(string(bs), resizeWidth, resizeHeight)
	if err != nil {
		log.Fatalln(err)
	}

	w.Write([]byte(localSession))
}

func screenshotLoop() {
	for {
		if webRTC.IsConnected() {
			rgbaImg := screenshot.GetScreenshot(0, 0, screenWidth, screenHeight, resizeWidth, resizeHeight)
			yuv := screenshot.RgbaToYuv(rgbaImg)
			webRTC.ImageChannel <- yuv
		}
		time.Sleep(10 * time.Millisecond)
	}
}
