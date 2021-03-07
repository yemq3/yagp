package main

import (
	"flag"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

var addr = flag.String("addr", "127.0.0.1:12345", "Websocket Server Address")
var trackingMethod = flag.String("t", "KCF", "Tracking Method")
var frameRate = flag.Int("frameRate", 24, "FrameRate")
var encodeQuality = flag.Int("encodeQuality", 75, "Encode Quality")

func runCore(messageCenter MessageCenter) {
	// Evaluator
	evaluator := NewEvaluator(messageCenter)
	go evaluator.run()
	// Evaluator

	// 持久层
	persister := NewPersister(messageCenter)
	go persister.run()
	// 持久层

	// websocket连接
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Infof("connecting to %s", u.String())

	network, err := NewNetwork(u, messageCenter)
	if err != nil {
		return
	}
	go network.run()
	// websocket连接

	// 初始化Encoder
	encoder := NewEncoder(*encodeQuality, network.NetworkChannel, messageCenter)
	go encoder.run()
	// 初始化Encoder

	// 初始化Tracker

	tracker, err := NewTracker(messageCenter, *trackingMethod)
	if err != nil {
		log.Errorln(err)
		return
	}
	go tracker.run()
	// 初始化Tracker

	// 初始化Processer
	controler := NewController(encoder.EncoderChannel, tracker.TrackerChannel)
	go controler.run()
	// 初始化Processer

	// 初始化Filter
	filter := NewFilter(controler.ControllerChannel, messageCenter)
	//filter.SetFilterFunc()
	go filter.run()
	// 初始化Filter

	// 初始化摄像头
	camera, err := NewCamera(*frameRate, filter.FilterChannel, true, "../benchmark/tracker/video.mp4")
	if err != nil {
		return
	}
	go camera.run()
	// 初始化摄像头
}

func main() {
	//log.SetLevel(log.DebugLevel)
	log.SetLevel(log.InfoLevel)
	flag.Parse()

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	//runtime.GOMAXPROCS(20)

	// MessageCenter用来发布Frame, Detection Result, Track Result，Qos等信息
	messageCenter := NewMessageCenter(true)
	go messageCenter.run()
	// MessageCenter用来发布Frame, Detection Result, Track Result，Qos等信息

	runCore(messageCenter)

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		for {
			select {
			case <-interrupt:
				log.Infoln("interrupt")

				os.Exit(0)
				return
			}
		}
	}()

	window := gocv.NewWindow("Oringin")
	defer window.Close()

	frameChannel := messageCenter.Subscribe(FilterFrame)
	defer messageCenter.Unsubscribe(frameChannel)

	responseChannel := messageCenter.Subscribe(NetworkResponse)
	defer messageCenter.Unsubscribe(responseChannel)

	trackingChannel := messageCenter.Subscribe(TrackerTrackResult)
	defer messageCenter.Unsubscribe(trackingChannel)

	var currentFrameID int
	var currentFrame gocv.Mat

	for {
		select {
		case msg := <-frameChannel:
			frame, ok := msg.Content.(Frame)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			currentFrameID = frame.FrameID
			currentFrame = frame.Frame
			window.IMShow(frame.Frame)
			window.WaitKey(1)
		case msg := <-responseChannel:
			response, ok := msg.Content.(Response)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			if response.FrameID < currentFrameID{

			}
			window.IMShow(currentFrame)
		case msg := <- trackingChannel:
			trackResult, ok := msg.Content.(TrackResult)
			if !ok {
				log.Errorf("wrong msg")
				return
			}
			if trackResult.FrameID < currentFrameID{

			}
			window.IMShow(currentFrame)
		}
	}
}
