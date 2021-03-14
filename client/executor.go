package main

import (
	"gocv.io/x/gocv"
)

// Executor 和其他组件交互的模块
type Executor struct {
	ControllerChannel chan Frame
	encoderChannel    chan EncodeTask
	trackerChannel    chan TrackTask
}

func (executor *Executor) sendEncodeTask(frame Frame, encodeQuality int, resizeFactor float64) {
	task := EncodeTask{
		frame:         frame,
		encodeQuality: encodeQuality,
		resizeFactor:  resizeFactor,
		encodeMethod:  gocv.JPEGFileExt, // 直接写定了，其他格式的效果都不如jpeg
	}
	executor.encoderChannel <- task
}

func (executor *Executor) sendTrackTask(frame Frame) {
	task := TrackTask{
		frame:  frame,
		isDrop: false,
	}
	executor.trackerChannel <- task
}

func (executor *Executor) DropTask(frame Frame) {
	task := TrackTask{
		frame:  frame,
		isDrop: true,
	}
	executor.trackerChannel <- task
}

// NewExecutor creates a new Executor
func NewExecutor(encoderChannel chan EncodeTask, trackerChannel chan TrackTask) Executor {
	executor := Executor{}

	executor.ControllerChannel = make(chan Frame)
	executor.encoderChannel = encoderChannel
	executor.trackerChannel = trackerChannel

	return executor
}
