package main

import (
	log "github.com/sirupsen/logrus"
)

type Controller struct {
	ControllerChannel chan Frame
	encoderChannel chan Frame
	trackerChannel chan Frame
}

func (controller *Controller) init(encoderChannel chan Frame, trackerChannel chan Frame) error {
	controller.ControllerChannel = make(chan Frame)
	controller.encoderChannel = encoderChannel
	controller.trackerChannel = trackerChannel

	return nil
}

func (controller *Controller) run() {
	counter := 0
	for {
		select {
		case frame := <-controller.ControllerChannel:
			if counter == 10{
				counter = 0
				log.Debugf("Frameid: %v, go to encoder", frame.FrameID)
				go func(frame Frame) {
					controller.encoderChannel <- frame
				}(frame)
			} else{
				counter++
				log.Debugf("Frameid: %v, go to tracker", frame.FrameID)
				go func(frame Frame) {
					controller.trackerChannel <- frame
				}(frame)
			}
		}
	}
}
