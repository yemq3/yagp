package main

import (
	"flag"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

var addr = flag.String("addr", "127.0.0.1:12345", "Websocket Server Address")
var trackingMethod = flag.String("t", "MedianFlow", "Tracking Method")
var frameRate = flag.Int("frameRate", 16, "FrameRate")
var encodeQuality = flag.Int("encodeQuality", 75, "Encode Quality")
var interval = flag.Int("interval", 1, "Interval")
var resultDir = flag.String("resultDir", "./result", "Path to result")

var (
	WIDTH = 0
	HEIGHT = 0
)

func runCore(messageCenter MessageCenter) {
	// 持久层
	persister := NewPersister(messageCenter, *resultDir)
	go persister.run()
	// 持久层

	// websocket连接
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Infof("connecting to %s", u.String())

	network, err := NewNetwork(u, messageCenter)
	if err != nil {
		return
	}
	// websocket连接

	// 初始化Encoder
	encoder := NewEncoder(*encodeQuality, network.NetworkChannel, messageCenter)
	// 初始化Encoder

	// 初始化Tracker

	tracker, err := NewTracker(messageCenter, *trackingMethod)
	if err != nil {
		log.Errorln(err)
		return
	}
	// 初始化Tracker

	// 初始化Processer
	controler := NewController(messageCenter, encoder.EncoderChannel, tracker.TrackerChannel, tracker.UpdateChannel, *interval)
	// 初始化Processer

	// 初始化Filter
	filter := NewFilter(controler.ControllerChannel, messageCenter)
	//filter.SetFilterFunc()
	// 初始化Filter

	// 初始化摄像头
	camera, err := NewCamera(*frameRate, filter.FilterChannel, true, "../benchmark/video.mp4")
	if err != nil {
		return
	}
	// 初始化摄像头

	// 必须sleep一下再run，要不然ui可能那边来不及Subscribe
	time.Sleep(1 * time.Second)
	go camera.run()
	go filter.run()
	go controler.run()
	go encoder.run()
	go network.run()
	go tracker.run()

}

func main() {
	//log.SetLevel(log.DebugLevel)
	log.SetLevel(log.InfoLevel)
	flag.Parse()
	// 绑定主线程，要不然mac会报错
	runtime.LockOSThread()

	// runtime.GOMAXPROCS(20)

	// 性能分析用
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	// MessageCenter用来发布Frame, Detection Result, Track Result，Qos等信息
	// ensureOrder 目前必须设置为true，因为代码很多部分都假设结果是顺序的= =，设成false能跑但是可能会出现问题
	messageCenter := NewMessageCenter(true)
	go messageCenter.run()
	// MessageCenter用来发布Frame, Detection Result, Track Result，Qos等信息

	go runCore(messageCenter)

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

	detectChannel := messageCenter.Subscribe(DetectResult)
	defer messageCenter.Unsubscribe(detectChannel)

	trackingChannel := messageCenter.Subscribe(TrackResult)
	defer messageCenter.Unsubscribe(trackingChannel)

	var currentFrame currentFrame

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
		case msg := <-detectChannel:
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
			response, ok := msg.Content.(ResultWithAbsoluteBox)
			if !ok {
				log.Errorf("get wrong msg")
				return
			}
			currentFrame.Boxes = response.Boxes
			currentFrame.resultFrameID = response.FrameID
			currentFrame.method = DETECT
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
			trackResult, ok := msg.Content.(ResultWithAbsoluteBox)
			if !ok {
				log.Errorf("wrong msg")
				return
			}
			currentFrame.Boxes = trackResult.Boxes
			currentFrame.resultFrameID = trackResult.FrameID
			currentFrame.method = TRACK
			display(currentFrame, window)
		}
	}

}
