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
	DoneTime int64 // 完成跟踪时的时间
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
	Boxes              []Box // 检测数据
	ClientToServerTime int64 // 客户端到服务器端的时间
	SendTime           int64 // 服务器端发送的时间
	ProcessTime        int64 // 服务器端处理请求用的时间
	GetTime            int64 // 客户端收到的时间
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
