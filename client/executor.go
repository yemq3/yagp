package main

import (
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

// Executor 和其他组件交互的模块
type Executor struct {
	ControllerChannel chan Frame
	encoderChannel    chan EncodeTask
	trackerChannel    chan TrackTask
	updateChannel     chan UpdateTask
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
		Frame:  frame,
		IsDrop: false,
	}
	executor.trackerChannel <- task
	log.Debugf("send track task %v", task)
}

func (executor *Executor) sendDropTask(frame Frame) {
	task := TrackTask{
		Frame:  frame,
		IsDrop: true,
	}
	executor.trackerChannel <- task
	log.Debugf("send drop task %v", task)
}

func (executor *Executor) sendUpdateTask(response ResultWithAbsoluteBox) {
	task := UpdateTask{
		FrameID:  response.FrameID,
		Boxes:    response.Boxes,
		Interval: 2,
	}
	executor.updateChannel <- task
	log.Debugf("send update task %v", task)
}

// NewExecutor creates a new Executor
func NewExecutor(encoderChannel chan EncodeTask, trackerChannel chan TrackTask, updateChannel chan UpdateTask) Executor {
	executor := Executor{}

	executor.ControllerChannel = make(chan Frame)
	executor.encoderChannel = encoderChannel
	executor.trackerChannel = trackerChannel
	executor.updateChannel = updateChannel

	return executor
}
