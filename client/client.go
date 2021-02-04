package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
)

var addr = flag.String("addr", "127.0.0.1:12345", "websocket server address")
var trackingMethod = flag.String("t", "KCF", "Tracking Method")

const (
	EncodeQuality int     = 75  // jpeg压缩质量
	FrameWidth    float64 = 800 // 摄像头参数
	FrameHeight   float64 = 600 // 摄像头参数
	FrameRate     float64 = 24  // 摄像头参数
)

func main() {
	//log.SetLevel(log.DebugLevel)
	log.SetLevel(log.InfoLevel)
	flag.Parse()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	//runtime.GOMAXPROCS(20)

	// MessageCenter用来发布Frame, Detection Result, Track Result，Qos等信息
	messageCenter := MessageCenter{}
	messageCenter.init()
	go messageCenter.run()
	// MessageCenter用来发布Frame, Detection Result, Track Result，Qos等信息

	// Evaluator
	evaluator := Evaluator{}
	evaluator.init(messageCenter)
	go evaluator.run()
	// Evaluator

	// 显示层
	display := Displayer{}
	if err := display.init(messageCenter); err != nil {
		return
	}
	go display.display()
	go display.displayResult()
	// 显示层

	// 持久层
	persister := Persister{}
	if err := persister.init(messageCenter); err != nil {
		return
	}
	go persister.run()
	// 持久层

	// websocket连接
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Infof("connecting to %s", u.String())

	network := Network{}
	if err := network.init(u, messageCenter); err != nil {
		return
	}
	go network.run()
	// websocket连接

	// 初始化Encoder
	encoder := Encoder{}
	if err := encoder.init(EncodeQuality, network.NetworkChannel, messageCenter); err != nil {
		return
	}
	go encoder.run()
	// 初始化Encoder

	// 初始化Tracker
	tracker := Tracker{}
	if err := tracker.init(messageCenter, *trackingMethod); err != nil{
		return
	}
	go tracker.run()
	// 初始化Tracker

	// 初始化Processer
	controler := Controller{}
	if err := controler.init(encoder.EncoderChannel, tracker.TrackerChannel); err != nil {
		return
	}
	go controler.run()
	// 初始化Processer

	// 初始化Filter
	filter := Filter{}
	if err := filter.init(controler.ControllerChannel, messageCenter); err != nil {
		return
	}
	//filter.SetFilterFunc()
	go filter.run()
	// 初始化Filter

	// 初始化摄像头
	camera := Camera{}
	if err := camera.init(FrameWidth, FrameHeight, FrameRate, filter.FilterChannel); err != nil {
		return
	}
	go camera.run()
	// 初始化摄像头

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	for {
		select {
		case <-interrupt:
			log.Infoln("interrupt")

			os.Exit(1)
			return
		}
	}
}
