package main

import (
	"gocv.io/x/gocv"
)

// Frame ...
type Frame struct {
	FrameID   int
	Frame     gocv.Mat
	Timestamp int64
	isDetect  bool
}

// TrackResult 跟踪结果
type TrackResult struct {
	FrameID  int
	Boxes    []Box
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

// Box 框的定义，其中坐标都是0到1的小数点
type Box struct {
	X1   float64
	Y1   float64
	X2   float64
	Y2   float64
	Conf float64
	Name string
}

type currentFrame struct {
	frame         gocv.Mat
	frameID       int
	Boxes         []Box
	method        int
	resultFrameID int
}

// Method的状态
const (
	NONE = iota
	DETECT
	TRACK
	BOTH
)
