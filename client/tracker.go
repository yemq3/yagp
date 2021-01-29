package main

import log "github.com/sirupsen/logrus"

type TrackResult struct {
	FrameID int
}

type Tracker struct {
	TrackerChannel chan Frame
	messageCenter MessageCenter
}

func (tracker *Tracker) init(messageCenter MessageCenter) error {
	tracker.TrackerChannel = make(chan Frame)
	tracker.messageCenter = messageCenter

	return nil
}

func (tracker Tracker) run()  {
	for{
		select {
		case frame := <- tracker.TrackerChannel:
			log.Debugf("Tracker get Frame %v", frame.FrameID)
			trackResult := TrackResult{}
			trackResult.FrameID = frame.FrameID

			go func(trackResult TrackResult) {
				msg := Message{
					Topic:   "Track",
					Content: trackResult,
				}
				tracker.messageCenter.Publish(msg)
			}(trackResult)
		}
	}
}
