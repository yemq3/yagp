package main

import (
	log "github.com/sirupsen/logrus"
	"math/rand"
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
	for {
		select {
		case frame := <-controller.ControllerChannel:
			if rand.Intn(100) < 90{
				log.Debugf("Frameid: %v, go to encoder", frame.FrameID)
				go func(frame Frame) {
					controller.encoderChannel <- frame
				}(frame)
			} else {
				log.Debugf("Frameid: %v, go to tracker", frame.FrameID)
				go func(frame Frame) {
					controller.trackerChannel <- frame
				}(frame)
			}
		}
	}
}
