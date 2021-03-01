package main

import (
	"gocv.io/x/gocv"
	"image"
)

const (
	EncodeQuality int     = 75  // jpeg压缩质量
	FrameWidth    float64 = 800 // 摄像头参数
	FrameHeight   float64 = 600 // 摄像头参数
	FrameRate     float64 = 24  // 摄像头参数
)

// Frame ...
type Frame struct {
	FrameID   int
	Frame     gocv.Mat
	Timestamp int64
	isDetect  bool
}

type TrackResult struct {
	FrameID  int
	Boxes    []image.Rectangle
	DoneTime int64
}

// Request ...
type Request struct {
	FrameID  int
	Frame    []byte
	SendTime int64
}

// Response ...
type Response struct {
	FrameID            int
	Boxes              []Box
	ClientToServerTime int64
	SendTime           int64
	ProcessTime        int64
}

// Box ...
type Box struct {
	X1      float64
	Y1      float64
	X2      float64
	Y2      float64
	Conf    float64
	Name    string
}
