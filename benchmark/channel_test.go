package main

import (
	"gocv.io/x/gocv"
	"testing"
)

type Frame struct {
	Frame gocv.Mat
}

func BenchmarkChannelSendStruct(b *testing.B) {
	c, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		return
	}
	img := gocv.NewMat()
	defer img.Close()
	c.Read(&img)

	frameChannel := make(chan Frame, 1000)
	go func() {
		for {
			//<-frameChannel
			f := <-frameChannel
			f.Frame.Size()
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		frame := Frame{Frame: img}
		frameChannel <- frame
	}
}

func BenchmarkChannelSendPointer(b *testing.B) {
	c, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		return
	}
	img := gocv.NewMat()
	defer img.Close()

	c.Read(&img)
	//b.Logf("%v", img.Size())

	frameChannel := make(chan *Frame, 1000)
	go func() {
		for {
			//<-frameChannel
			f := <-frameChannel
			f.Frame.Size()
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		frame := &Frame{Frame: img}
		frameChannel <- frame
	}
}
