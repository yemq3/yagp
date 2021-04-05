package main

import (
	"github.com/yemq3/yagp/box"
	"gocv.io/x/gocv"
)

// Frame ...
type Frame struct {
	FrameID   int
	Frame     gocv.Mat
	Timestamp int64
	isDetect  bool
}

// ResultWithRelativeBox 检测/跟踪结果(框为相对值)
type ResultWithRelativeBox struct {
	FrameID  int
	Boxes    []box.RelativeBox
	DoneTime int64 // 完成检测/跟踪的时间
	Method   int
}

// ResultWithAbsoluteBox 检测/跟踪结果(框为绝对值)
type ResultWithAbsoluteBox struct {
	FrameID  int
	Boxes    []box.AbsoluteBox
	DoneTime int64 // 完成检测/跟踪的时间
	Method   int
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
	Boxes              []box.RelativeBox // 检测数据
	ClientToServerTime int64             // 客户端到服务器端的时间
	SendTime           int64             // 服务器端发送的时间
	ProcessTime        int64             // 服务器端处理请求用的时间
	GetTime            int64             // 客户端收到的时间
}

type currentFrame struct {
	frame         gocv.Mat
	frameID       int
	Boxes         []box.AbsoluteBox
	method        int
	resultFrameID int
}

// Method的状态
const (
	NONE = iota
	DETECT
	TRACK
	DROP
)
