package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
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
	// persister := NewPersister(messageCenter)
	// go persister.run()
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

func display(frame currentFrame, window *gocv.Window) {
	red := color.RGBA{255, 0, 0, 0}
	// blue := color.RGBA{0, 0, 255, 0}
	// log.Infof("%v", frame.frameID)

	if frame.method == None {
		window.IMShow(frame.frame)
		window.WaitKey(1)
		return
	}

	copyFrame := gocv.NewMat()
	frame.frame.CopyTo(&copyFrame)
	height := copyFrame.Size()[0]
	weight := copyFrame.Size()[1]

	for _, box := range frame.Boxes {
		x0 := max(int(box.X1*float64(weight)), 0)
		y0 := max(int(box.Y1*float64(height)), 0)
		x1 := min(int(box.X2*float64(weight)), weight)
		y1 := min(int(box.Y2*float64(height)), height)
		rect := image.Rect(x0, y0, x1, y1)
		gocv.Rectangle(&copyFrame, rect, red, 2)
	}

	if frame.resultFrameID <= frame.frameID {
		text := fmt.Sprintf("delay: %v", frame.frameID-frame.resultFrameID)
		position := image.Point{X: 50, Y: 80}
		gocv.PutText(&copyFrame, text, position, gocv.FontHersheyPlain, 8, red, 8)
	}

	window.IMShow(copyFrame)
	window.WaitKey(1)
}

func main() {
	//log.SetLevel(log.DebugLevel)
	log.SetLevel(log.InfoLevel)
	flag.Parse()

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	// runtime.GOMAXPROCS(1)

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

	var currentFrame currentFrame
	// var isOutofDate bool // 每次拿到新的一帧时设为

	for {
		select {
		case msg := <-frameChannel:
			frame, ok := msg.Content.(Frame)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			currentFrame.frame = frame.Frame
			currentFrame.frameID = frame.FrameID
			display(currentFrame, window)
		case msg := <-responseChannel:
		priority:
			for {
				select {
				case msg := <-frameChannel:
					frame, ok := msg.Content.(Frame)
					if !ok {
						log.Errorf("get wrong msg")
						return
					}
					currentFrame.frame = frame.Frame
					currentFrame.frameID = frame.FrameID
					display(currentFrame, window)
				default:
					break priority
				}
			}
			response, ok := msg.Content.(Response)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			currentFrame.Boxes = response.Boxes
			currentFrame.resultFrameID = response.FrameID
			currentFrame.method = Detect
			display(currentFrame, window)
		case msg := <-trackingChannel:
		priority2:
			for {
				select {
				case msg := <-frameChannel:
					frame, ok := msg.Content.(Frame)
					if !ok {
						log.Errorf("get wrong msg")
						return
					}
					currentFrame.frame = frame.Frame
					currentFrame.frameID = frame.FrameID
					display(currentFrame, window)
				default:
					break priority2
				}
			}
			trackResult, ok := msg.Content.(TrackResult)
			if !ok {
				log.Errorf("wrong msg")
				return
			}
			currentFrame.Boxes = trackResult.Boxes
			currentFrame.resultFrameID = trackResult.FrameID
			currentFrame.method = Track
			display(currentFrame, window)
		}
	}
}
