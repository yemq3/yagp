package main

import (
	log "github.com/sirupsen/logrus"
)

// Controller 决定每帧是使用跟踪还是检测
// TODO: 在运行时根据性能数据修改参数
type Controller struct {
	ControllerChannel chan Frame
	encoderChannel chan Frame
	trackerChannel chan Frame
}

// NewController creates a new Controller
func NewController(encoderChannel chan Frame, trackerChannel chan Frame) Controller {
	controller := Controller{}

	controller.ControllerChannel = make(chan Frame)
	controller.encoderChannel = encoderChannel
	controller.trackerChannel = trackerChannel

	return controller
}


func (controller *Controller) run() {
	log.Infof("Controller running...")
	counter := 0
	for {
		select {
		case frame := <-controller.ControllerChannel:
			if counter == 0{
				counter = 1
				log.Debugf("Frameid: %v, go to encoder", frame.FrameID)
				go func(frame Frame) {
					controller.encoderChannel <- frame
				}(frame)
			} else{
				counter = 0
				log.Debugf("Frameid: %v, go to tracker", frame.FrameID)
				go func(frame Frame) {
					controller.trackerChannel <- frame
				}(frame)
			}
		}
	}
}
