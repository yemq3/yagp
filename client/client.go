package main

import (
	"flag"
	"net/url"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
)

var addr = flag.String("addr", "127.0.0.1:8000", "http service address")

const (
	EncodeQuality int     = 75  // jpeg压缩质量
	FrameWidth    float64 = 800 // 摄像头参数
	FrameHeight   float64 = 600 // 摄像头参数
	FrameRate     float64 = 24  // 摄像头参数
)

func main() {
	log.SetLevel(log.InfoLevel)
	flag.Parse()

	// 显示层
	display := Displayer{}
	if err := display.init(); err != nil {
		return
	}
	go display.display()
	go display.displayDetectionRes()
	// 显示层

	// 持久层
	persister := Persister{}
	if err := persister.init(display.NewFrameNotify, display.NewResponseNotify); err != nil {
		return
	}
	// 持久层

	// websocket连接
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Infof("connecting to %s", u.String())

	network := Network{}
	if err := network.init(u, persister); err != nil {
		return
	}
	go network.run()
	// websocket连接

	// 初始化Encoder
	encoder := Encoder{}
	if err := encoder.init(EncodeQuality, network.SendChannel); err != nil {
		return
	}
	go encoder.run()
	// 初始化Encoder

	// 初始化摄像头
	camera := Camera{}
	if err := camera.init(FrameWidth, FrameHeight, FrameRate, encoder.EncodeChannel, persister); err != nil {
		return
	}
	camera.run(defaultFilterFunc)
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
